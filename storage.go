package caddystoragevalkey

import (
	"context"
	"fmt"
	"io/fs"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/caddyserver/certmagic"
	"github.com/valkey-io/valkey-go"
	"github.com/valkey-io/valkey-go/valkeylock"
)

const (
	LOCKER_PREFIX = "caddylock"

	ENTRY_KEY_VALUE        = "value"
	ENTRY_KEY_LASTMODIFIED = "last_modified"
	ENTRY_KEY_SIZE         = "size"

	TIMEFORMAT = time.RFC3339

	// The default scan count is only ten. However, when having a lot of certificates
	// in the storage, iterating or listing the files can take quite a long time.
	// Increasing this allows for faster iteration.
	SCAN_COUNT = 50
)

type CaddyStorageValkey struct {
	client valkey.Client
	locker valkeylock.Locker
	locks  sync.Map
}

type CaddyStorageValkeyOptions struct {
	LockMajority int
}

func NewCaddyStorageValkey(clientOptions valkey.ClientOption, options CaddyStorageValkeyOptions) (*CaddyStorageValkey, error) {
	// Create a new client for valkey
	valkeyClient, err := valkey.NewClient(clientOptions)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	// Create a new locker for valkey
	valkeyLocker, err := valkeylock.NewLocker(valkeylock.LockerOption{
		ClientOption:   clientOptions,
		KeyPrefix:      LOCKER_PREFIX,
		NoLoopTracking: true,
		KeyMajority:    int32(options.LockMajority),
	})

	if err != nil {
		// Cleanup unused client
		valkeyClient.Close()
		return nil, err
	}

	return &CaddyStorageValkey{client: valkeyClient, locker: valkeyLocker}, nil
}

func (c *CaddyStorageValkey) Lock(ctx context.Context, key string) error {
	// Abort when the cancel function for the lock already exists
	if c.Exists(ctx, key) {
		return fmt.Errorf("lock already exists locally for key '%s'", key)
	}

	// Acquire the lock for the given key
	_, cancel, err := c.locker.TryWithContext(ctx, key)

	if err != nil {
		return err
	}

	// Remmeber the cancel function in order to unlock the lock
	c.locks.Store(key, cancel)

	return nil
}

func (c *CaddyStorageValkey) Unlock(ctx context.Context, key string) error {
	// When lock found, unlock it
	if unlock, ok := c.locks.Load(key); ok {
		// Unlock and delete it
		unlock.(context.CancelFunc)()
		c.locks.Delete(key)

		return nil
	}

	return fmt.Errorf("lock does not exists locally for key '%s'", key)
}

func (c *CaddyStorageValkey) Store(ctx context.Context, key string, value []byte) error {
	// Store the key and value with metdata
	return c.client.Do(
		ctx,
		c.client.B().Hmset().
			Key(key).
			FieldValue().
			FieldValue(ENTRY_KEY_VALUE, string(value[:])).
			FieldValue(ENTRY_KEY_LASTMODIFIED, time.Now().Format(TIMEFORMAT)).
			FieldValue(ENTRY_KEY_SIZE, fmt.Sprint(len(value))).Build()).Error()
}

func (c *CaddyStorageValkey) Load(ctx context.Context, key string) ([]byte, error) {
	// Get only the value from valkey without meta info
	value, err := c.client.Do(
		ctx,
		c.client.B().Hget().
			Key(key).
			Field(ENTRY_KEY_VALUE).Build()).AsBytes()

	// Caddy expects a specific fs Error for when the key is not present
	if value == nil || valkey.IsValkeyNil(err) {
		return nil, fs.ErrNotExist
	}

	return value, err
}

func (c *CaddyStorageValkey) Delete(ctx context.Context, key string) error {
	return c.client.Do(ctx, c.client.B().Del().Key(key).Build()).Error()
}

func (c *CaddyStorageValkey) Exists(ctx context.Context, key string) bool {
	r, err := c.client.Do(ctx, c.client.B().Exists().Key(key).Build()).AsBool()
	if err != nil {
		return false
	}

	return r
}

func (c *CaddyStorageValkey) List(ctx context.Context, prefix string, recursive bool) ([]string, error) {
	r := []string{}
	initialCursorId := uint64(0)
	cursorId := initialCursorId

	// For handling non-recursive list
	keysMap := make(map[string]bool)

	for {
		// Scan based on the given prefix
		entry, err := c.client.Do(
			ctx,
			c.client.B().Scan().
				Cursor(cursorId).
				Match(fmt.Sprintf("%s*", prefix)).
				Count(SCAN_COUNT).Build()).AsScanEntry()
		if err != nil {
			return nil, err
		}

		if !recursive {
			// for non-recursive split path and look for unique keys just under given prefix
			for _, key := range entry.Elements {
				dir := strings.Split(strings.TrimPrefix(key, prefix+"/"), "/")
				keysMap[dir[0]] = true
			}
		} else {
			// for recursive, we accept all elements
			r = append(r, entry.Elements...)
		}

		// Scan is done, when we arrived at the initial cursor again
		if entry.Cursor == initialCursorId {
			break
		}

		// Move to next cursor
		cursorId = entry.Cursor
	}

	if !recursive {
		// for non-recursive extract actual keys from keys map
		for key := range keysMap {
			r = append(r, path.Join(prefix, key))
		}
	}

	return r, nil
}

func (c *CaddyStorageValkey) Stat(ctx context.Context, key string) (certmagic.KeyInfo, error) {
	// Minimal keyinfo (IsTerminal is always true, as we only create files, no directories)
	info := certmagic.KeyInfo{Key: key, IsTerminal: true}

	// Get meta info for key
	value, err := c.client.Do(
		ctx,
		c.client.B().Hmget().
			Key(key).
			Field(ENTRY_KEY_LASTMODIFIED, ENTRY_KEY_SIZE).Build()).ToArray()

	if err != nil {
		return info, err
	}

	if len(value) != 2 {
		return info, fmt.Errorf("unexpected return length of reading values for key '%s'", key)
	}

	// Parse last modified
	lastModifiedRaw, err := value[0].ToString()
	if err != nil {
		return info, err
	}

	lastModified, err := time.Parse(TIMEFORMAT, lastModifiedRaw)
	if err != nil {
		return info, err
	}
	info.Modified = lastModified

	// Parse size
	sizeRaw, err := value[1].ToString()
	if err != nil {
		return info, err
	}

	size, err := strconv.ParseInt(sizeRaw, 10, 64)
	if err != nil {
		return info, err
	}
	info.Size = size

	return info, nil
}

func (c *CaddyStorageValkey) Close() error {
	// Cleanup all held locks by this instance
	c.locks.Range(func(key, value any) bool {
		value.(context.CancelFunc)()

		return true
	})
	c.locks.Clear()

	// Close all connections
	c.client.Close()
	c.locker.Close()

	return nil
}

var (
	_ certmagic.Storage = (*CaddyStorageValkey)(nil)
)

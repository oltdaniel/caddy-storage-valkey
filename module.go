package caddystoragevalkey

import (
	"errors"
	"strconv"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/certmagic"
	"github.com/valkey-io/valkey-go"
)

const (
	ID_MODULE_STATE = "caddy.storage.valkey"
)

type StorageValkeyModule struct {
	Url string `json:"url,omitempty"`

	InitAddress    []string `json:"address,omitempty"`
	ReplicaAddress []string `json:"replica,omitempty"`
	SelectDb       int      `json:"db,omitempty"`

	ShuffleInit       bool   `json:"shuffle_init,omitempty"`
	SentinelMasterSet string `json:"sentinel_master_set,omitempty"`

	LockMajority       int    `json:"lock_majority,omitempty"`
	DisableClientCache bool   `json:"disable_client_cache,omitempty"`
	SendToReplicas     string `json:"send_to_replicas,omitempty"`

	storage *CaddyStorageValkey
}

func init() {
	caddy.RegisterModule(StorageValkeyModule{})
}

func (StorageValkeyModule) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  ID_MODULE_STATE,
		New: func() caddy.Module { return new(StorageValkeyModule) },
	}
}

func (m *StorageValkeyModule) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			configKey := d.Val()
			var configVal []string

			if d.NextArg() {
				// configuration item with single parameter
				configVal = append(configVal, d.Val())
			} else {
				// configuration item with nested parameter list
				for nesting := d.Nesting(); d.NextBlock(nesting); {
					configVal = append(configVal, d.Val())
				}
			}
			// There are no valid configurations where configVal slice is empty
			if len(configVal) == 0 {
				return d.Errf("no value supplied for configuraton key '%s'", configKey)
			}

			switch configKey {
			case "url":
				{
					if len(configVal) > 1 {
						return d.Err("expected only a single value for `url`")
					}

					m.Url = configVal[0]
				}
			case "address":
				m.InitAddress = configVal
			case "replica":
				m.ReplicaAddress = configVal
			case "db":
				{
					selectDb, err := parseConfigValToInt(configVal)
					if err != nil {
						return d.WrapErr(err)
					}
					m.SelectDb = int(selectDb)
				}
			case "shuffle_init":
				{
					shuffleInit, err := parseConfigValToBool(configVal)
					if err != nil {
						return d.WrapErr(err)
					}

					m.ShuffleInit = shuffleInit
				}
			case "sentinel_master_set":
				{
					if len(configVal) > 1 {
						return d.Err("expected only a single value for `sentinel_master_set`")
					}

					m.SentinelMasterSet = configVal[0]
				}
			case "lock_majority":
				{
					lockMajority, err := parseConfigValToInt(configVal)
					if err != nil {
						return d.WrapErr(err)
					}

					if lockMajority < 1 {
						return d.Err("impossible value for `lock_majority` (value > 0 required)")
					}

					m.LockMajority = int(lockMajority)
				}
			case "disable_client_cache":
				{
					disableClientCache, err := parseConfigValToBool(configVal)
					if err != nil {
						return d.WrapErr(err)
					}

					m.DisableClientCache = disableClientCache
				}
			case "send_to_replicas":
				{
					if len(configVal) > 1 {
						return d.Err("expected only a single value for `send_to_replica`")
					}

					m.SendToReplicas = configVal[0]
				}
			default:
				// Unknown key for this config
				d.ArgErr()
			}
		}
	}

	return nil
}

func parseConfigValToInt(configVal []string) (int, error) {
	if len(configVal) != 1 {
		return 0, errors.New("can only accept single value as integer")
	}

	// NOTE: Because int is "at-least 32bits", we only parse 32bits
	val, err := strconv.ParseInt(configVal[0], 10, 32)
	if err != nil {
		return 0, err
	}

	return int(val), nil
}

func parseConfigValToBool(configVal []string) (bool, error) {
	if len(configVal) != 1 {
		return false, errors.New("can only accept single value as bool")
	}

	val, err := strconv.ParseBool(configVal[0])
	if err != nil {
		return false, err
	}

	return val, nil
}

func (m *StorageValkeyModule) Validate() error {
	// NOTE: The majority of checking will be done by creating a new valkey client.

	// Verify no other client option is set when the URL has been set
	isUrlSet := (m.Url != "")

	if len(m.InitAddress) > 0 && isUrlSet {
		return errors.New("setting the `address` and `url` option is not allowed")
	}

	// NOTE: I'm aware that setting the db option can be set to zero, but this is the default
	// and doesn't change the behavior in any way. The workaround is not worth it.
	if m.SelectDb != 0 && isUrlSet {
		return errors.New("setting the `db` and `url` option is not allowed")
	}

	// Check for sensible value on lock majority
	if m.LockMajority < 1 {
		return errors.New("impossible value for `lock_majority` option (value > 0 required)")
	}

	// Check SendToReplicas for valid strategy
	switch m.SendToReplicas {
	case "", "none":
		// This is the default option which equals to none/no command to replica
		break
	case "readonly":
		// This option sends readonly commands to the replica
		break
	default:
		return errors.New("invalid value for `send_to_replicas`")
	}

	return nil
}

func (m *StorageValkeyModule) Provision(ctx caddy.Context) error {
	// Create a new client options object based on the parsed config
	var clientOptions *valkey.ClientOption

	if m.Url != "" {
		optionsFromUrl, err := valkey.ParseURL(m.Url)

		if err != nil {
			return err
		}

		clientOptions = &optionsFromUrl
	} else {
		clientOptions = &valkey.ClientOption{
			InitAddress: m.InitAddress,
			SelectDB:    m.SelectDb,
		}
	}

	// Add replica addresses when entries present
	if len(m.ReplicaAddress) > 0 {
		clientOptions.Standalone.ReplicaAddress = m.ReplicaAddress
		clientOptions.SendToReplicas = func(cmd valkey.Completed) bool {
			return false
		}
	}

	// Transfer Disable Client Cache option
	clientOptions.DisableCache = m.DisableClientCache

	// Set SendToReplicas readonly strategy if present
	if m.SendToReplicas == "readonly" {
		clientOptions.SendToReplicas = func(cmd valkey.Completed) bool {
			return cmd.IsReadOnly()
		}
	}

	// Create caddy valkey storage specific options
	options := CaddyStorageValkeyOptions{
		LockMajority: m.LockMajority,
	}

	// Provision a new storage instance
	valkeyStorage, err := NewCaddyStorageValkey(*clientOptions, options)

	if err != nil {
		return err
	}

	m.storage = valkeyStorage

	return nil
}

func (m StorageValkeyModule) Cleanup() error {
	if m.storage != nil {
		m.storage.Close()
	}

	return nil
}

func (m *StorageValkeyModule) CertMagicStorage() (certmagic.Storage, error) {
	return m.storage, nil
}

// Interface guards
var (
	_ caddy.Module           = (*StorageValkeyModule)(nil)
	_ caddy.Provisioner      = (*StorageValkeyModule)(nil)
	_ caddy.CleanerUpper     = (*StorageValkeyModule)(nil)
	_ caddy.Validator        = (*StorageValkeyModule)(nil)
	_ caddyfile.Unmarshaler  = (*StorageValkeyModule)(nil)
	_ caddy.StorageConverter = (*StorageValkeyModule)(nil)
)

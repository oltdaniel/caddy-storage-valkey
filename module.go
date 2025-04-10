package caddystoragevalkey

import (
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
	InitAddress []string
	SelectDb    int

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
			case "address":
				m.InitAddress = configVal
			case "select_db":
				{
					if len(configVal) != 1 {
						d.Err("can only accept single db value")
					}

					selectDb, err := strconv.ParseInt(configVal[0], 10, 64)
					if err != nil {
						return d.WrapErr(err)
					}
					m.SelectDb = int(selectDb)
				}
			default:
				// Unknown key for this config
				d.ArgErr()
			}
		}
	}

	return nil
}

func (m *StorageValkeyModule) Validate() error {
	return nil
}

func (m *StorageValkeyModule) Provision(ctx caddy.Context) error {
	// Provision a new storage instance
	valkeyStorage, err := NewCaddyStorageValkey(valkey.ClientOption{
		InitAddress: m.InitAddress,
		SelectDB:    m.SelectDb,
	})

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

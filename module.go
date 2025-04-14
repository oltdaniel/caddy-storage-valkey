package caddystoragevalkey

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
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

	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`

	TlsInsecure bool   `json:"tls_insecure,omitempty"`
	TlsCaCert   string `json:"tls_ca_cert,omitempty"`
	TlsCliCert  string `json:"tls_cli_cert,omitempty"`
	TlsCliKey   string `json:"tls_cli_key,omitempty"`

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
			case "username":
				{
					if len(configVal) > 1 {
						return d.Err("expected only a single value for `username`")
					}

					m.Username = configVal[0]
				}
			case "password":
				{
					if len(configVal) > 1 {
						return d.Err("expected only a single value for `password`")
					}

					m.Password = configVal[0]
				}
			case "tls_insecure":
				{
					tlsInsecure, err := parseConfigValToBool(configVal)
					if err != nil {
						return d.WrapErr(err)
					}

					m.TlsInsecure = tlsInsecure
				}
			case "tls_ca_cert":
				{
					if len(configVal) > 1 {
						return d.Err("expected only a single value for `tls_ca_cert`")
					}

					m.TlsCaCert = configVal[0]
				}
			case "tls_cli_cert":
				{
					if len(configVal) > 1 {
						return d.Err("expected only a single value for `tls_cli_cert`")
					}

					m.TlsCliCert = configVal[0]
				}
			case "tls_cli_key":
				{
					if len(configVal) > 1 {
						return d.Err("expected only a single value for `tls_cli_key`")
					}

					m.TlsCliKey = configVal[0]
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

func isFilePath(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func checkPemStringOrFilepath(val string) (bool, error) {
	if cert, _ := pem.Decode([]byte(val)); cert == nil {
		if !isFilePath(val) {
			return false, errors.New("value is no certificate or filepath")
		}
	}

	return true, nil
}

func validatePemStringOrFilepathOption(val string) bool {
	// Not setting the options is valid
	if len(val) == 0 {
		return true
	}

	// Check if it is a PEM string or filepath
	ok, err := checkPemStringOrFilepath(val)

	return ok && err == nil
}

func (m *StorageValkeyModule) Validate() error {
	// NOTE: The majority of checking will be done by creating a new valkey client.

	// Verify no other client option is set when the URL has been set
	isUrlSet := (m.Url != "")

	if len(m.InitAddress) > 0 && isUrlSet {
		return errors.New("setting the `address` and `url` option is not allowed")
	}

	if len(m.Username) > 0 && isUrlSet {
		return errors.New("setting the `username` and `url` option is not allowed")
	}

	if len(m.Password) > 0 && isUrlSet {
		return errors.New("setting the `password` and `url` option is not allowed")
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

	// Verify TLS options are PEM or filepaths
	if !validatePemStringOrFilepathOption(m.TlsCaCert) {
		return errors.New("given value is no PEM string or filepath for key `tls_ca_cert`")
	}
	if !validatePemStringOrFilepathOption(m.TlsCliCert) {
		return errors.New("given value is no PEM string or filepath for key `tls_cli_cert`")
	}
	if !validatePemStringOrFilepathOption(m.TlsCliKey) {
		return errors.New("given value is no PEM string or filepath for key `tls_cli_key`")
	}

	return nil
}

func (m *StorageValkeyModule) Provision(ctx caddy.Context) error {
	// Apply defaults where required

	if m.LockMajority < 1 {
		m.LockMajority = 2
	}

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

	// Check whether any TLS option has been set
	isTlsConfigured := (m.TlsInsecure ||
		len(m.TlsCaCert) > 0 ||
		len(m.TlsCliCert) > 0 ||
		len(m.TlsCliKey) > 0)

	if isTlsConfigured {
		// Initialize client TLS config if not present
		// NOTE: It can be present when we parse the URL and it has the TLS mentioned
		if clientOptions.TLSConfig == nil {
			clientOptions.TLSConfig = &tls.Config{
				InsecureSkipVerify: m.TlsInsecure,
			}
		} else {
			clientOptions.TLSConfig.InsecureSkipVerify = m.TlsInsecure
		}

		// Initialize CA Certificate if present
		if len(m.TlsCaCert) > 0 {
			caCertPool := x509.NewCertPool()
			if certData, _ := pem.Decode([]byte(m.TlsCaCert)); certData != nil {
				if cert, err := x509.ParseCertificate(certData.Bytes); err == nil {
					caCertPool.AddCert(cert)
				} else {
					return errors.New("failed to add `tls_ca_cert` PEM string to certificate pool")
				}
			} else if isFilePath(m.TlsCaCert) {
				cert, err := os.ReadFile(m.TlsCaCert)
				if err != nil {
					return err
				}
				if ok := caCertPool.AppendCertsFromPEM(cert); !ok {
					return errors.New("failed to add `tls_ca_cert` file content to certificate pool")
				}
			} else {
				return errors.New("failed to add `tls_ca_cert` is no PEM string or filepath")
			}
			clientOptions.TLSConfig.RootCAs = caCertPool
		}

		// Configure client certificate
		if len(m.TlsCliCert) > 0 || len(m.TlsCliKey) > 0 {
			// SMall sanity check
			if len(m.TlsCliCert) > 0 && len(m.TlsCliKey) == 0 {
				return errors.New("both client certificate and key need to be provided, key is missing")
			} else if len(m.TlsCliCert) == 0 && len(m.TlsCliKey) > 0 {
				return errors.New("both client certificate and key need to be provided, certificate is missing")
			}

			var certPem []byte
			var keyPem []byte

			// NOTE: The following two blocks have been copied in order to keep the verbose error messages

			if certData, _ := pem.Decode([]byte(m.TlsCliCert)); certData != nil {
				certPem = []byte(m.TlsCliCert)
			} else if isFilePath(m.TlsCliCert) {
				rawCert, err := os.ReadFile(m.TlsCliCert)
				if err != nil {
					return err
				}
				if certData, _ := pem.Decode(rawCert); certData != nil {
					certPem = rawCert
				} else {
					return errors.New("invalid PEM in `tls_cli_cert` file")
				}
			} else {
				return errors.New("failed to add `tls_cli_cert` is no PEM string or filepath")
			}

			if keyData, _ := pem.Decode([]byte(m.TlsCliKey)); keyData != nil {
				keyPem = []byte(m.TlsCliKey)
			} else if isFilePath(m.TlsCliKey) {
				rawKey, err := os.ReadFile(m.TlsCliKey)
				if err != nil {
					return err
				}
				if keyData, _ := pem.Decode(rawKey); keyData != nil {
					keyPem = rawKey
				} else {
					return errors.New("invalid PEM in `tls_cli_key` file")
				}
			} else {
				return errors.New("failed to add `tls_cli_key` is no PEM string or filepath")
			}

			// Build certificate out of keypair
			clientCertificate, err := tls.X509KeyPair(certPem, keyPem)
			if err != nil {
				return err
			}

			// Given we have a client certificate, we require and verify everything for security purposes
			clientOptions.TLSConfig.ClientAuth = tls.RequireAndVerifyClientCert
			clientOptions.TLSConfig.Certificates = []tls.Certificate{clientCertificate}
		}
	}

	// Set username and password connection details
	if len(m.Username) > 0 {
		clientOptions.Username = m.Username
	}
	if len(m.Password) > 0 {
		clientOptions.Password = m.Password
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

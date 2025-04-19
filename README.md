<p align="center">
	<picture>
		<source media="(prefers-color-scheme: dark)" srcset="https://user-images.githubusercontent.com/1128849/210187358-e2c39003-9a5e-4dd5-a783-6deb6483ee72.svg">
		<source media="(prefers-color-scheme: light)" srcset="https://user-images.githubusercontent.com/1128849/210187356-dfb7f1c5-ac2e-43aa-bb23-fc014280ae1f.svg">
		<img src="https://user-images.githubusercontent.com/1128849/210187356-dfb7f1c5-ac2e-43aa-bb23-fc014280ae1f.svg" alt="Caddy" width="200">
	</picture>
    &nbsp;&nbsp;&nbsp;&nbsp;
    <picture>
        <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/valkey-io/assets/refs/heads/main/Valkey%20Branding/logo%20svgs/valkey-horizontal-color-light.svg">
		<source media="(prefers-color-scheme: light)" srcset="https://raw.githubusercontent.com/valkey-io/assets/refs/heads/main/Valkey%20Branding/logo%20svgs/valkey-horizontal-color.svg">
		<img src="https://raw.githubusercontent.com/valkey-io/assets/refs/heads/main/Valkey%20Branding/logo%20svgs/valkey-horizontal-color.svg" alt="Valkey" width="200">
	</picture>
	<br>
	<h3 align="center">Caddy x Valkey</h3>
</p>


> This project does not affiliate with Caddy nor Valkey. It only extends Caddy with custom code to integrate the valkey storage into Caddy. Logo Copyright belongs to the corresponding project.

# Caddy Storage Valkey

Caddy storage backend module with the native valkey golang client and caddy, nothing else.

> This module is still in development. Breaking changes will likely come. No stability checks yet.

## Example

```
{
    storage valkey {
        address 127.0.0.1:6379
    }
}

mydomain.localhost {
    respond "Hello World" 200
}
```

## Why?

Mainly for just building it and integrating the native valkey golang client as a caddy storage module.

> Thanks to @gamalan ([`gamalan/caddy-tlsredis`](https://github.com/gamalan/caddy-tlsredis)) for showing what a possible implementation for this kind of database could look like.

## Installation

Download a caddy binary from `caddyserver.com` with this package included [here](https://caddyserver.com/download?package=github.com%2Foltdaniel%2Fcaddy-storage-valkey).

### CLI Download (experimental)

> This is equal to the version above but replaces your existing binary with the new one including the package.

Caddy has a feature to add packages to your current installation by running the following command:

```bash
caddy add-package github.com/oltdaniel/caddy-storage-valkey
```

### DIY Route

Build a custom binary of the latest caddy release with this module enabled.

```bash
CADDY_VERSION=latest xcaddy build --with github.com/oltdaniel/caddy-storage-valkey
./caddy run
```

## Configuration Interface

### storage module `valkey`

> [!IMPORTANT]  
> All important client options should be exposed to the config. If there are any missing that could have any purposes in this specific use-case, please open an Issue.

#### Examples

```
# Connecting to single valkey node
storage valkey {
    # Server specific connection information can be passed in url format or as seperate config options
    # See https://github.com/redis/redis-specifications/blob/1252427cdbc497f66a7f8550c6b5f2f35367dc92/uri/redis.txt
    url valkey://localhost:6379/0

    lock_majority 1
    disable_client_cache true
}

storage valkey {
    # Server address can be passed as single entry or list instead of url
    address 127.0.0.1:6379

    db 0

    lock_majority 1
    disable_client_cache true
}

# Connecting with specific user to single node
storage valkey {
    url valkey://caddy:pleasechangeme@localhost:6382
}

storage valkey {
    address localhost:6382

    username caddy
    password pleasechangeme
}

# Connecting to TLS single node
storage valkey {
    url valkeys://localhost:6380

    tls_insecure false
    tls_min_version tlsv1.2

    # Any certificate or key can be passed as a PEM string or filepath as described in the table
    tls_ca_cert <<CACERT
    -----BEGIN CERTIFICATE-----
    ...
    -----END CERTIFICATE-----
    CACERT
}

storage valkey {
    url valkeys://localhost:6380

    tls_insecure false
    tls_min_version tlsv1.2

    # Any certificate or key can be passed as a PEM string or filepath as described in the table
    tls_ca_cert tests/ca.crt
}


# Connecting to TLS client auth single node
storage valkey {
    url valkeys://localhost:6381

    tls_insecure false
    tls_min_version tlsv1.2

    tls_ca_cert <<CACERT
    -----BEGIN CERTIFICATE-----
    ...
    -----END CERTIFICATE-----
    CACERT
    tls_client_cert <<CLICERT
    -----BEGIN CERTIFICATE-----
    ...
    -----END CERTIFICATE-----
    CLICERT
    tls_client_cert <<CLIKEY
    -----BEGIN PRIVATE KEY-----
    ...
    -----END PRIVATE KEY-----
    CLIKEY
}

# Connecting to standalone valkey with replicas
storage valkey {
    url valkey://localhost:6379/0

    replica {
        localhost:6376
    }

    lock_majority 1
    disable_client_cache true
    send_to_replicas readonly
}

# Connecting to valkey cluster
storage valkey {
    # See https://github.com/redis/redis-specifications/blob/1252427cdbc497f66a7f8550c6b5f2f35367dc92/uri/redis.txt
    url valkey://localhost:7001?addr=localhost:7002&addr=localhost:7003

    shuffle_init true

    lock_majority 2
    disable_client_cache true
}

# Connecting to valkey sentinels
storage valkey {
    # See https://github.com/redis/redis-specifications/blob/1252427cdbc497f66a7f8550c6b5f2f35367dc92/uri/redis.txt
    url valkey://localhost:7001?addr=localhost:7002&addr=localhost:7003

    sentinel_master_set my_master

    lock_majority 2
    disable_client_cache true
}
```

#### Values

| Name | Values | Description |
|-|-|-|
| `url` | any valkey client compatible uri schema | Any valid URL can be passed as documented by the valkey go client library [`valkey.ParseURL`](https://pkg.go.dev/github.com/valkey-io/valkey-go#ParseURL). This setting will conflict with any other client option set and will cause this module to throw an error due to an invalid config. |
| `address` | single or list of valkey servers | This option accepts a single or a list of valkey server addresses in any format supported by the valkey go client `InitAddress` option. |
| `replica` | single or list of valkey replica read-only servers | This option accepts a single or a list of valkey server addresses in any format supported by the valkey go client `StandaloneOption.ReplicaAddress` option. |
| `db` | valid integer for selecting the valkey database <br><br>Default: `0` | The range of a valid value in this case depends on your server configuration. Typical range is `0-15` (total 16). |
| `shuffle_init` | accepted input for [`strconv.ParseBool`](https://pkg.go.dev/strconv#ParseBool) <br><br>Default: `false` | Indicates to the client to shuffle all available addresses before connecting to the first entry. |
| `sentinel_master_set` | sentinel master set name | This is the name you configured for your master set in you valkey sentinels setup. |
| `lock_majority` | any integer larger than 0 <br><br>Default: `2` | The number of keys the client needs to aqcuire to receive the ownership of the requested lock. For more details take a look at the documentation of the [`valkey-go/valkeylock`](https://github.com/valkey-io/valkey-go/tree/main/valkeylock) package. |
| `disable_client_cache` | accepted input for [`strconv.ParseBool`](https://pkg.go.dev/strconv#ParseBool) <br><br>Default: `false` | Indicates whether to disable client side caching. |
| `send_to_replicas` | `none`, `readonly` <br><br>Default: `none` | Defines the strategy to determine what should be send to the replicas. |
| `username` | username to authenticate against server | Sets the username to use to authenticate against server. This value is ignored, when using URL format for connection. |
| `password` | password to authenticate against server | Sets the password to use to authenticate against server. This value is ignored, when using URL format for connection. |
| `tls_ca_cert` | ca certificate as string or filepath | Sets the CA certificate for the client in order to verify CA certificate upon connection. |
| `tls_insecure` | accepted input for [`strconv.ParseBool`](https://pkg.go.dev/strconv#ParseBool) <br><br>Default: `false` | Can disable/enable the verification of server CA certificate when connecting via TLS. <br><br> **NOTE: Should not be used in production.** |
| `tls_min_version` | `tlsv1.2`, `tlsv1.3` <br><br>Default: `tlsv1.2` | Set the minimum TLS version that the connection needs to use. <br><br> **NOTE: Older versions have been excluded as they are not recommended and the default for Valkey is TLSv1.2 and TLSv1.3.** |
| `tls_client_cert` | client certificate as string or filepath | Sets the certificate for the client to use for TLS authentication. Needs to be combined with `tls_client_key`. |
| `tls_client_key` | client certificate key as string or filepath | Sets the certificate key for the client to use for TLS authentication. Needs to be combined with `tls_client_cert`. |

### More?

Do you have a neat way of using this library in your Caddyfile? Feel free to submit it.

## Internals

We use the most simple commands in order to make this work and avoid managing any extra structures. Each file is stored as a Hash, with `value`, `last_modified` and `size` in order to store the content of the file and its metadata without any additional serialization. Walking through directories is simply done by doing an Scan and processing of the records in order to return valid results. This means, there is no additional command to repair any internal structures, as there are only Valkey native data structures and mostly single commands for a single action.

The Lock structure is handled by the sub-package `valkeylock` of the Valkey Go Client Library and some essential aspects are exposed via the configuration.

### Resources

- [Valkey Go Client Library](https://github.com/valkey-io/valkey-go)
- [Old Caddy storage redis module](https://github.com/gamalan/caddy-tlsredis)

## Development

Clone, create example config and run with xcaddy.

```bash
git clone https://github.com/oltdaniel/caddy-storage-valkey.git
cd caddy-storage-valkey

CADDY_VERSION=master xcaddy run
# or
CADDY_VERSION=master xcaddy build --with github.com/oltdaniel/caddy-storage-valkey=.
./caddy run
```

Additionally, there is a [`docker-compose.yml`](./docker-compose.yml) which contains a demo setup for many different valkey server setups that can be used for testing.

### Testing Performance

In order to test locking and the load the setup can handle, there is a testing script in [`./scripts/generate-benchmark.sh`](./scripts/generate-benchmark.sh) which will generate an Caddyfile with a huger number of domains for which internally signed certificates are generated with a lifetime of 1 hour and a storage cleanup intervall of 60 seconds, to stress this storage module.

> A test with 10.000 domains, showed the cleanup to take about 90 seconds to finish. There haven't been any long running tests yet or other more extreme tests.

### Testing TLS

In order to test the TLS feature locally, there is a small script [`./scripts/generate-test-certs.sh`](./scripts/generate-test-certs.sh) which generates all the necessary certificates for testing. The ports in the examples above match the correct Container which is already configured to be used for tetsing each scenario.

## License

![GitHub License](https://img.shields.io/github/license/oltdaniel/caddy-storage-valkey)
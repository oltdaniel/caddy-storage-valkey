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

```caddyfile
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

```caddyfile
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
    tls_client_key <<CLIKEY
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

# Using caddy placeholders
storage valkey {
    url {env.VALKEY_URI}
}
```

#### Values

| Name | Values | Placeholders | Description |
|-|-|-|-|
| `url` | any valkey client compatible uri schema | yes | Any valid URL can be passed as documented by the valkey go client library [`valkey.ParseURL`](https://pkg.go.dev/github.com/valkey-io/valkey-go#ParseURL). This setting will conflict with any other client option set and will cause this module to throw an error due to an invalid config. |
| `address` | single or list of valkey servers | yes | This option accepts a single or a list of valkey server addresses in any format supported by the valkey go client `InitAddress` option. |
| `replica` | single or list of valkey replica read-only servers | yes | This option accepts a single or a list of valkey server addresses in any format supported by the valkey go client `StandaloneOption.ReplicaAddress` option. |
| `db` | valid integer for selecting the valkey database <br><br>Default: `0` | no | The range of a valid value in this case depends on your server configuration. Typical range is `0-15` (total 16). |
| `shuffle_init` | accepted input for [`strconv.ParseBool`](https://pkg.go.dev/strconv#ParseBool) <br><br>Default: `false` | no | Indicates to the client to shuffle all available addresses before connecting to the first entry. |
| `sentinel_master_set` | sentinel master set name | no | This is the name you configured for your master set in you valkey sentinels setup. |
| `lock_majority` | any integer larger than 0 <br><br>Default: `2` | no | The number of keys the client needs to aqcuire to receive the ownership of the requested lock. For more details take a look at the documentation of the [`valkey-go/valkeylock`](https://github.com/valkey-io/valkey-go/tree/main/valkeylock) package. |
| `disable_client_cache` | accepted input for [`strconv.ParseBool`](https://pkg.go.dev/strconv#ParseBool) <br><br>Default: `false` | no | Indicates whether to disable client side caching. |
| `send_to_replicas` | `none`, `readonly` <br><br>Default: `none` | no | Defines the strategy to determine what should be send to the replicas. |
| `username` | username to authenticate against server | yes | Sets the username to use to authenticate against server. This value is ignored, when using URL format for connection. |
| `password` | password to authenticate against server | yes | Sets the password to use to authenticate against server. This value is ignored, when using URL format for connection. |
| `tls_ca_cert` | ca certificate as string or filepath | yes | Sets the CA certificate for the client in order to verify CA certificate upon connection. |
| `tls_insecure` | accepted input for [`strconv.ParseBool`](https://pkg.go.dev/strconv#ParseBool) <br><br>Default: `false` | no | Can disable/enable the verification of server CA certificate when connecting via TLS. <br><br> **NOTE: Should not be used in production.** |
| `tls_min_version` | `tlsv1.2`, `tlsv1.3` <br><br>Default: `tlsv1.2` | no | Set the minimum TLS version that the connection needs to use. <br><br> **NOTE: Older versions have been excluded as they are not recommended and the default for Valkey is TLSv1.2 and TLSv1.3.** |
| `tls_client_cert` | client certificate as string or filepath | yes | Sets the certificate for the client to use for TLS authentication. Needs to be combined with `tls_client_key`. |
| `tls_client_key` | client certificate key as string or filepath | yes | Sets the certificate key for the client to use for TLS authentication. Needs to be combined with `tls_client_cert`. |

### More?

Do you have a neat way of using this library in your Caddyfile? Feel free to submit it.

## Internals

We use the most simple commands in order to make this work and avoid managing any extra structures. Each file is stored as a Hash, with `value`, `last_modified` and `size` in order to store the content of the file and its metadata without any additional serialization. Walking through directories is simply done by doing an Scan and processing of the records in order to return valid results. This means, there is no additional command to repair any internal structures, as there are only Valkey native data structures and mostly single commands for a single action.

The Lock structure is handled by the sub-package `valkeylock` of the Valkey Go Client Library and some essential aspects are exposed via the configuration.

In regards to TLS, this module does not have any function to reload the TLS certificates while running. For this we recommend to rely on Caddy itself, using the reload functionality. This can be either achieved using the `caddy reload` command or using the reload function for your prefered system service tool.

### Exploring storage structure

If you like, you can just connect directly to the Valkey Instance you are running using the `valkey-cli` and explore the storage structure. For this, simply connect to the instance you configured and move around with the following commands:

```bash
$ valkey-cli -u valkey://localhost:6379/0
# Or using docker/podman/...: docker run -it --network=host ghcr.io/valkey-io/valkey:alpine valkey-cli -u valkey://localhost:6379/0
127.0.0.1:6379> KEYS *
1) "pki/authorities/local/intermediate.key"
2) "certificates/local/helloworld.localhost/helloworld.localhost.key"
3) "certificates/local/helloworld.localhost/helloworld.localhost.crt"
4) "pki/authorities/local/intermediate.crt"
5) "pki/authorities/local/root.key"
6) "certificates/local/helloworld.localhost/helloworld.localhost.json"
7) "last_clean.json"
8) "pki/authorities/local/root.crt"
127.0.0.1:6379> HGETALL "certificates/local/helloworld.localhost/helloworld.localhost.crt"
1) "size"
2) "1356"
3) "last_modified"
4) "2026-03-21T11:09:50+01:00"
5) "value"
6) "-----BEGIN CERTIFICATE-----\nMIIByjCCAW+gAwIBAgIRAMaMVn4X3VJ0bkrymg6Z/uAwCgYIKoZIzj0EAwIwMzEx\nMC8GA1UEAxMoQ2FkZHkgTG9jYWwgQXV0aG9yaXR5IC0gRUNDIEludGVybWVkaWF0\nZTAeFw0yNjAzMjExMDA5NTBaFw0yNjAzMjExMTA5NTBaMAAwWTATBgcqhkjOPQIB\nBggqhkjOPQMBBwNCAAT/QMUM4eDjuXPb9i/qAL68oT8niY5fIUOrRwm1pOEBO3KA\nzTKIZGXjSMsg0JpMD1af07D1xikwQx37FZA+WN/Bo4GWMIGTMA4GA1UdDwEB/wQE\nAwIHgDAdBgNVHSUEFjAUBggrBgEFBQcDAQYIKwYBBQUHAwIwHQYDVR0OBBYEFPy9\nEDChE1UHfI/YtDiRGhKvKDVJMB8GA1UdIwQYMBaAFKaRlYPZfNOUEyaTxegIIhTv\nAEhSMCIGA1UdEQEB/wQYMBaCFGhlbGxvd29ybGQubG9jYWxob3N0MAoGCCqGSM49\nBAMCA0kAMEYCIQDER26QI8bjwICzJAdXnXX1LjwUTpVZBjknofIZ12FJWAIhAIpI\nUzC757DYpWaIhXmJFlj2GR/Q/lEAyYVXdXFSPImn\n-----END CERTIFICATE-----\n-----BEGIN CERTIFICATE-----\nMIIBxzCCAW6gAwIBAgIRAOlTgRnSVthuSnjKrnFOBCAwCgYIKoZIzj0EAwIwMDEu\nMCwGA1UEAxMlQ2FkZHkgTG9jYWwgQXV0aG9yaXR5IC0gMjAyNiBFQ0MgUm9vdDAe\nFw0yNjAzMjExMDA5MDhaFw0yNjAzMjgxMDA5MDhaMDMxMTAvBgNVBAMTKENhZGR5\nIExvY2FsIEF1dGhvcml0eSAtIEVDQyBJbnRlcm1lZGlhdGUwWTATBgcqhkjOPQIB\nBggqhkjOPQMBBwNCAASPOvhKfMfrNcwka3cDa0E08XyAKwsQhADI5SfXRProjb1t\nkN8vqnxxrbh65C3c3txheOZ15xyEyK+o/5X5NJ9yo2YwZDAOBgNVHQ8BAf8EBAMC\nAQYwEgYDVR0TAQH/BAgwBgEB/wIBADAdBgNVHQ4EFgQUppGVg9l805QTJpPF6Agi\nFO8ASFIwHwYDVR0jBBgwFoAUcKWmo+Y10saoMPHtIHqTFQWbYMMwCgYIKoZIzj0E\nAwIDRwAwRAIgboxCiOT+hbLXEwsOyyAkU+4UI/WozNfcKKKncCeYiYcCIDEK1vVN\nkxC9vmL4HL5lNAuls2pRts0YDi27qu/2BOPr\n-----END CERTIFICATE-----\n"
```

### Resources

- [Valkey CLI](https://valkey.io/topics/cli/)
- [Valkey Hash-specific Commands](https://valkey.io/commands/#hash)
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

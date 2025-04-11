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
> Currently there aren't all configuration options exposed to the caddy configuration. Depending on the complexity, they will be slowly added to support all technically possible configuration options.

#### Examples

```
# Connecting to single valkey node
storage valkey {
    # See https://github.com/redis/redis-specifications/blob/1252427cdbc497f66a7f8550c6b5f2f35367dc92/uri/redis.txt
    url redis://localhost:6379/0

    lock_majority 1
    disable_client_cache true
}

storage valkey {
    address 127.0.0.1:6379

    db 0

    lock_majority 1
    disable_client_cache true
}

# Connecting to standalone valkey with replicas
storage valkey {
    # See https://github.com/redis/redis-specifications/blob/1252427cdbc497f66a7f8550c6b5f2f35367dc92/uri/redis.txt
    url redis://localhost:6379/0

    replica {
        localhost:6376
    }

    lock_majority 1
    disable_client_cache true
    send_to_replicas readonly
}

storage valkey {
    address 127.0.0.1:6379

    replica {
        localhost:6376
    }

    db 0

    lock_majority 1
    disable_client_cache true
    send_to_replicas readonly
}

# Connecting to valkey cluster
storage valkey {
    # See https://github.com/redis/redis-specifications/blob/1252427cdbc497f66a7f8550c6b5f2f35367dc92/uri/redis.txt
    url redis://localhost:7001?addr=localhost:7002&addr=localhost:7003

    shuffle_init true

    lock_majority 2
    disable_client_cache true
}

storage valkey {
    address {
        127.0.0.1:7001
        127.0.0.1:7002
        127.0.0.1:7003
    }

    shuffle_init true

    lock_majority 2
    disable_client_cache true
}

# Connecting to valkey sentinels
storage valkey {
    # See https://github.com/redis/redis-specifications/blob/1252427cdbc497f66a7f8550c6b5f2f35367dc92/uri/redis.txt
    url redis://localhost:7001?addr=localhost:7002&addr=localhost:7003

    sentinel_master_set my_master

    lock_majority 2
    disable_client_cache true
}

storage valkey {
    address {
        127.0.0.1:7001
        127.0.0.1:7002
        127.0.0.1:7003
    }

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

### Testing

In order to test locking and the load the setup can handle, there is a testing script in [`./scripts/generate_benchmark.sh`](./scripts/generate_benchmark.sh) which will generate an Caddyfile with a huger number of domains for which internally signed certificates are generated with a lifetime of 1 hour and a storage cleanup intervall of 60 seconds, to stress this storage module.

> A test with 10.000 domains, showed the cleanup to take about 90 seconds to finish. There haven't been any long running tests yet or other more extreme tests.

## License

![GitHub License](https://img.shields.io/github/license/oltdaniel/caddy-storage-valkey)
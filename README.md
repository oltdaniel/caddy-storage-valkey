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
        address {
            127.0.0.1:6379
        }
        select_db 0
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
storage valkey {
    address 127.0.0.1:6379
    select_db 0
}

storage valkey {
    address {
        127.0.0.1:6379
        127.0.0.1:6380
    }

    select_db 0
}
```

#### Values

| Name | Values | Description |
|-|-|-|
| `address` | single or list of valkey servers | This option accepts a single or a list of valkey server addresses in any format supported by the valkey go client `InitAddress` option. |
| `select_db` | valid integer for selecting the valkey database | The range of a valid value in this case depends on your server configuration. Typical range is `0-15` (total 16). |

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

Additionally, there is a [`docker-compose.yml`](./docker-compose.yml) which contains a demo setup for many different valkey go client compatible clients which can be spun up for testing.

### Testing

In order to test locking and the load the setup can handle, there is a testing script in [`./scripts/generate_benchmark.sh`](./scripts/generate_benchmark.sh) which will generate an Caddyfile with a huger number of domains for which internally signed certificates are generated with a lifetime of 1 hour and a storage cleanup intervall of 60 seconds, to stress this storage module.

> A test with 10.000 domains, showed the cleanup to take about 90 seconds to finish. There haven't been any long running tests yet or other more extreme tests.

## License

![GitHub License](https://img.shields.io/github/license/oltdaniel/caddy-storage-valkey)
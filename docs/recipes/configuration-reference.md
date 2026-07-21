---
title: Configuration reference
description: Every config.yml setting and command-line flag AppservR reads.
---

# Configuration reference

`config.yml` is written next to the executable the first time you run AppservR, populated with defaults. Settings here need a server restart to take effect; everything else (apps, users, groups, access control) is managed live through the admin web interface.

## Settings

| Key | Default | Description |
|---|---|---|
| `server.host` | `localhost` | Address to listen on. Set to `0.0.0.0` to accept connections from other machines. |
| `server.port` | `8080` | Port to listen on. |
| `RScript` | `Rscript` on Linux, auto-detected under `C:/Program Files/R` on Windows | Path to the `Rscript` executable used to launch Shiny apps. Set this explicitly if R isn't in a standard location, or if you want to point at a specific R installation. |
| `database.path` | `<executable folder>/data.db` | Path to the SQLite database file holding apps, users, and groups. |
| `mode` | `prod` | `prod` logs at WARNING level; `debug` logs at DEBUG level. |

`database.type` also exists in the file but `sqlite` is the only backend AppservR supports today, so it shouldn't be changed.

## Command-line flags

The `serve` command accepts flags that override the corresponding `config.yml` setting for that run, without writing them back to the file:

```sh
./appservR serve -a 0.0.0.0 -p 9090 -m debug
```

| Flag | Long form | Overrides |
|---|---|---|
| `-a` | `--address` | `server.host` |
| `-p` | `--port` | `server.port` |
| `-m` | `--mode` | `mode` |

## Environment variables

| Variable | Effect |
|---|---|
| `APPSERVR_AUTH_SECRET` | Signing secret for login sessions. See [Running in production →](production-considerations.md) for why you'll want to set this. |

---
title: "Running in production"
description: "Session persistence, backups, and upgrading an AppservR install."
lead: "Session persistence, backups, and upgrading an AppservR install."
date: 2026-07-20T00:00:00+00:00
lastmod: 2026-07-20T00:00:00+00:00
draft: false
images: []
menu:
  docs:
    parent: "recipes"
weight: 230
toc: true
---

## Login sessions

Logins last 15 minutes, and by default that clock resets for everyone the moment the server restarts: AppservR signs login sessions with a random secret generated when the process starts, so restarting the service invalidates every session that was still active, forcing a fresh login.

To keep sessions valid across restarts, set the `APPSERVR_AUTH_SECRET` environment variable to a fixed value before starting AppservR (as a service, this is set in the service's environment, not in `config.yml`). Any long random string works; the important part is that it stays the same across restarts and isn't shared outside your team.

## Data and backups

Everything AppservR needs to restore an install lives in two places next to the executable:

* `config.yml`, the server configuration.
* the SQLite database file (`data.db` by default, see [Configuration reference →]({{< relref "configuration-reference" >}}) for how to change its location), holding every app, user, and group.

Your apps' own R source directories are wherever you pointed each app's **Shiny app directory** field, so back those up too if they aren't already under version control elsewhere.

## Upgrading

To upgrade, stop the service, replace the executable with the new version, and start it again. `config.yml` and the database file are untouched by the upgrade itself; AppservR runs its database migrations automatically on startup, so there's no separate migration step to run.

As always, back up the database file before upgrading across a major version.

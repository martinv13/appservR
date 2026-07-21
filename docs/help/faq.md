---
title: FAQ
description: Answers to frequently asked questions.
---

# FAQ

## Why am I logged out every time the server restarts?

By default, AppservR signs login sessions with a secret it generates fresh each time it starts, so every restart invalidates existing sessions. Set the `APPSERVR_AUTH_SECRET` environment variable to a fixed value to keep sessions valid across restarts. See [Running in production →](../recipes/production-considerations.md).

## Can I deploy an app straight from a Git repository?

Not yet. The admin UI has a "Git repository" option in the app source picker, but it's disabled and marked "not available yet" — only a local or network directory works today. See [Configure your first app →](../recipes/configure-first-app/index.md).

## What platforms does AppservR run on?

Prebuilt binaries are published for Windows (32 and 64-bit) and Linux (64-bit). There's no macOS build.

## My app isn't showing up, or seems stuck. Where do I look?

Each app's own page in the admin interface (`/admin/apps/<name>`) shows the console output of its running R workers. A worker that started successfully prints a "Listening on" line; anything else printed there (an R error, a missing package) is usually the actual cause. See [Configure your first app →](../recipes/configure-first-app/index.md).

## Can several apps run on the same server?

Yes. Each app gets its own path, and paths can be nested (`/` and `/sales-dashboard` and `/sales-dashboard/internal` can all coexist). Each app also runs its own R worker processes, independent from every other app's.

## How do I move AppservR to a new server, or back it up?

Back up `config.yml` and the SQLite database file next to the executable, plus whichever directories your apps' source code lives in. See [Running in production →](../recipes/production-considerations.md).

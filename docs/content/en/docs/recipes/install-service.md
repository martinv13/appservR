---
title: "Install as a service"
description: "AppservR can be installed as a service to run automatically in the background."
lead: "AppservR can be installed as a service to run automatically in the background."
date: 2020-10-13T15:21:01+02:00
lastmod: 2020-10-13T15:21:01+02:00
draft: false
images: []
menu:
  docs:
    parent: "recipes"
weight: 220
toc: true
---

## Manage service via the CLI

The `appservr service` command allows installing and managing the AppservR service on Windows an Linux.

{{< alert icon="💡" text="Managing services requires to run a command prompt as admin on Windows." />}}

### Install

`appservr service install` sub-command installs AppservR as a service (running the executable file at its current location).

On Windows, only admins can install services, and services run with administrative privileges by default. We recommend that you change this, using the services configuration panel after install, in order to have it ran by a non-admin account. Otherwise, you allow AppservR admin to run arbitrary scripts as admin, which is a security threat (AppservR uses your system installation of R and is not isolated from your system).

### Start

`appservr service start` starts the service.

### Stop

`appservr service stop` stops the service.

### Uninstall

`appservr service remove` uninstalls the service if it is already stopped, or schedule removal right after it is stopped.

## Run without admin rights on Windows

Without admin privilege you cannot install AppservR as a service on Windows. However, you can still run it as a scheduled task to run on startup or at session start, which should be a decent alternative in most cases.

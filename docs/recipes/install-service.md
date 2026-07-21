---
title: Install as a service
description: AppservR can be installed as a service to run automatically in the background.
---

# Install as a service

## Manage service via the CLI

The `./appservR service` command allows installing and managing the AppservR service on Windows and Linux.

!!! tip
    Managing services requires to run a command prompt as admin on Windows.

### Install

`./appservR service install` sub-command installs AppservR as a service (running the executable file at its current location).

On Windows, only admins can install services, and services run with administrative privileges by default. We recommend that you change this, using the services configuration panel after install, in order to have it run by a non-admin account. Otherwise, you allow AppservR admins to run arbitrary scripts as admin, which is a security threat (AppservR uses your system installation of R and is not isolated from your system).

### Start

`./appservR service start` starts the service.

### Stop

`./appservR service stop` stops the service.

### Uninstall

`./appservR service remove` uninstalls the service if it is already stopped, or schedules its removal for right after it stops.

## Run without admin rights on Windows

Without admin privilege you cannot install AppservR as a service on Windows. However, you can still run it as a scheduled task to run on startup or at session start, which should be a decent alternative in most cases.

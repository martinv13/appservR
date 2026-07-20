---
title: "Configure your first app"
description: "Walkthrough of the app configuration form: name, path, access control, source directory, and workers."
lead: "Walkthrough of the app configuration form: name, path, access control, source directory, and workers."
date: 2026-07-20T00:00:00+00:00
lastmod: 2026-07-20T00:00:00+00:00
draft: false
images: []
menu:
  docs:
    parent: "recipes"
weight: 195
toc: true
---

Once you're logged in as an admin, go to [http://localhost:8080/admin/apps](http://localhost:8080/admin/apps) and click **New App**. This walks through each field on that form.

![The apps list, with a couple of apps already configured](admin-apps-list.png)

## Main settings

![The new app form, filled in](admin-app-form.png)

* **Name of the app**: letters, numbers, hyphens, and underscores only. This is also the name shown on the apps list.
* **Path of the app**: where the app is served, for example `/sales-dashboard`. It must start with `/` and can't be `/admin` or `/auth`, since those are reserved for AppservR itself. Several apps can share the same server on different paths, including nested ones (`/` and `/sales-dashboard` and `/sales-dashboard/internal` can all coexist).
* **Is active**: unchecked apps are configured but not running, and won't use any of your R workers.
* **Grant access to**: who can reach the app once it's active.
  * **Everyone**: public, no login required.
  * **All authenticated users**: any user who signed up on this AppservR instance.
  * **Specific user groups**: pick from the groups you've created under the Groups tab.

## App source

AppservR currently only supports a **local or network directory** as an app's source: a path to a folder containing either `app.R`, or both `server.R` and `ui.R`. This can be a path on the machine AppservR runs on, or a network share it can read. A **Git repository** source is planned but not available yet, which is why that option is disabled in the form.

## Serving

**Number of process workers** sets how many separate `Rscript` processes AppservR keeps running for this app. Each user session is bound to one worker for the length of their session, so more workers mean more users can be served in parallel. Since Shiny apps typically hold their state in server-side R process memory, this is also what lets AppservR scale a single app beyond what one R process could handle alone.

## After saving

Once saved, the app's own page shows the console output of each running worker, which is the first place to look if an app isn't starting (see the "Listening on" line below, confirming the worker actually started).

![An active app's detail page, showing console output from a running worker](admin-app-detail-running.png)

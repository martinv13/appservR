---
title: "Manage user authentication"
description: "AppservR manages user authentication and access restriction to your apps."
lead: "AppservR manages users authentication and access restriction to your apps."
date: 2020-10-13T15:21:01+02:00
lastmod: 2020-10-13T15:21:01+02:00
draft: false
images: []
menu:
  docs:
    parent: "recipes"
weight: 210
toc: true
---

## Manage users

You can create new users via the admin interface, or users can sign up themselves.

Users belong to groups that you can create through the admin interface. By default, the first user to sign up belongs to the `admins` group. The `admins` group cannot be renamed nor removed, and admin users cannot remove themselves from this group.

## Access control

You can control who is allowed to access each app:

* Everyone
* All authenticated users
* Users belonging to specific groups

This setting can be found in each app page in the admin interface.

## Get authenticated user from your app

You R app can be aware of the authentication infos for the current user, if any. This information is accessible in the HTTP headers sent to your app by the AppservR proxy.

The default app (*Old Faithful Geyser Data*) demonstrates how to use this information to identify the user currently logged in.

From Shiny, username and displayed name can be accessed from your server function as follows:

``` r
server <- function(input, output, session) {

    output$username <- renderText({
        if (exists("HTTP_APPSERVR_USERNAME", envir=session$request)) {
            get("HTTP_APPSERVR_USERNAME", envir=session$request)
        } else {
            "unknown"
        }
    })
        
    output$displayed_name <- renderText({
        if (exists("HTTP_APPSERVR_DISPLAYEDNAME", envir=session$request)) {
            get("HTTP_APPSERVR_DISPLAYEDNAME", envir=session$request)
        } else {
            "unknown"
        }
    })
    
    ...

}

```

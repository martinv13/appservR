---
title: "About"
description: "About AppservR"
lead: "About AppservR"
date: 2020-11-12T13:26:54+01:00
lastmod: 2020-11-12T13:26:54+01:00
draft: false
images: []
menu:
  docs:
    parent: "help"
weight: 310
toc: true
---

## Why this project?

There are already some alternatives to deploy R Shiny apps:

* [shinyapps.io](https://shinyapps.io) is Posit's own hosted service: no server to manage at all, but you don't control the environment and usage is metered.
* [Shiny Server](https://www.rstudio.com/products/shiny/shiny-server/) (the free, open-source edition) is self-hosted like AppservR, but Linux-only, and has no built-in authentication (that's a Shiny Server Pro feature).
* [ShinyProxy](https://www.shinyproxy.io/) runs each user session in its own Docker container, which is powerful for isolation and scaling, but means taking on a Docker deployment to get there.

and maybe some others.

AppservR's trade-off is different: a single cross-platform binary with authentication built in and no container runtime required, at the cost of some of the isolation and scaling headroom those other tools offer.

However, working in an organization where data is not the primary focus and with limited IT admin resources (100% Windows!), I found no easy way to share Shiny apps internally as proof of concept.

AppservR is intended as a solution for simple deployment with limited admin resources, without conceding performance.

Also it was a fun project to start with the Go programming language during COVID-19 lockdowns weekends.

## About this website

This documentation website was built using [Doks](https://getdoks.org/docs/prologue/introduction/), an amazing theme for [Hugo](https://gohugo.io/).

You can contribute if you think that some content is missing clarity etc. on this [GitHub repository](https://github.com/appservR/appservR/issues).

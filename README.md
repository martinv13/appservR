# AppservR

AppservR allows deploying R Shiny applications easily on your server. Distributed as a standalone binary with no dependency except R itself, it can be used on any platform to serve Shiny apps to your users.

Under the hood, AppservR manages several instances of your apps, using Rscript executable, and proxy requests from your users to these instances.

## Features

As R is a single-process software, deploying R apps often requires third-party software to allow dealing with multiple clients at the same time. AppservR achieves this with a ready-to-use, cross-platform binary with no dependencies. It aims at simplicity, while not giving up on performance. Its distinctive features compared to other solutions are the following:

* **Cross platform**: written in the Go programming language, it runs on most platforms (compiled versions available for Windows and Linux),
* **No dependencies**: AppservR does not require Java, Docker or other dependency; it ships as a standalone binary which simply runs your apps using the Rscript executable from your R installation and proxy client requests accordingly,
* **Authentication**: built in authentication to restrict access to specific apps or customize user experience,
* **Unlimited apps**: you can run as many apps as you wish on different paths, including nested paths (i.e. for instance "/" and "/myapp" and "/myapp/private"),
* **Hot config**: you can configure your apps through a web interface without restarting the server, which means that you do not need admin permissions on the server to add or update a Shiny app, but only a AppservR admin account.

## Install

* Download the latest release for your platform from the [releases](https://github.com/appservR/appservR/releases/latest) page.
* Extract the downloaded archive at a location of your choice.
* Run the executable.
* Navigate to `http://localhost:8080` to see a demo Shiny app running!

Please check out our [Documentation website](https://appservR.github.io) for more details.

## What is Shiny anyway?

[Shiny](https://shiny.rstudio.com/) is an amazing R package created and maintained primarily by [Rstudio](https://rstudio.com) to build interactive web applications with the R programming language. It is more flexible than most BI software, and much easier to set up compared to a "complete" web app (backend + frontend). Also, it is free and open-source. Whether you want to build simple business dashboards or full-featured data apps, Shiny is probably worth looking at.

## Alternatives

[Rstudio](https://rstudio.com) offer professional solutions to Shiny apps' deployment (no free version for multi-process server however). 

[ShinyProxy](https://www.openanalytics.eu/tags/shinyproxy/) is another open-source solution with a different approach as it runs a new Shiny app instance in a new Docker container for every client of your app.

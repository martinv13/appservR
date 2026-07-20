---
title: "Quick Start"
description: "A one page tutorial to use AppservR."
lead: "A one page tutorial to use AppservR."
date: 2020-11-16T13:59:39+01:00
lastmod: 2020-11-16T13:59:39+01:00
draft: false
images: []
menu:
  docs:
    parent: "introduction"
weight: 110
toc: true
---

## Requirements

AppservR runs your Shiny apps using your system installation of R. You need to have R installed on your system to use it. You will also need the `shiny` package and all other R packages your apps need. [Learn more about Shiny →](https://shiny.posit.co/)

{{< alert icon="💡" text="Make sure first that you are able to run your Shiny app properly using R/Rstudio to check that all required packages are installed." />}}

AppservR itself ships as prebuilt binaries for Windows (32 and 64-bit) and Linux (64-bit). There's no macOS build at the moment.

## Download & Install

Get the latest AppservR executable for your platform [on Github](https://github.com/appservR/appservR/releases/latest). You will need to choose the download option suited to your platform.

{{< rawhtml >}}
<p><a class="btn btn-primary" href="https://github.com/appservR/appservR/releases/latest">Download</a></p>
{{< /rawhtml >}}

To install, just uncompress the archive at a location of your choice.  

## Run

AppservR is a command-line program with a web graphical user interface. Most settings can be configured using only the web interface, which means that you do not need to access the server to configure a new Shiny app.

Start a terminal window in the directory where AppservR executable file is and then type:

``` bash
./appservR serve 
```

Go to [http://localhost:8080](http://localhost:8080). You should see a running sample app (the classic Shiny's *Old Faithful Geyser Data*, with a twist: it also displays the username of whoever is logged in, see [Manage user authentication →]({{< relref "manage-auth" >}})). If not, see below.

## Configure

When you run AppservR for the first time, it will create in the same directory as the executable file:

* a `config.yml` configuration file,
* an `apps` folder containing a sample Shiny app,
* a SQLite database file holding apps and users settings.

Settings which require restarting the server to take effect are defined in the `config.yml` file. Other settings can be changed via the web interface.

### Configuration file

`config.yml` configuration file is created in the same directory as the executable file, and allows setting the port and host to listen on, the location of the Rscript executable, etc. It is populated with default settings which should work for most users. See the full list of settings in [Configuration reference →]({{< relref "configuration-reference" >}}).

In particular, AppservR will look for the *Rscript* executable at the default installation location for your platform. If it is installed elsewhere, you will not see the sample app when navigating to [http://localhost:8080](http://localhost:8080), and you should set the appropriate location in the `config.yml` file.

### Web interface

Go to [http://localhost:8080/admin](http://localhost:8080/admin) to access the admin interface.

You will first need to register as a user via the login > signup form. The first user to sign up is registered as an admin user.

After you have created your admin account and successfully logged in, you can configure existing or new Shiny apps at [http://localhost:8080/admin/apps](http://localhost:8080/admin/apps). Several apps can live on the same server and be accessed using the path you configure. See [Configure your first app →]({{< relref "configure-first-app" >}}) for a walkthrough of the form.

On each app configuration page you can also see the console output of the R session used to run your app, which can be useful to troubleshoot issues.

## Go further

Check our tutorials on more advanced uses of AppservR. [Recipes →]({{< relref "recipes" >}})

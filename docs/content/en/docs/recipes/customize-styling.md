---
title: "Customize styling"
description: "Built-in pages (login, etc.) can be customized to match your image."
lead: "Built-in pages (login, etc.) can be customized to match your image."
date: 2020-10-13T15:21:01+02:00
lastmod: 2020-10-13T15:21:01+02:00
draft: false
images: []
menu:
  docs:
    parent: "recipes"
weight: 200
toc: true
---

## Customize html templates

All buit-in pages, such as signup, login, error pages, are embeded in AppservR executable. However, they can be changed easily if you want to translate them to your language, add your company's logo, etc.

You just need to create a `templates` directory at the same location as the executable file. AppservR will look for existing files in this directory before falling back to the default templates.

Please check out [this folder on Github](https://github.com/appservR/appservR/tree/main/templates) and download the templates files that you want to modify in your local `templates` folder (keeping the exact same subdirectories and filenames). The template engine is [Go's default template package](https://pkg.go.dev/text/template). Please check out their doc to learn about the syntax. However, you would likely just need to modify the html.

## Add/modify static files

In a similar way as the `templates` folder, you can create an `assets` folder at the same location as AppservR executable. Static files will be looked up in this directory, fall back to default bundle files if they do not exist, and return a `404 not found` error if the file is not present neither in your `assets` local folder or in the default bundled [`assets`](https://github.com/appservR/appservR/tree/main/assets) folder.

This can be useful to add external javascript, css or images.

The content of the `assets` folder is served under the `/assets/` url path.

Beware that static files will be served to all users, and you cannot manage access to this folder.

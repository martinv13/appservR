---
title: Deploying R apps the easy way
description: AppservR deploys R Shiny apps on Windows and Linux from a single binary, no admin rights or infrastructure team required.
---

# Deploying R apps the easy way

AppservR deploys R Shiny apps on Windows and Linux from a single binary, no admin rights or infrastructure team required.

[Get Started :material-arrow-right:](introduction/welcome.md){ .md-button .md-button--primary }

Open-source MIT Licensed. [GitHub](https://github.com/appservR/appservR)

<div class="grid cards" markdown>

-   **Unlimited apps**

    Serves as many R apps as you need. Each app runs its own Rscript instances, using the R installation on your system.

-   **Access management**

    Authentication included: you can restrict access to each app to specific users and access authenticated user info from within your app.

-   **Hot config**

    Most settings can be configured through a web admin interface and do not require accessing or restarting the server.

-   **Cross-platform**

    Written in the Go programming language, AppservR can run on most platforms. Binaries for Windows and Linux are available.

-   **No dependencies**

    Ships as a standalone binary. No need to run Java or Docker; you will just need a working R installation.

-   **Multi-process**

    Apps with higher traffic can be configured to use multiple processes and balance load between app instances.

</div>

## Who it's for

AppservR is built for small teams that need to share R Shiny apps with other team members, without a platform team, a container registry, or admin rights on a server to get there. If you can copy a folder onto a Windows or Linux machine and start an executable, you can run AppservR.

It's also resource-efficient: each app runs a small pool of R processes, and new users are routed to whichever one currently has the fewest people on it, so several users typically share the same process instead of each needing their own. That's a different trade-off than giving every user session its own container, and it means a modest server can support more concurrent users than you might expect.

That combination matters most in IT environments with real constraints: no Docker, no Java runtime to install, no separate reverse proxy to configure and keep up to date. One binary handles process management, authentication, and routing for every app you add. Getting a proof of concept running on a shared machine that your team can reach takes minutes, not a change request.

AppservR is not meant to replace a full production deployment platform built for heavy traffic. For that, dedicated commercial or container-based solutions are a better fit. But for testing an idea, sharing a dashboard with a handful of users, or proving a concept before asking for infrastructure budget, it removes most of the setup cost.

[Get started :material-arrow-right:](introduction/welcome.md){ .md-button .md-button--primary }
[Recipes :material-arrow-right:](recipes/index.md){ .md-button }

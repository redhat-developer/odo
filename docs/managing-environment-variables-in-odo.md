---
# Page settings
layout: default
keywords:
comments: false

# Hero section
title: Managing environment variables
description: Manipulate both config and preferences files to your liking

# Micro navigation
micro_nav: true

# Page navigation
page_nav:
    prev:
        content: Debugging applications in odo
        url: '/docs/debugging-applications-in-odo'
    next:
        content: Configuring the odo CLI
        url: '/docs/configuring-the-odo-cli'
---
`odo` stores component-specific configurations and environment variables
in the `config` file. You can use the `odo config` command to set,
unset, and list environment variables for components without the need to
modify the `config` file.

# Setting and unsetting environment variables

  - To set an environment variable in a component:
    
        $ odo config set --env <variable>=<value>

  - To unset an environment variable in a component:
    
        $ odo config unset --env <variable>

  - To list all environment variables in a component:
    
        $ odo config view

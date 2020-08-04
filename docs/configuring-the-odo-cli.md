---
# Page settings
layout: default
keywords:
comments: false

# Hero section
title: Configuring the odo CLI
description: Configure your terminal for autocompletion

# Micro navigation
micro_nav: true
---
# Using command completion

> **Note**
> 
> Currently command completion is only supported for bash, zsh, and fish
> shells.

odo provides a smart completion of command parameters based on user
input. For this to work, odo needs to integrate with the executing
shell.

  - To install command completion automatically:
    
    1.  Run:
        
            $ odo --complete
    
    2.  Press `y` when prompted to install the completion hook.

  - To install the completion hook manually, add `complete -o nospace -C
    <full path to your odo binary> odo` to your shell configuration
    file. After any modification to your shell configuration file,
    restart your shell.

  - To disable completion:
    
    1.  Run:
        
            $ odo --uncomplete
    
    2.  Press `y` when prompted to uninstall the completion hook.

> **Note**
> 
> Re-enable command completion if you either rename the odo executable
> or move it to a different directory.

# Ignoring files or patterns

You can configure a list of files or patterns to ignore by modifying the
`.odoignore` file in the root directory of your application. This
applies to both `odo push` and `odo watch`.

If the `.odoignore` file does *not* exist, the `.gitignore` file is used
instead for ignoring specific files and folders.

To ignore `.git` files, any files with the `.js` extension, and the
folder `tests`, add the following to either the `.odoignore` or the
`.gitignore` file:

    .git
    *.js
    tests/

The `.odoignore` file allows any glob expressions.

---
# Page settings
layout: default
keywords:
comments: false

# Hero section
title: Setting up a secure Devfile registry
description: Learn how to setup a secure private registry that only you or your team can access

# Micro navigation
micro_nav: true

---
# Introduction to secure devfile registry

**What is a secure devfile registry?**

A secure devfile registry is a devfile registry that a user can only access using credentials.

**Where to host secure devfile registry?**

A user can host a secure devfile registry on a private GitHub repository or an enterprise GitHub repository.

# Adding a secure devfile registry on a GitHub repository

1.  Creating new GitHub repository to host the secure devfile registry:
    
    Please [create a new private or enterprise GitHub repository](https://docs.github.com/en/github/creating-cloning-and-archiving-repositories/creating-a-new-repository) and push the devfile registry to the created repository. The sample GitHub-hosted devfile registry can be found [here](https://github.com/odo-devfiles/registry/).

2.  Creating a personal access token to access the secure devfile registry
    
    Please [create a personal access token](https://docs.github.com/en/github/authenticating-to-github/creating-a-personal-access-token), select `repo` as token scope.

3.  Keyring setup
    
    There is no specific keyring setup for secure devfile registry, you only need to ensure the keyring which is working properly on your system, if you hit issues please follow the below instructions to troubleshoot the issues of your keyring with respect to the corresponding platforms.
    
      - [Mac keychain](https://support.apple.com/en-ca/guide/keychain-access/welcome/mac)
    
      - [GNOME keyring setup on RedHat Enterprise Linux](https://nurdletech.com/linux-notes/agents/keyring.html)
    
      - [GNOME keyring setup on Ubuntu Linux](https://howtoinstall.co/en/ubuntu/xenial/gnome-keyring)
    
      - [Linux GNOME keyring](https://help.gnome.org/users/seahorse/stable/index.html.en)
    
      - [Windows credential manager](https://support.microsoft.com/en-ca/help/4026814/windows-accessing-credential-manager)

4.  Adding secure devfile registry
    
    Please run `odo registry add <registry name> <registry URL> --token <token>` to add the secure devfile registry to odo, for more registry related commands please refer to `odo registry --help`.
    
      - \<registry name\>: user-defined devfile registry name.
    
      - \<registry URL\>: the URL of GitHub repository that you create on step 1.
    
      - \<token\>: the personal access token that you created on step 2.

# Steps for setting up a secure starter project on a GitHub repository

1.  Creating a new GitHub repository to host the secure starter project
    
      - Please [create a new private or enterprise GitHub repository](https://docs.github.com/en/github/creating-cloning-and-archiving-repositories/creating-a-new-repository) and push the starter project to the created repository. The sample GitHub-hosted starter project can be found [here](https://github.com/odo-devfiles/nodejs-ex).
    
      - Ensure the `starterProjects` section in the corresponding devfile of your secure devfile registry links to the secure starter project, for example:
        
            starterProjects:
              - name: nodejs-starter
                git:
                  remotes:
                    origin: "<secure starter project link>"

2.  Creating a personal access token to access the secure starter project
    
    Please [create a personal access token](https://docs.github.com/en/github/authenticating-to-github/creating-a-personal-access-token), select `repo` as token scope.

3.  Creating a devfile component from the secure devfile registry and downloading the secure starter project
    
    Please run `odo create nodejs --registry <registry name> --starter --starter-token <starter project token>`
    
      - \<registry name\>: user-defined devfile registry name.
    
      - \<starter project token\>: the personal access token that you create on step 2.

> **Note**
> 
> GitHub only supports user-scoped personal access tokens. If the repository that hosts the secure registry and the repository that hosts the secure starter project are created under the same GitHub user, then the token can be used for both downloading the devfile and starter project. For that case you don’t need to explicitly pass in the flag `--starter-token <starter project token>`, odo can automatically use one token to download both devfile and starter project.

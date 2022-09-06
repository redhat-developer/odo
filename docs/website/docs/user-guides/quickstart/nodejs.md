---
title: Developing with Node.JS
sidebar_position: 1
---

## Step 0. Creating the initial source code (optional)

import InitialSourceCodeInfo from './_initial_source_code.mdx';

<InitialSourceCodeInfo/>


For Node.JS we will use the [Express](https://expressjs.com/) framework for our example.

1. Install Express:
```console
npm install express --save
```

2. Generate an example project:
```console
npx express-generator
```
```console
$ npx express-generator
  warning: the default view engine will not be jade in future releases
  warning: use `--view=jade' or `--help' for additional options


   create : public/
   create : public/javascripts/
   create : public/images/
   create : public/stylesheets/
   create : public/stylesheets/style.css
   create : routes/
   create : routes/index.js
   create : routes/users.js
   create : views/
   create : views/error.jade
   create : views/index.jade
   create : views/layout.jade
   create : app.js
   create : package.json
   create : bin/
   create : bin/www

   install dependencies:
     $ npm install

   run the app:
     $ DEBUG=express:* npm start
```

Your source code has now been generated and created in the directory.

## Step 1. Connect to your cluster and create a new namespace or project

import ConnectingToCluster from './_connecting_to_cluster.mdx';

<ConnectingToCluster/>

## Step 2. Initializing your application (`odo init`)

import CreatingApp from './_creating_app.mdx';

<CreatingApp name="nodejs" port="3000" language="javascript" framework="Node.JS"/>

## Step 3. Developing your application continuously (`odo dev`)

import RunningCommand from './_running_command.mdx';

<RunningCommand name="nodejs" port="3000" language="javascript" framework="Node.JS"/>
---
title: Developing with Node.JS
sidebar_position: 1
---

## Step 0. Creating the initial source code (optional)

import InitialSourceCodeInfo from './docs-mdx/initial_source_code_description.mdx';

<InitialSourceCodeInfo/>


For Node.JS we will use the [Express](https://expressjs.com/) framework for our example.

1. Install Express:
```console
npm install express --save
```
<details>
<summary>Example</summary>

```shell
$ npm install express --save

added 57 packages, and audited 58 packages in 6s

7 packages are looking for funding
  run `npm fund` for details

found 0 vulnerabilities
```
</details>

2. Generate an example project:
```console
npx express-generator
```
<details>
<summary>Example</summary>

```shell
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
</details>

Your source code has now been generated and created in the directory.

## Step 1. Connect to your cluster and create a new namespace or project

import ConnectingToCluster from './docs-mdx/connecting_to_the_cluster_description.mdx';

<ConnectingToCluster/>

## Step 2. Initializing your application (`odo init`)

import InitSampleOutput from './docs-mdx/nodejs/nodejs_odo_init_output.mdx';
import InitDescription from './docs-mdx/odo_init_description.mdx';

<InitDescription framework="Node.JS" initout=<InitSampleOutput/> />

## Step 3. Developing your application continuously (`odo dev`)

import DevSampleOutput from './docs-mdx/nodejs/nodejs_odo_dev_output.mdx';

import DevDescription from './docs-mdx/odo_dev_description.mdx';

<DevDescription framework="Node.JS" devout=<DevSampleOutput/> />


_You can now follow the [advanced guide](../advanced/deploy/nodejs.md) to deploy the application to production._

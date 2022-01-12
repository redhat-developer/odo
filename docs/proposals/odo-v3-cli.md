# odo v3 CLI

Over the years of odo development we picked up a lot of commands.
Odo v3 will introduce new commands (`odo dev` #5299,  `odo init` #5297). 
This will change the command flow from what we currently have in v2. We need to make sure that the whole odo CLI is consistent, and all commands follow the same pattern.

There are also some commands that are there since the original odo v1 and were originally designed for  s2i  approach only, those commands or flags should be removed, or reworked to better fit Devfile



## odo v2 commands that should be removed 

- `odo component` every command from “odo component *” already exists as a root command. 
- `odo watch` will be replaced with `odo dev`
- `odo push` will be replaced with `odo dev`
- `odo unlink` will be replaced with `odo delete binding`
- `odo link` will be replaced with `odo create binding`
- `odo url` will be replaced with `odo create endpoint`
- ? `odo test` might be replaced with a flag in `odo dev` ? 
- `odo service` will be replaced with `odo create service`
- `odo storage` if needed it will be replaced with `odo create storage`
- `odo env` should be integrated with `odo config`

## commands that should be present in v3.0.0-alpha1

### general rules for odo cli
- Once even a single flag is provided it ruins in non-interactive mode. All required information needs to be provided via flags
- If command is executed without flags it can enter interactive mode

### `odo init`
Bootstrapping completely  new project (starting from empty directory)
- [ ] #5297 

### `odo deploy`
Deploying application (outerloop)
- [ ] #5298 



### `odo dev`
Running the application on the cluster for **development** (innerloop)
- [ ] #5299 


### `odo list`
Listing "stuff" created by odo.

- without arguments lists components in local devfile and in cluster
- `odo list services` list services defined in local devfile and in cluster
- `odo list endpoints` list endpoints defined in local devfile and in cluster
- `odo list bindings` list bindings defined in local devfile and in cluster 


### `odo preference`
Configures odo behavior like timeouts.
- UpdateNotification  - keep
- NamePrefix          - remove
- Timeout             - add explanations  
- BuildTimeout        - remove
- PushTimeout         - add explanations
- Ephemeral           - keep
- ConsentTelemetry    - keep 



### `odo project`
mostly as it is

### `odo registry`
mostly as it is just drop  support "github"  registry

### `odo describe`
TODO

### `odo exec`
mostly as it is

### `odo status`
mostly as it is

## commands that will be added in v3.0.0-alpha2

### `odo create binding`


### `odo create binding`


### `odo build-images`
mostly as it is

### `odo login` / `odo logout`
mostly as it is

### `odo utils`
mostly as it is

### `odo version`
mostly as it is



### Setting environment variables
TODO

### `odo create endpoint`
TODO

### `odo create storage`
TODO

### `odo create url`
TODO

### `odo delete endpoint`
TODO

### `odo delete storage`
TODO

### `odo delete url`
TODO

### `odo app`
TODO

### `odo catalog`
TODO

### `odo config`
TODO





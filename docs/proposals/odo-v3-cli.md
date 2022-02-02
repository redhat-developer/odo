# odo v3 CLI

## TODO:

- [ ] define JSON outputs for each command

## Changes from odo v2 command perspective

Commands can be removed from odo for now. Functionality of those commands won't be there in alpha1.

- `app`
- `config`
- `catalog`
- `debug`
- `env`
- `service`
- `storage`
- `test`
- `unlink`
- `status`
- `log`
- `link`
- `exec`
- `url`
- `describe`
- `component`
  - `describe`
  - `exec`
  - `link`
  - `log`
  - `status`
  - `test`
  - `unlink`

Commands with the functionality that will be in alpha1 but the command will have a different name parameters.

- `component` - as a root command this will be removed, for info about subcommand check their root level equivalents.
  - `create`
  - `delete`
  - `list`
  - `push`
  - `watch`
- `preference` - mostly as it is just with cleanup and integration of `odo registry`
- `registry` - integrated into `odo preference`
- `watch` - new command will be `odo dev`
- `build-images` - new command will be `odo build images`
- `deploy` - mostly as it is just with some new flags
- `push` - technically this will be in `odo dev`
- `create` - new command will be `odo init` and also integrated into `odo dev` and  `odo deploy`
- `delete` - new command will be `odo delete component`
- `list` - this will be reworked, it will combine both `odo url list` and `odo list`
- `project` - new command will be `odo list/delete/create namespace`


Commmands that will remain as they are in v2

- `login`
- `logout`
- `utils`
- `version`



## odo v3 CLI

### general rules for odo cli

- Once even a single flag is provided it ruins in non-interactive mode. All required information needs to be provided via flags
- If command is executed without flags it can enter interactive mode
- every command should have `-o` flag to specify output format, and every command should support `-o json`

### CLI structure for v3.0.0-alpha1


- **[`odo login`](odo-v3-cli/odo-login-logout.md)** - no changes required
- **[`odo logout`](odo-v3-cli/odo-login-logout.md)** - no changes required
- **[`odo init`](odo-v3-cli/odo-init.md)** - [#5297](https://github.com/redhat-developer/odo/issues/5297) [#](https://github.com/redhat-developer/odo/issues/5408) new command
- **[`odo dev`](odo-v3-cli/odo-dev.md)** - [#5299](https://github.com/redhat-developer/odo/issues/5299) new command based on v2 `odo watch`
- **[`odo deploy`](odo-v3-cli/odo-deploy.md)** - [#5298](https://github.com/redhat-developer/odo/issues/5298) - mostly as it is in v2, with new interactive mode and flags.
- **[`odo preference`](odo-v3-cli/odo-preference.md)** -  [#5402](https://github.com/redhat-developer/odo/issues/5402)
mostly as it is, just cleanup
- **`odo build`**
  - **`image`** - the same as v2 `odo build-images`
- **[`odo list`](odo-v3-cli/odo-list.md)** - list everything. It combines all list outputs from all the subcommands, except namespace.
- **`odo delete`**
  - **[`component`](odo-v3-cli/odo-delete-component.md)** - similar as v2 `odo delete`, but flags and output needs to be reworked
- **`odo version`** - as it is in v2
- **`odo utils`** -  as it is in v2



### CLI structure for v3.0.0-alpha2


- [`odo login`](odo-v3-cli/odo-login-logout.md) - no changes required
- [`odo logout`](odo-v3-cli/odo-login-logout.md) - no changes required
- [`odo init`](odo-v3-cli/odo-init.md) - [#5297](https://github.com/redhat-developer/odo/issues/5297) [#](https://github.com/redhat-developer/odo/issues/5408) new command
- [`odo dev`](odo-v3-cli/odo-dev.md) - [#5299](https://github.com/redhat-developer/odo/issues/5299) new command based on v2 `odo watch`
- [`odo deploy`](odo-v3-cli/odo-deploy.md) - [#5298](https://github.com/redhat-developer/odo/issues/5298) - mostly as it is in v2, with new interactive mode and flags.
- [`odo preference`](odo-v3-cli/odo-preference.md) -  [#5402](https://github.com/redhat-developer/odo/issues/5402)
mostly as it is, just cleanup
- **[`odo config`](odo-v3-cli/odo-config.md)** - TODO needs to be reworked
- `odo build`
  - `image` - the same as v2 `odo build-images`
- **[`odo list`](odo-v3-cli/odo-list.md)** - list everything. It combines all list outputs from all the subcommands, except namespace.
  - **[`component`](odo-v3-cli/odo-list-component.md)** - similar as v2 `odo list`, but flags and output needs to be reworked
  - **[`endpoint`](odo-v3-cli/odo-list-endpoint.md)** - similar as v2 `odo url list`, but flags and output needs to be reworked
  - **[`namespace`](odo-v3-cli/odo-list-namespace.md)** -  similar as v2 `odo project list`
  - **`binding`** - new command
  - **`service`** - similar as v2 `odo service list`, but flags and output needs to be reworked
  - **`catalog`** - list all components and services `--type=components,services`  `--filter=java`
- **`odo create`**
  - **`namespace`**
  - **[`endpoint`](odo-v3-cli/odo-create-endpoint.md)** - TODO similar as v2 `odo url create`, but flags and output needs to be reworked
  - **`binding`** - similar as v2 `odo link`, but flags and output needs to be reworked
  - **`service`** - similar as v2 `odo service create`, but flags and output needs to be reworked
- `odo delete`
  - [`component`](odo-v3-cli/odo-delete-component.md) - similar as v2 `odo delete`, but flags and output needs to be reworked
  - **`namespace`**
  - **[`endpoint`](odo-v3-cli/odo-delete-endpoint.md)** - TODO similar as v2 `odo url delete`, but flags and output needs to be reworked
  - **`binding`**  - similar as v2 `odo unlink`, but flags and output needs to be reworked
  - **`service`** - similar as v2 `odo service delete`, but flags and output needs to be reworked
- **`odo set`**
  - **`namespace`** - the same as v2 `odo project set`
- **odo describe**
  - **`component`** - similar as v2 `odo describe`, but flags and output needs to be reworked
  - **`endpoint`** - new command. Shows detailed information about existing endpoint.
  - **`binding`** - new command. Shows detailed information about existing binding
  - **`service`** - new command.  Shows detailed information about existing binding
  - **`catalog`** - `--type=components,services`
- **`odo logs`** - simplifed version of the `odo log` from v2
- `odo version` - as it is in v2
- `odo utils` -  as it is in v2


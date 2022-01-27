# odo v3 CLI

## TODO:

- [ ] define JSON outputs for each command


Over the years of odo development we picked up a lot of commands.
Odo v3 will introduce new commands (`odo dev` #5299, `odo init` #5297).
This will change the command flow from what we currently have in v2. We need to make sure that the whole odo CLI is consistent, and all commands follow the same pattern.

There are also some commands that are there since the original odo v1 and were originally designed for s2i approach only, those commands or flags should be removed, or reworked to better fit Devfile

## Changes from odo v2 command perspective


- `odo component` **removed** every command from “odo component *” already exists as a root command.
- `odo watch` replaced with [`odo dev`](odo-v3-cli/odo-dev.md)
- `odo push`  replaced with [`odo dev`](odo-v3-cli/odo-dev.md)
- `odo unlink`  replaced with [`odo delete binding`](odo-v3-cli/odo-delete-binding.md)
- `odo link`  replaced with [`odo create binding`](odo-v3-cli/odo-create-binding.md)
- `odo url`  replaced with [`odo create endpoint`](odo-v3-cli/odo-create-endpoint.md)
- `odo test`  replaced `odo run test`
- `odo service`  replaced with [`odo create service`](odo-v3-cli/odo-create-service.md)
- `odo env` integrated with [`odo config`](odo-v3-cli/odo-config.md)
- `odo debug` replaced with [`odo run debug`](odo-v3-cli/odo-run-debug.md)
- `odo registry` removed, but the functionality will be added to [`odo preference`](odo-v3-cli/odo-preference.md)
- `odo preference` - mostly as it is, with additional incorporation of  `odo registry`
- `odo login` - as it is
- `odo logout` - as it is
- `odo build-images` - as it is
- `odo deploy` - with new interactive mode
- `odo storage` -  removed, we can consider adding it back, but users can still add storage manuals into `devfile.yaml`
- `odo exec` -  removed. We will need to rework this command later


### general rules for odo cli

- Once even a single flag is provided it ruins in non-interactive mode. All required information needs to be provided via flags
- If command is executed without flags it can enter interactive mode
- every command should have `-o` flag to specify output format, and every command should support `-o json`


## odo v3 CLI structure

- **[`odo login`](odo-v3-cli/odo-login-logout.md)** - no changes required
- **[`odo logout`](odo-v3-cli/odo-login-logout.md)** - no changes required
- **[`odo init`](odo-v3-cli/odo-init.md)** - new command
- **[`odo dev`](odo-v3-cli/odo-dev.md)** - new command based on v2 `odo watch`
- **[`odo registry`](odo-v3-cli/odo-registry.md)** - mostly as it is, just cleanup
- **[`odo preference`](odo-v3-cli/odo-preference.md)** - mostly as it is, just cleanup
- **[`odo config`](odo-v3-cli/odo-config.md)** -  needs to be reworked
- **[`odo build-images`](odo-v3-cli/odo-build-images.md)** - no changes required
- **`odo list`**
  - **[`component`](odo-v3-cli/odo-list-component.md)** - similar as v2 `odo list`, but flags and output needs to be reworked
  - **[`endpoint`](odo-v3-cli/odo-list-endpoint.md)** - similar as v2 `odo url list`, but flags and output needs to be reworked
  - **[`namespace`](odo-v3-cli/odo-list-namespace.md)** -  similar as v2 `odo project list`
  - `binding` - new command
  - `service` - similar as v2 `odo service list`, but flags and output needs to be reworked
- **`odo create`**
  - **[`component`](odo-v3-cli/odo-create-component.md)** - similar as v2 `odo create`, but flags and output needs to be reworked
  - **[`endpoint`](odo-v3-cli/odo-create-endpoint.md)** - similar as v2 `odo url create`, but flags and output needs to be reworked
  - `binding` - similar as v2 `odo link`, but flags and output needs to be reworked
  - `service` - similar as v2 `odo service create`, but flags and output needs to be reworked
- **`odo delete`**
  - **[`component`](odo-v3-cli/odo-delete-component.md)** - similar as v2 `odo delete`, but flags and output needs to be reworked
  - **[`endpoint`](odo-v3-cli/odo-delete-endpoint.md)** - similar as v2 `odo url delete`, but flags and output needs to be reworked
  - `binding`  - similar as v2 `odo unlink`, but flags and output needs to be reworked
  - `service` - similar as v2 `odo service delete`, but flags and output needs to be reworked
- odo describe
  - `component` - similar as v2 `odo describe`, but flags and output needs to be reworked
  - `endpoint` - new command. Shows detailed information about existing endpoint.
  - `binding` - new command. Shows detailed information about existing binding
  - `service` - new command.  Shows detailed information about existing binding
- **`odo catalog`** - command will the same as in v2
  - **`list`**
    - **`components`**
    - **`services`**
  - **`describe`**
    - **`component`**
    - **`service`**
  - **`search`**
    - **`component`**
    - **`service`**
- **`odo version`** - as it is in v2
- **`odo utils`** -  as it is in v2


**bold** commands will be present in v3.0.0-alpha1.
The rest will be added back in following alpha releases.



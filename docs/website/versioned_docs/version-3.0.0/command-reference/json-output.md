---
title: JSON Output
sidebar_position: 20
---

For `odo` to be used as a backend by graphical user interfaces (GUIs),
the useful commands can output their result in JSON format.

When used with the `-o json` flags, a command:
- that terminates successully, will:
  - terminate with a zero exit status,
  - will return its result in JSON format in its standard output stream.
- that terminates with an error, will:
  - terminate with a non-zero exit status,
  - will return an error message in its standard error stream, in the unique field `message` of a JSON object, as in `{ "message": "file not found" }`

## odo alizer -o json

The `alizer` command analyzes the files in the current directory to select the best devfile to use,
from the devfiles in the registries defined in the list of preferred registries with the command `odo preference registry`.

The output of this command contains a devfile name and a registry name:

```bash
$ odo alizer -o json
{
    "devfile": "nodejs",
    "devfileRegistry": "DefaultDevfileRegistry"
}
$ echo $?
0
```

If the command is executed in an empty directory, it will return an error in the standard error stream and terminate with a non-zero exit status:

```bash
$ odo alizer -o json
{
	"message": "No valid devfile found for project in /home/user/my/empty/directory"
}
$ echo $?
1
```

## odo init -o json

The `init` command downloads a devfile and, optionally, a starter project. The usage for this command can be found in the [odo init command reference page](init.md).

The output of this command contains the path of the downloaded devfile and its content, in JSON format.

```bash
$ odo init -o json \
    --name aname \
    --devfile go \
    --starter go-starter
{
	"devfilePath": "/home/user/my-project/devfile.yaml",
	"devfileData": {
		"devfile": {
			"schemaVersion": "2.1.0",
      [...]
		},
		"supportedOdoFeatures": {
			"dev": true,
			"deploy": false,
			"debug": false
		}
	},
	"forwardedPorts": [],
	"runningIn": [],
	"managedBy": "odo"
}
$ echo $?
0
```

If the command fails, it will return an error in the standard error stream and terminate with a non-zero exit status:

```bash
# Executing the same command again will fail
$ odo init -o json \
    --name aname \
    --devfile go \
    --starter go-starter
{
	"message": "a devfile already exists in the current directory"
}
$ echo $?
1
```

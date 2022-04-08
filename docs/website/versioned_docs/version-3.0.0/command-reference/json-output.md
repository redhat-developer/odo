---
title: json output
sidebar_position: 20
---

For `odo` to be used as a backend by graphical user interfaces (GUIs),
the useful commands can output their result in JSON format.

When used with the `-o json` flags, a command:
- that terminates successully, will:
  - terminate with an exit status of value 0,
  - will return its result in JSON format in its standard output stream.
- that terminates with an error, will:
  - terminate with an exit status of value 1,
  - will return an error message in its standard error stream, in the unique field `message` of a JSON object, as in `{ "message": "file not found" }`

## odo alizer -o json

The `alizer` command analyzes the files in the current directory to select the best devfile to use,
from the devfiles in the registries defined in the list of preferred registries with the command `odo preference registry`.

The output of this command contains a devfile name and a registry name:

```
$ odo alizer -o json
{
    "devfile": "nodejs",
    "registry": "DefaultDevfileRegistry"
}
$ echo $?
0
```

If the command is executed in an empty directory, it will return an error and terminate with an exit status of value `1`:

```
$ odo alizer -o json
{
	"message": "No valid devfile found for project in /home/user/my/empty/directory"
}
$ echo $?
1
```

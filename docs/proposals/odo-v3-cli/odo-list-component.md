
# `odo list component`
list devfile components deployed to the cluster in the current namespace.


## Flags
- `--namespace` - list components from the given namespace instead of current namespace.
- `--path` - find and list all components that are in a given path or in its subdirectories.

## example
```
$ odo list
Components in the "mynamspace" namespace:

  NAME            APPLICATION    TYPE         MANAGED BY ODO   RUNNING IN
* frontend        myapp          nodejs       Yes              Dev,Deploy
  backend         myapp          springboot   Yes              Deploy
  created-by-odc  asdf           python       No               Unknown
```

```
$ odo list --path /home/user/my-components/
Components present in the /home/user/my-components/ path

  NAME            APPLICATION    TYPE         MANAGED BY ODO   RUNNING IN  PATH
  frontend        myapp          nodejs       Yes              Dev         frontend
  backend         myapp          springboot   Yes              Deploy      backend
  backend         myapp          springboot   Yes              None        asdf

```

- row/component marked with `*` at the begging of the line is the one that is also in the current directory.
- `TYPE` corresponds to the `langauge` field in `devfile.yaml` tools, this should also correspond to `odo.dev/project-type` label.
- `RUNNING IN` indicates in what modes the component is running. `Dev` means the component is running in development mode (`odo dev`). `Deploy` indicates that the component is running in deploy mode (`odo deploy`), `None` means that the component is currently not running on cluster. `Unknown` indicates that odo can't detect in what mode is component running. `Unknown` will also be used for component that are running on the cluster but are not managed by odo.
- `PATH` column is displayed only if the command was executed with `--path` flag. It shows the path in which the component "lives". This is relative path to a given `--path`.


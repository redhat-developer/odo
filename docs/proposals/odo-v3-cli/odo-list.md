# `odo list`

list all odo related information from the current namespace

## Flags
- `--namespace` - list components from the given namespace instead of current namespace.

## example
```
$ odo list
Components in the "mynamspace" namespace:

  NAME              TYPE         MANAGED BY       RUNNING IN
* frontend          nodejs       odo              Dev,Deploy
  backend           springboot   odo              Deploy
  created-by-odc    python       Unknown          Unknown
```



- row/component marked with `*` at the begging of the line is the one that is also in the current directory.
- `NAMESPACE` indicates in what namespace the component is. This is  mainly intended for when user
- `TYPE` corresponds to the `langauge` field in `devfile.yaml` tools, this should also correspond to `odo.dev/project-type` label.
- `RUNNING IN` indicates in what modes the component is running. `Dev` means the component is running in development mode (`odo dev`). `Deploy` indicates that the component is running in deploy mode (`odo deploy`), `None` means that the component is currently not running on cluster. `Unknown` indicates that odo can't detect in what mode is component running. `Unknown` will also be used for component that are running on the cluster but are not managed by odo.
- `PATH` column is displayed only if the command was executed with `--path` flag. It shows the path in which the component "lives". This is relative path to a given `--path`.





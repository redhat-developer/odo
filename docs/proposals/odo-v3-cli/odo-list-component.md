
# `odo list component`
list devfile components deployed to the cluster in the current namespace.


## Flags
- `--namespace` - list components from the given namespace instead of current namespace.

## example
```
$ odo list components
Components in the "mynamspace" namespace:

  NAME            APPLICATION    TYPE         MANAGED BY       RUNNING IN
* frontend        myapp          nodejs       odo              Dev,Deploy
  backend         myapp          springboot   odo              Deploy
  created-by-odc  asdf           python       Unknown          Unknown
```



- row/component marked with `*` at the beginning of the line is the one that is also in the current directory.
- `TYPE` corresponds to the `langauge` field in `devfile.yaml` tools, this should also correspond to `odo.dev/project-type` label.
- `RUNNING IN` indicates in what modes the component is running. `Dev` means the component is running in development mode (`odo dev`). `Deploy` indicates that the component is running in deploy mode (`odo deploy`), `None` means that the component is currently not running on cluster. `Unknown` indicates that odo can't detect in what mode is component running. `Unknown` will also be used for component that are running on the cluster but are not managed by odo.


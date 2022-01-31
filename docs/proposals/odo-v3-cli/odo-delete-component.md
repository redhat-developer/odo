# `odo delete component`
Delete component.
By default it deletes component from cluster (deletes all resources from both `odo dev` and `odo deploy` that were crated on the cluster that belong to the component) that in the current directory.

Similarly to `odo describe component` this commands works in two modes.
Without `--name` `--application` or `--namespace` it works with the component in the current directory.
If user defines `--name` and `--application` than it works with the component that should exist on the cluster, and it should not touch local component even if the current directory is component directory.
`--namespace` flag is optional, but it makes sense only together with `--name` and `--application`, if not used than it uses the current namespace as defined in `KUBECONFIG`.


## flags
- `-o json` - show output in json format
- `-f` `--force` - force deletion, don't ask for confirmation
- `--name` - name of the existing component on the cluster
- `--application` - name of the application that the component belongs to
- `--namespace` - namespace


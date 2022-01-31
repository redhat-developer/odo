
# `odo describe component`

Show detailed information about component.

By default, it shows detail about component in the current directory.

Alternatively users can use flags (`--name`, `--application`, `--namespace`) to specify existing component on the cluster and show information about it.

To describe component from cluster the `--name` and `--application` flags are always required. `--namespace` is optional and if not specified it will use the current namespace as defined in `KUBECONFIG`.


## Flags

- `-o` (string) output information in a specified format (json).
- `--name` name of the existing component on the cluster
- `--application` name of the application on the clsuter
- `--namespace` name of the cluster namespace.


## example
```
$ odo describe component

Component Name: devfile-nodejs-deploy
Type: nodejs
Environment Variables:
 · PROJECTS_ROOT=/projects
 · PROJECT_SOURCE=/projects
URLs:
 · URL named http-3000 will be exposed via 3000
Linked Services:
 · PostgresCluster/hippo
```


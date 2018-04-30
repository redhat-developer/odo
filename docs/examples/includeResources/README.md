# includeResources

Kedge might not support all the things that Kubernetes has, but kedge would not
come in your way to define anything that Kubernetes understands.

For e.g. right now there is no way to define Kubernetes cron jobs in kedge,
but you can still specify the Kubernetes cron job file. In this field called
`includeResources`.

See snippet from [app.yaml](app.yaml):

```yaml
includeResources:
- cronjob.yaml
```

So in this way you can specify list of files under root level field called
`includeResources`. These files have configuration that Kubernetes understands,
kedge won't do any processing on these files and feed them directly to
Kubernetes.

Also the file paths under `includeResources` should be relative to the kedge file
in which this config is specified.

This does not change anything with respect to deploying applications.

```console
$ kedge apply -f app.yaml
```

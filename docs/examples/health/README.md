# Health

At container level instead of defining `livenessProbe` and 
`readinessProbe` you can define a field called `helath`.
And then that gets replicated in `livenessProbe` and 
`readinessProbe`.

See the snippet below from [web.yaml](web.yaml):

```yaml
containers:
...
  health:
    httpGet:
      path: /
      port: 80
    initialDelaySeconds: 20
    timeoutSeconds: 5
...
```

When this is expanded the same content is replicated in both 
fields:

```yaml
$ kedge generate -f web.yaml
...
        livenessProbe:
          httpGet:
            path: /
            port: 80
          initialDelaySeconds: 20
          timeoutSeconds: 5
        readinessProbe:
          httpGet:
            path: /
            port: 80
          initialDelaySeconds: 20
          timeoutSeconds: 5
...
```

But if `health` is defined with `livenessProbe` or `readinessProbe`
the tool will error out, so define only one.

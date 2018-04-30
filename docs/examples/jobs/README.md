# Job

Now you can add support for Kubernetes controller type Job. All you need to do is provide a
root level field called `controller` like this:

```yaml
name: pival
controller: job
containers:
...
``` 

All of the information you provide for a normal controller will be same. By default the value
of the `restartPolicy` will be `OnFailure`.


At root level the fields that are available are from `PodSpec` and `JobSpec`. For example of
job definition in kedge look at file [`job.yaml`](job.yaml).
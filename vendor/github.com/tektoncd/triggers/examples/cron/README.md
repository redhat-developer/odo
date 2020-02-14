# Cron Triggers

The following example uses a Kubernetes
[CronJob](https://kubernetes.io/docs/concepts/workloads/controllers/cron-jobs/)
to implement a basic cron trigger that runs every minute.

This works by using a cron job that emits a HTTP request to the EventListener
Service endpoint.

To create the cron trigger and all related resources, run:

```
kubectl apply -f .
```

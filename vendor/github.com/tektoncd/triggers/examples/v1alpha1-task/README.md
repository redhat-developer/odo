## v1alpha1 Task EventListener

Creates an EventListener that creates a v1alpha1 TaskRun.

### Try it out locally:

1. Create the service account:

   ```shell script
   kubectl apply -f examples/role-resources/triggerbinding-roles
   kubectl apply -f examples/role-resources/
   ```

1. Create the v1alpha1 EventListener:

   ```shell script
   kubectl apply -f examples/v1alpha1-task/v1alpha1-task-listener.yaml
   ```

1. Port forward:

   ```shell script
   kubectl port-forward \
    "$(kubectl get pod --selector=eventlistener=v1alpha1-task-listener -oname)" \
     8080
   ```

   **Note**: Instead of port forwarding, you can set the
   [`serviceType`](https://github.com/tektoncd/triggers/blob/master/docs/eventlisteners.md#serviceType)
   to `LoadBalancer` to expose the EventListener with a public IP.

1. Test by sending the sample payload.

   ```shell script
   curl -v \
   -H 'Content-Type: application/json' \
   --data "{}" \
   http://localhost:8080
   ```

   The response status code should be `201 Created`

1. You should see a new TaskRun that got created:

   ```shell script
   kubectl get taskruns | grep v1alpha1-task-run-
   ```

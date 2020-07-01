## GitLab Push EventListener

Creates an EventListener that listens for Gitlab webhook events.

### Try it out locally:

1. Create the service account:

   ```shell script
   kubectl apply -f examples/role-resources/triggerbinding-roles
   kubectl apply -f examples/role-resources/
   ```

1. Create the Gitlab EventListener:

   ```shell script
   kubectl apply -f examples/gitlab/gitlab-push-listener.yaml
   ```

1. Port forward:

   ```shell script
   kubectl port-forward \
    "$(kubectl get pod --selector=eventlistener=gitlab-listener -oname)" \
     8080
   ```

   **Note**: Instead of port forwarding, you can set the
   [`serviceType`](https://github.com/tektoncd/triggers/blob/master/docs/eventlisteners.md#serviceType)
   to `LoadBalancer` to expose the EventListener with a public IP.

1. Test by sending the sample payload.

   ```shell script
   curl -v \
   -H 'X-GitLab-Token: abcde' \
   -H 'X-Gitlab-Event: Push Hook' \
   -H 'Content-Type: application/json' \
   --data-binary "@examples/gitlab/gitlab-push-event.json" \
   http://localhost:8080
   ```

   The response status code should be `201 Created`

1. You should see a new TaskRun that got created:

   ```shell script
   kubectl get taskruns | grep gitlab-run-
   ```

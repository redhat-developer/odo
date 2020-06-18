## Bitbucket EventListener

Creates an EventListener that listens for Bitbucket webhook events.

### Try it out locally:

1. To create the Bitbucket trigger and all related resources, run:

   ```bash
   kubectl apply -f examples/bitbucket/
   ```

1. Port forward:

   ```bash
   kubectl port-forward \
    "$(kubectl get pod --selector=eventlistener=bitbucket-listener -oname)" \
     8080
   ```

   **Note**: Instead of port forwarding, you can set the
   [`serviceType`](https://github.com/tektoncd/triggers/blob/master/docs/eventlisteners.md#serviceType)
   to `LoadBalancer` to expose the EventListener with a public IP.

1. Test by sending the sample payload.

   ```bash
   curl -v \
   -H 'X-Event-Key: repo:refs_changed' \
   -H 'X-Hub-Signature: sha1=c401e56873be567c1403d8504763760a39236c51' \
   -d '{"repository": {"links": {"clone": [{"href": "ssh://git@localhost:7999/~test/helloworld.git", "name": "ssh"}, {"href": "http://localhost:7990/scm/~test/helloworld.git", "name": "http"}]}}, "changes": [{"ref": {"displayId": "master"}}]}' \
   http://localhost:8080
   ```

   The response status code should be `201 Created`
   
   [`HMAC`](https://www.freeformatter.com/hmac-generator.html) tool used to create X-Hub-Signature.
   
   In [`HMAC`](https://www.freeformatter.com/hmac-generator.html) `string` is the *body payload* and `secretKey` is the *given secretToken*.

1. You should see a new TaskRun that got created:

   ```bash
   kubectl get taskruns | grep bitbucket-run-
   ```

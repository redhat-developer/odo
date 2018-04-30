# Build container images

### Build image

```console
$ kedge build -i surajd/ticker:0.1 -p
INFO[0000] Building image 'surajd/ticker:0.1' from directory 'build' 
INFO[0000] Image 'surajd/ticker:0.1' from directory 'build' built successfully 
INFO[0000] Pushing image "surajd/ticker:0.1" to registry "docker.io" 
INFO[0000] Multiple authentication credentials detected. Will try each configuration. 
INFO[0000] Attempting authentication credentials for "172.30.1.1:5000" 
ERRO[0003] Unable to push image "surajd/ticker:0.1" to registry "172.30.1.1:5000". Error: unauthorized: incorrect username or password 
INFO[0003] Attempting authentication credentials for "https://index.docker.io/v1/" 
INFO[0057] Successfully pushed image "surajd/ticker:0.1" to registry "https://index.docker.io/v1/" 
```

In above example flag `-p` specifies that push image after build is complete.

**Note**:

* Above image name is `surajd/ticker:0.1`, you can put your docker hub username.
* If you are using Kubernetes in minikube run `eval $(minikube docker-env)`, before
running build command.

### Deploying application

Deploy the kedge configs in `configs` directory:

```console
$ kedge apply -f configs/
service "ticker" created
deployment "ticker" created
service "redis" created
deployment "redis" created
```

Verify that the application is running fine:

```console
$ curl `minikube ip`:30771
<h3>Hello Kubernauts</h3> <br/><h2>Number of Hits:</2> 1<br/>
```


# Getting started with Triggers

To get started with Triggers, let's put it to work building and deploying a real
image. In the following guide, we will use `Triggers` to handle a real GitHub
webhook request to kickoff a PipelineRun.

## Install dependencies

Before we can use the Triggers project, we need to get some dependencies out of
the way.

- [A Kubernetes Cluster](https://kubernetes.io/docs/setup/)
  - This guide depends on an having access to a Kubernetes cluster which is
    publicly reachable from the internet.
  - The cluster also needs the ability to
    [create ingress resources](https://kubernetes.io/docs/concepts/services-networking/ingress/).
  - Most cloud providers k8s offerings work for this purpose...
    - but ingress does not work out of the box for GKE clusters.
    - For now, GKE users should consider using the
      [nginx ingress](https://kubernetes.github.io/ingress-nginx/deploy/#gce-gke).
- [Install Tekton Pipelines](https://github.com/tektoncd/pipeline/blob/master/docs/install.md#installing-tekton-pipelines)
  - Pipelines is the backbone of Tekton and will allow us to accomplish the work
    we plan to do.
- [Install Triggers](../install.md)
  - Of course we need to install our project as well, so we can accept and
    process events into PipelineRuns!
- Pick a GitHub repo with a Dockerfile as your build object (or you can fork
  [this one](https://github.com/iancoffey/ulmaceae)).
  - Clone this repo locally - we will come back to this repo later.

## Configure the cluster

Now that we have our cluster ready, we need to setup our getting-started
namespace and RBAC. We will keep everything inside this single namespace for
easy cleanup. In the unlikely event that you get stuck/flummoxed, the best
course of action might be to just delete this namespace and start fresh.

- Create the _getting-started_ namespace, where all our resources will live.
  - `kubectl create namespace getting-started`
- [Create the admin user, role and rolebinding](./rbac/admin-role.yaml)
  - `kubectl apply -f ./docs/getting-started/rbac/admin-role.yaml`
- [Create the create-webhook user, role and rolebinding](./rbac/webhook-role.yaml)
  - `kubectl apply -f ./docs/getting-started/rbac/webhook-role.yaml`
  - This will allow our webhook to create the things it needs to.

## Install the Pipeline and Triggers

### [Install the Pipeline](./pipeline.yaml)

Now we have to install the Pipeline we plan to use and also our Triggers
resources.

`kubectl apply -f ./docs/getting-started/pipeline.yaml`

Our Pipeline will do a few things.

- Retrieve the source code
- Build and push the source code into a Docker image
- Push the image to the specified repository
- Run the image locally

#### What does it do?

The Pipeline will build a Docker image with
[img](https://github.com/genuinetools/img) and deploy it locally via kubectl
image.

### [Install the TriggerTemplate, TriggerBinding and EventListener](./triggers.yaml)

The Triggers project will pickup from there.

- We will setup an `EventListener` to accept and process GitHub Push events
- A `TriggerTemplate` to create templated PipelineResource and PipelineRun
  resources per event received by the `EventListener`.
  - First, **update** the `triggers.yaml` file to reflect the Docker repository
    you wish to push the image blob to.
  - You will need to replace the `DOCKERREPO-REPLACEME` string everywhere it is
    needed.
  - Once you have updated the triggers file, you can apply it!
    - `kubectl apply -f ./docs/getting-started/triggers.yaml`
  - If that succeeded, your cluster is ready to start handling Events.

## Add Ingress and GitHub-Webhook Tasks

We will need an ingress to handle incoming webhooks and we will make use of our
new ingress by configuring GitHub with our GitHub Task.

First lets create our ingress Task.

`kubectl apply -f ./docs/create-ingress.yaml -n getting-started`

Now lets create our webhook Task.

`kubectl apply -f ./docs/create-webhook.yaml -n getting-started`

## Run Ingress Task

### Update the Ingress TaskRun

**Note**: If you are running on GKE, the default Ingress will not work. Instead,
follow the instructions to use an Nginx Ingress
[here](../exposing-eventlisteners.md#Using-Nginx-Ingress)

Lets first update the TaskRun to make any needed changes

Edit the `docs/getting-started/ingress-run.yaml` file to adjust the settings.

At the mimimum, you will need to update the ExternalDomain field to match your
own DNS name.

### Run the Ingress Task

When you are ready, run the ingress Task.

`kubectl apply -f docs/getting-started/ingress-run.yaml`

## Run GitHub Webhook Task

You will need to create a
[GitHub Personal Access Token](https://help.github.com/en/articles/creating-a-personal-access-token-for-the-command-line#creating-a-token)
with the following access.

- `public_repo`
- `admin:repo_hook`

Next, create a secret like so with your access token.

```
apiVersion: v1
kind: Secret
metadata:
  name: webhook-secret
  namespace: getting-started
stringData:
  token: YOUR-GITHUB-ACCESS-TOKEN
  secret: random-string-data
```

### Update webhook task run

Now lets update the GitHub Task run.

There are a few fields to change, but these fields must be updated at the
minimum.

- GitHubOrg: The GitHub org you are using for this getting-started.
- GitHubUser: Your GitHub username.
- GitHubRepo: The repo we will be using for this example.

### Run the Webhook Task

Now lets run our updated webhook task.

`kubectl apply -f docs/getting-started/webhook-run.yaml`

## Watch it work!

- Commit and push an empty commit to your development repo.
  - `git commit -a -m "build commit" --allow-empty && git push origin mybranch`
- Now, you can follow the Task output in `kubectl logs`.
  - First the image builder task.
    - `kubectl logs -l somelabel=somekey --all-containers`
  - Then our deployer task.
    - `kubectl logs -l tekton.dev/pipeline=getting-started-pipeline -n getting-started --all-containers`
- We can see now that our CI system is working! Images pushed to this repo
  result in a running pod in our cluster.
  - We can examine our pod like so.
    - kubectl logs tekton-triggers-built-me -n getting-started --all-containers

Now we can see our new image running our cluster, after having been retrieved,
tested, vetted and built, docker pushed (and pulled) and finally ran on our
cluster as a Pod.

## Clean up

- Delete the _getting-started_ namespace!
  - `kubectl delete namespace getting-started`

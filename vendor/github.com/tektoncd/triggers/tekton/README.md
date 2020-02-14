# Tekton Repo CI/CD

_Why does Tekton triggers have a folder called `tekton`? Cuz we think it would
be cool if the `tekton` folder were the place to look for CI/CD logic in most
repos!_

We use Tekton Pipelines to build, test and release Tekton Triggers!

This directory contains the
[`Tasks`](https://github.com/tektoncd/pipeline/blob/master/docs/tasks.md) and
[`Pipelines`](https://github.com/tektoncd/pipeline/blob/master/docs/pipelines.md)
that we use.

The Pipelines and Tasks in this folder are used for:

1. [Manually creating official releases from the official cluster](#create-an-official-release)

To start from scratch and use these Pipelines and Tasks:

1. [Install Tekton v0.3.1](https://github.com/tektoncd/pipeline/blob/master/tekton/README.md#install-tekton)
1. [Setup the Tasks and Pipelines](https://github.com/tektoncd/pipeline/blob/master/tekton/README.md#setup)
1. [Create the required service account + secrets](https://github.com/tektoncd/pipeline/blob/master/tekton/README.md#service-account-and-secrets)

## Create an official release

Official releases are performed from
[the `prow` cluster](https://github.com/tektoncd/plumbing#prow)
[in the `tekton-releases` GCP project](https://github.com/tektoncd/plumbing/blob/master/gcp.md).
This cluster
[already has the correct version of Tekton installed](#install-tekton).

To make a new release:

1. (Optionally) [Apply the latest versions of the Tasks + Pipelines](#setup)
2. (If you haven't already)
   [Install `tkn`](https://github.com/tektoncd/cli#installing-tkn)
3. [Run the Pipeline](#run-the-pipeline)
4. Create the new tag and release in GitHub
   ([see one of way of doing that here](https://github.com/tektoncd/pipeline/issues/530#issuecomment-477409459)).
   _TODO(tektoncd/pipeline#530): Automate as much of this as possible with
   Tekton._
5. Add an entry to [the README](../README.md) at `HEAD` for docs and examples
   for the new release ([README.md#read-the-docs](README.md#read-the-docs)).
6. Update the new release in GitHub with the same links to the docs and
   examples, see
   [v0.1.0](https://github.com/tektoncd/pipeline/releases/tag/v0.1.0) for
   example.

### Run the Pipeline

To use [`tkn`](https://github.com/tektoncd/cli) to run the
`publish-tekton-pipelines` `Task` and create a release:

1. Pick the revision you want to release and update the
   [`resources.yaml`](./resources.yaml) file to add a `PipelineResoruce` for it,
   e.g.:

   ```yaml
   apiVersion: tekton.dev/v1alpha1
   kind: PipelineResource
   metadata:
     name: tekton-triggers-vX-Y-
   spec:
     type: git
     params:
       - name: url
         value: https://github.com/tektoncd/triggers
       - name: revision
         value: vX.Y.Z-invalid-tags-boouuhhh # REPLACE with the commit you'd like to build from
   ```

1. To run against your own infrastructure (if you are running
   [in the production cluster](https://github.com/tektoncd/plumbing#prow) the
   default account should already have these creds, this is just a bonus - plus
   `release-right-meow` might already exist in the cluster!), also setup the
   required credentials for the `release-right-meow` service account, either:

   - For
     [the GCP service account `release-right-meow@tekton-releases.iam.gserviceaccount.com`](#production-service-account)
     which has the proper authorization to release the images and yamls in
     [our `tekton-releases` GCP project](https://github.com/tektoncd/plumbing#prow)
   - For
     [your own GCP service account](https://cloud.google.com/iam/docs/creating-managing-service-accounts)
     if running against your own infrastructure

1. [Connect to the production cluster](https://github.com/tektoncd/plumbing#prow):

   ```bash
   gcloud container clusters get-credentials prow --zone us-central1-a --project tekton-releases
   ```

1. Run the `release-pipeline` (assuming you are using the production cluster and
   [all the Tasks and Pipelines already exist](#setup)):

   ```shell
   # Create the resoruces - i.e. set the revision that you wan to build from
   kubectl apply -f tekton/resources.yaml

   # Change thie environment variable to the verison you would like to use.
   # Be careful: due to #983 it is possible to overwrite previous releases.
   export VERSION_TAG=v0.X.Y

   tkn pipeline start \
   \
   \
   \
   \
   \
    --resource=builtControllerImage=triggers-controller-image \
   \
   e
   ```

_TODO(tektoncd/pipeline#569): Normally we'd use the image `PipelineResources` to
control which image registry the images are pushed to. However since we have so
many images, all going to the same registry, we are cheating and using a
parameter for the image registry instead._

## Supporting scripts and images

Some supporting scripts have been written using Python 2.7:

- [koparse](./koparse) - Contains logic for parsing `release.yaml` files created
  by `ko`

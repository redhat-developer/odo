# Binding an Imported app to a Route/Ingress

## Introduction

Binding information can be present in standalone k8s objects like routes/ingress, services, deployments too. This scenario illustrates using any resource ( CR / non-CR ) which has a spec and a status as a backing service.

Binding metadata is being read from annotations on the backing service ( like CR, Route, Service, basically any kubernetes object with a spec and status, along with associated CRD or CSV.

Here's how the operator resolves the binding metadata:

1) Look up annotations in the CR or kubernetes resource,
2) Look up annotations in CRD
3) Look up descriptors in CSV ( overrides the CRD annotations ..)
Provide cumulative annotations : (1) and (2 & 3).


## Actions to Perform by Users in 2 Roles

In this example there are 2 roles:

* Cluster Admin - Installs the operator into the cluster
* Application Developer - Imports Node.js applications, creates a Route

### Cluster Admin

First, let's be the cluster admin. We need to install the service binding operator in the cluster:

Navigate to `Operators`->`OperatorHub` in the OpenShift console; in the `Developer Tools` category select `Service Binding Operator`

![Service Binding Operator as shown in OperatorHub](../../assets/operator-hub-sbo-screenshot.png)

and install an `alpha` version.

This makes the `ServiceBindingRequest` custom resource available for the application developer.


### Application Developer

Now, let's play the role of an application developer. The application needs a namespace to live in so let's create one:

``` shell
cat <<EOS |kubectl apply -f -
---
kind: Namespace
apiVersion: v1
metadata:
  name: service-binding-demo
EOS
```

#### Import an application

In this example we will import an arbitrary [Node.js application](https://github.com/pmacik/nodejs-rest-http-crud).

In the OpenShift Console switch to the Developer perspective. (Make sure you have selected the `service-binding-demo` project). Navigate to the `+ADD` page from the menu and then click on the `[Import from Git]` button. Fill in the form with the following:

* `Git Repo URL` = `https://github.com/pmacik/nodejs-rest-http-crud`
* `Project` = `service-binding-demo`
* `Application`->`Create New Application` = `NodeJSApp`
* `Name` = `nodejs-rest-http-crud-git`
* `Builder Image` = `Node.js`
* `Create a route to the application` = checked

and click on the `[Create]` button.

#### Create a Route and annotate it:

Now let's create a kubernetes resource - `Route` (for our case) and annotate it with the value that we would like to be injected for binding. For this case it is the spec.host

``` shell
cat <<EOS |kubectl apply -f -
---
kind: Route
apiVersion: route.openshift.io/v1
metadata:
  name: example
  namespace: service-binding-demo
  annotations:
    openshift.io/host.generated: 'true'
    servicebindingoperator.redhat.io/spec.host: 'binding:env:attribute' #annotate here.
spec:
  host: example-sbo.apps.ci-ln-smyggvb-d5d6b.origin-ci-int-aws.dev.rhcloud.com
  path: /
  to:
    kind: Service
    name: example
    weight: 100
  port:
    targetPort: 80
  wildcardPolicy: None
EOS
```

Now create a ServiceBindingRequest as below:

``` shell
cat <<EOS |kubectl apply -f -
---
apiVersion: apps.openshift.io/v1alpha1
kind: ServiceBindingRequest

metadata: 
  name: binding-request
  namespace: service-binding-demo

spec: 
  applicationSelector: 
    group: ""
    resource: deployments
    resourceRef: nodejs-rest-http-crud-git
    version: v1

  backingServiceSelectors:                                                                                                        
   - group: route.openshift.io
      version: v1
      kind: Route # <--- not NECESSARILY a CR
      resourceRef: example 
EOS
```

When the `ServiceBindingRequest` was created the Service Binding Operator's controller injected the Route information that was annotated to be injected into the application's `Deployment` as environment variables via an intermediate `Secret` called `binding-request`.

Check the contents of `Secret` - `binding-request` by executing `oc get secrets binding-request -o yaml` for the following result:

`ROUTE_HOST: example-sbo.apps.ci-ln-smyggvb-d5d6b.origin-ci-int-aws.dev.rhcloud.com`


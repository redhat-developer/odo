# odo link to helm chart

## Motivation
Helm is probably the most widely used tool to deploy standardized services (databases, caches, proxies etc..) on Kubernets.
Odo users should be able easily consume services deployed using Helm.


## User stories
### User Story 1
As a developer I want to be able to create a link (`odo link <helm-chart>`) between my component and a database that I’ve deployed using Helm so I can connect to a database from my application.

### User Story 2
As a developer consuming services deployed using Helm I want to be able to list what I can consume (link to).

## Design overview
###  Changes in `odo link` 
To cover [User Story 1](#user-story-1) `odo link` command needs to be able to generate ServiceBinding pointing to Service that belongs to deployed Helm Chart.




```yaml
apiVersion: operators.coreos.com/v1alpha1
kind: ServiceBinding
metadata:
  name: example-servicebinding
  namespace: tkral-test
spec:
  application:
    group: apps
    name: python-mysql
    resource: deployments
    version: v1
  services:
    - group: apps
      name: my-mysql
      version: v1
      kind: Deployments
```


#### Challenges/Blockers:
##### What should be target (`sepc.services`) in `ServiceBinding` CR?
##### When I used `Deployment` in `spec.services` SBO did not detected any binding information (`Secret` was empty).
This issue might be related to this https://github.com/redhat-developer/service-binding-operator/issues/722
##### How to make  SBO detect correct binding information?




### Changes in  `odo service`
To cover [User Story 2](user-story-2) `odo service list` command  needs to list all deployed Helm Charts.
The Name field in the list output should uniquely identify service (Helm Chart) so users can use that name in `odo link` command.
```bash
odo link <service-name>
```

One approach could be to list `Service`s defined in Helm charts
```bash
kubectl get svc -l app.kubernetes.io/managed-by=Helm -o jsonpath="Service.name | Helm.release-name{'\n'}{range .items[*]}{.metadata.name}  |  {.metadata.annotations.meta\.helm\.sh\/release-name}{'\n'}{end}"
```

```
Service.name | Helm.release-name
mongodb  |  mongodb
my-mysql  |  my-mysql
my-mysql-slave  |  my-mysql
```

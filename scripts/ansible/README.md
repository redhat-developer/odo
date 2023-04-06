# OVERVIEW
This directory contains ansible manifest files that create the complete infrastructure on IBMCloud.

This script is used for automation and are using in two github-actions.
- to create a staging environment to test if the changes are working as expected
- to will apply new changes to the production infrastructure

## ***NOTE***

>Deleation of stagin environment is done manually, as the github-action will only create the staging env, testing the staging env is manual.

### __How to create complete infra?__
> NOTE: you will need to export path to ssh_key for login pourpose (`SSHKEY` is variable name)
Run the following commands
``` shell
# export ssh_key path
export SSHKEY=/path/to/ssh/key
# expose the Region and API key for ansible script
export IC_REGION="eu-de"     
export IC_API_KEY="<API_KEY>"

ansible-playbook create-infra.yaml
```


### __How to delete complete infra?__

Run the following commands
``` shell
# expose the Region and API key for ansible script
export IC_REGION="eu-de"     
export IC_API_KEY="<API_KEY>"

ansible-playbook delete-infra.yaml
```

### Manual Steps to setup cluster 
Manual steps need to be done on each cluster
#### 1. [install operators](./Cluster/kubernetes-cluster/README.md#kubernetes-operators) 

> ___NOTE: if want to use nfs for storage___
#### 2. [adding nfs server to cluster](./Cluster/kubernetes-cluster/README.md#nfs-provisioner)

>___NOTE: only do this step if configuring nfs on cluster___
#### 3. [remove storage](./Cluster/README.md#to-remove-storage-addon-from-cluster)
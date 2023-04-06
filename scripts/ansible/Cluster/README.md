# OVERVIEW
This directory contains ansible files that create/destroy clusters on IBMCloud.

This script is used for automation and are using in two github-actions.
- to create a staging environment to test if the changes are working as expected
- that will apply those new changes to the actual infrastructure

## Pre-requisite
- ansible
- `pip3 install openshift`
- `ansible-galaxy collection install -r requirements.yml`

## ***NOTE***

>Deleation of stagin environment is done manually, as the github-action will only create the staging env, testing the staging env is manual.
>By default the scripts are configured for staging environment( to make sure infra is not modified by mistake )

## ___How to check your changes?___

Create a PR with

- changes in README File Present in ansible: 
  - first commit should only have changes from Readme.md file. so that it creates the cluster similar to main infra
- changes in yaml files 
  - later commits will have the changes that you want to test on staging environment

### __How to delete staging environment?__

Run the following commands
``` shell
# expose the Region and API key for ansible script
export IC_REGION="eu-de"     
export IC_API_KEY="<API_KEY>"

ansible-playbook delete-clusters.yaml
```
### To remove storage addon from cluster
```shell
ibmcloud ks cluster addon disable vpc-block-csi-driver -c <cluster-ID>
```
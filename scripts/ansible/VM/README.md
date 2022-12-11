This Readme contains info on how to setup VSI on ibmcloud to test odo with podman

This Directory contains files required to create the infra for testing odo with podman

### Files and there usages
- create.yml - ansible script to create vsi and other resources required for VSI on IBMCLOUD
- delete-only-vsi.yml  - ansible script to delete only `vsi` and `floatingIP` (public IP) and leave VPC,subnet, etc as it is for test automation.  
- destroy.yml - ansible script to delete all the resources created by `create.yml` ansible script
- image.sh - a shell script to create/delete an image  
- install_dependency.sh  - a shell script that is used in `create.yml` for seting up the VSI with podman and other pre-requiest
- list_vsi_images_and_profiles.yml - used in `create.yml`
- MORE_README.md - Detailed info about the test infra and how its done with steps
- vars.yml - setup variables for testing


### PreRequisite
```
ansible-galaxy collection install ibm.cloudcollection
```

### STEPS to create infra:

1. export IBMCLOUD key and IBMCLOURD ZONE
```
export IC_API_KEY=<KEY_VALUE>
export IC_REGION="eu-de"
```

2. Run `create.yml` to create vsi
```
ansible-playbook create.yml
```

3. Create an Image from the vsi created in previous step
```
./image.sh create
```
```
NOTE: ./image.sh create will display name of the image created that need to passed in IBMCLOUD pipeline as a variable
eg: 
./image.sh create
use this value "riding-stratus-embellish-snippet-rename" in ibmcloud test as value of VSI IMAGE for image creation refference 
Here: riding-stratus-embellish-snippet-rename is the name of image
``` 

5. Delete VSI and assigned floating IP
```
ansible-playbook delete-only-vsi.yml
```


#### For Cleanup

Steps:

1. delete all resources related to VSI
```
ansible-playbook destroy.yml
```
2. Remove the image created
```
export IMAGE=riding-stratus-embellish-snippet-rename
./image.sh delete
```
#! /usr/bin/sh

# set -e, to exit if any of the commands fails
set -e

##################################################
# pre requiests for this script
# 1. need ibmcloud cli
# 2. need to be logged in ibmcloud env
# 3. While using delete passing image name with env var `IMAGE`

# VARIABLES FOR SCRIPT
VSI_NAME=${VSI_NAME:-"odo-test-automation-vsi"}
IMAGE=${IMAGE:-""}

# create vsi image in ibmcloud  
create_image() {
    # 2
    ibmcloud is instance-stop --force $VSI_NAME

    # 3
    VSI_BASE=$(ibmcloud is instances odo-test-automation-vsi --output json | jq -r '.[0].boot_volume_attachment.volume.id')

    # 4
    IMAGE=$(ibmcloud is image-create --source-volume $VSI_BASE --output json | jq '.name')

    echo "use this value $IMAGE in ibmcloud test as value of VSI IMAGE for image creation refference"
}

# For cleanup
# This should be used when we are creating a new image for the tests
delete_image() {
    if [ "$IMAGE" == "" ]; then
        echo -e "please specify IMAGE value \neg: export IMAGE=riding-stratus-embellish-snippet-rename"
        exit 1
    fi
    ibmcloud is image-delete --force $IMAGE
}

if [ "$1" == "create" ]; then
    create_image
elif [ "$1" == "delete" ]; then
    delete_image
else
    echo "please pass, either 'create' or 'delete' keywords"
fi

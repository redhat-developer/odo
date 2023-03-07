# ReadMe 
This directory contains yaml files to create NFS server 

### NFS provisioner (how to configure nfs for cluster)

You can run the following commands upon a cluster to deploy the NFS provisioner to this cluster (either Kubernetes or OpenShift). You will need to uninstall the "Block Storage for VPC" add-on installed by default, to make the NFS provisioner work correctly.

```
$ helm repo add nfs-subdir-external-provisioner \
    https://kubernetes-sigs.github.io/nfs-subdir-external-provisioner/

$ helm install nfs-subdir-external-provisioner \
    nfs-subdir-external-provisioner/nfs-subdir-external-provisioner \
    --set nfs.server=<IP_FOR_NFS> \
    --set nfs.path=/mnt/nfs \
    --set storageClass.defaultClass=true \
    --set storageClass.onDelete=delete
    --version=4.0.15
```

> learn more about nfs-subdir-external-provisioner from https://artifacthub.io/packages/helm/nfs-subdir-external-provisioner/nfs-subdir-external-provisioner

### check if nfs is working or not

login using the floating IP

### **NOTE**

ibmcoud storage provided with cluster doesnt works with nfs storge(if nfs storage is set as default). So make sure to diable addon `vpc-block-csi-driver` from cluster for which you want to use **nfs-storage**

#### *command to delete/remove storage addons from cluster*

```shell
ibmcloud ks cluster addon disable vpc-block-csi-driver
```

### helpful commands

1. Fetch IP for nfs configuration 
```shell
IP_FOR_NFS=$(ibmcloud is instance <nfs-instance-name> --output json | jq -r ".primary_network_interface.primary_ip.address")
```

2. Fetch Floating IP of NFS-Server
```shell
NFS_IP=$(ibmcloud is instance k8s-nfs-server --output json | jq -r ".primary_network_interface.floating_ips[0].address" )
```

3. Create/Delete just NFS server
> NOTE: you will need to export path to ssh_key for login pourpose (`SSHKEY` is variable name)
```
$ export SSHKEY=/path/to/ssh/key

$ ansible-playbook create.yaml \
    -e name_prefix=odo-tests \
    -e cluster_zone="eu-de-2"

$ ansible-playbook delete.yaml \
    -e name_prefix=odo-tests
```
# Ansible Playbooks for odo testing

## IBM Cloud Kubernetes Cluster

This ansible playbook deploys a VPC Kubernetes/OpenShift cluster on IBM Cloud and an NFS server on the same VPC (to be used for dynamic storage provisioning - deploying the NFS provisioner is required, see below).
It uses the [IBM Cloud Ansible Collections](https://github.com/IBM-Cloud/ansible-collection-ibm/).

### VPC Resources

The following VPC infrastructure resources will be created (Ansible modules in
parentheses):

* Resource group (ibm_resource_group)
* VPC (ibm_is_vpc)
* Security Group (ibm_is_security_group_rule)
* Public gateway (ibm_is_public_gateway)
* VPC Subnet (ibm_is_subnet)
* SSH Key (ibm_is_ssh_key)
* Virtual Server Instance (ibm_is_instance)
* Floating IP (ibm_is_floating_ip)
* Cloud Object Storage (ibm_resource_instance)
* VPC Kubernetes Cluster (ibm_container_vpc_cluster)

All created resources (expect resource group and SSH Key) will be inside the created Resource Group.

Note that:
- ibm_is_security_group_rule is not idempotent: each time the playbook is ran, an entry in the Inbound Rules of the Security Group allowing port 22 will be added. You should remove the duplicates from the UI and keep only one entry.
- I (feloy) didn't find a way to uninstall an addon from a cluster using the IBM Cloud ansible collection (https://github.com/IBM-Cloud/ansible-collection-ibm/issues/70). You will need to remove the "Block Storage for VPC" default add-on if you install an NFS provisioner for this cluster.


### Configuration Parameters

The following parameters can be set by the user, either by editing the `vars.yaml` or by usning the `-e` flag from the `ansible-galaxy` command line:

* `name_prefix`: Prefix used to name created resources
* `cluster_zone`: Zone on which will be deployed the resources
* `total_ipv4_address_count`: Number of IPv4 addresses available in the VPC subnet
* `ssh_public_key`: SSH Public key deployed on the NFS server
* `nfs_image`: The name of the image used to deploy the NFS server
* `kube_version`: Kubernetes or OpenShift version. The list of versions can be obtained with `ibmcloud ks versions`
* `node_flavor`: Flavor of workers of the cluster. The list of flavors can be obtained with `ibmcloud ks flavors --zone ${CLUSTER_ZONE} --provider vpc-gen2`
* `workers`: Number of workers on the cluster
* `cluster_id_file`: File on which the cluster ID will be saved
* `nfs_ip_file`: File on which the private IP of the NFS server will be saved

### Running

#### Set API Key and Region

1. [Obtain an IBM Cloud API key](https://cloud.ibm.com/docs/account?topic=account-userapikey&interface=ui).

2. Export your API key to the `IC_API_KEY` environment variable:

    ```
    export IC_API_KEY=<YOUR_API_KEY_HERE>
    ```

3. Export desired IBM Cloud region to the 'IC_REGION' environment variable:

    ```
    export IC_REGION=<REGION_NAME_HERE>
    ```

You can get available regions supporting Kubernetes clusters on the page https://cloud.ibm.com/docs/containers?topic=containers-regions-and-zones.

#### Install Ansible collections

To install the required Ansible collections, run:

```
ansible-galaxy collection install -r requirements.yml
```

#### Create

To create all resources, run the 'create' playbook:

For example:

```
$ ansible-playbook create.yml \
    -e name_prefix=odo-tests-openshift \
    -e kube_version=4.7_openshift \
    -e cluster_id_file=/tmp/openshift_id \
    -e nfs_ip_file=/tmp/nfs_openshift_ip \
    --key-file <path_to_private_key> # For an OpenShift cluster v4.7

$ ansible-playbook create.yml \
    -e name_prefix=odo-tests-kubernetes \
    -e kube_version=1.20 \
    -e cluster_id_file=/tmp/kubernetes_id \
    -e nfs_ip_file=/tmp/nfs_kubernetes_ip \
    --key-file <path_to_private_key> # For a Kubernetes cluster v1.20
```

The `path_to_private_key` file contains ths SSH private key associated with the SSH public key set in the `ssh_public_key` configuration parameter.

#### Destroy

To destroy all resources run the 'destroy' playbook:

```
ansible-playbook destroy.yml -e name_prefix=...
```

## Kubernetes Operators

This ansible playbook deploys operators on a Kubernetes cluster. The cluster should be running the Operator Lifecycle Manager ([OLM](https://olm.operatorframework.io/)), either natively for an OpenShift cluster, or by installing it on a Kubernetes cluster.

To install OLM on a Kubernetes cluster go to the ([OLM releases page](https://github.com/operator-framework/operator-lifecycle-manager/releases/)), the latest version is displayed at the top, execute the commands as described under the "Scripted" section. At the time this document was written the latest version was v0.21.2:

```
curl -sL https://github.com/operator-framework/operator-lifecycle-manager/releases/download/v0.21.2/install.sh | bash -s v0.21.2
```


### Running

1. Install necessary Python modules:
```
pip3 install ansible openshift
```

2. Install Ansible collections

To install the required Ansible collections, run:

```
ansible-galaxy collection install -r requirements.yml
```

3. Connect to the cluster and make sure your `kubeconfig` points to the cluster.

4. Install the operators for OpenShift / Kubernetes:
```
ansible-playbook operators-openshift.yml
```
or
```
ansible-playbook operators-kubernetes.yml
```

## NFS provisioner

You can run the following commands upon a cluster to deploy the NFS provisioner to this cluster (either Kubernetes or OpenShift). You will need to uninstall the "Block Storage for VPC" add-on installed by default, to make the NFS provisioner work correctly.

```
$ helm repo add nfs-subdir-external-provisioner \
    https://kubernetes-sigs.github.io/nfs-subdir-external-provisioner/

$ helm install nfs-subdir-external-provisioner \
    nfs-subdir-external-provisioner/nfs-subdir-external-provisioner \
    --set nfs.server=$(</tmp/nfs_ip) \
    --set nfs.path=/mnt/nfs \
    --set storageClass.defaultClass=true \
    --set storageClass.onDelete=delete
```

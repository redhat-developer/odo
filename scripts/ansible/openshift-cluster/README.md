# IBM Cloud OpenShift Cluster

This ansible playbook deploys a VPC OpenShift cluster on IBM Cloud.
It uses the [IBM Cloud Ansible Collections](https://github.com/IBM-Cloud/ansible-collection-ibm/).

## VPC Resources

The following VPC infrastructure resources will be created (Ansible modules in
parentheses):

* Resource group (ibm_resource_group)
* VPC (ibm_is_vpc)
* Public gateway (ibm_is_public_gateway)
* VPC Subnet (ibm_is_subnet)
* Cloud Object Storage (ibm_resource_instance)
* VPC OpenShift Cluster (ibm_container_vpc_cluster)

All created resources (expect resource group) will be inside the created Resource Group.

## Configuration Parameters

The following parameters can be set by the user:

* `name_prefix`: Prefix used to name created resources
* `total_ipv4_address_count`: Number of IPv4 addresses available in the VPC subnet
* `cluster_zone`: Zone on which will be deployed the cluster
* `node_flavor`: Flavor of workers of the cluster. The list of flavors can be obtained with `ibmcloud ks flavors --zone ${CLUSTER_ZONE} --provider vpc-gen2`
* `workers`: Number of workers on the cluster

## Running

### Set API Key and Region

1. [Obtain an IBM Cloud API key].

2. Export your API key to the `IC_API_KEY` environment variable:

    ```
    export IC_API_KEY=<YOUR_API_KEY_HERE>
    ```

3. Export desired IBM Cloud region to the 'IC_REGION' environment variable:

    ```
    export IC_REGION=<REGION_NAME_HERE>
    ```

You can get available regions supporting Kubernetes clusters on the page https://cloud.ibm.com/docs/containers?topic=containers-regions-and-zones.

### Create

1. To create all resources, run the 'create' playbook:

    ```
    ansible-playbook create.yml
    ```

### Destroy

1. To destroy all resources run the 'destroy' playbook:

    ```
    ansible-playbook destroy.yml
    ```

## HOW TO: Create VPC Resources

The following VPC infrastructure resources will be created (Ansible modules in
parentheses):

* VPC (ibm_is_vpc)
* Subnet (ibm_is_subnet)
* VSI (ibm_is_instance)
* Floating IP Address (ibm_is_floating_ip)
* Security Group Rule (ibm_is_security_group_rule)

## Configuration Parameters

The following parameters can be set by the user:

* `name_prefix`: Prefix used to name created resources
* `vsi_image`: VSI image name ([retrieve available images])
* `vsi_profile`: VSI profile name ([retrieve available profiles])
* `ssh_public_key`: SSH Public Key
* `total_ipv4_address_count`: Number of IPv4 addresses in VPC Subnet
* `zone`: IBM Cloud zone
* `ssh-login-key`: Path to login key

## Running

### Set API Key and Region

1. [Obtain an IBM Cloud API key].

2. Export your API key to the `IC_API_KEY` environment variable:

    ```
    export IC_API_KEY=<YOUR_API_KEY_HERE>
    ```
    eg: `export IC_API_KEY=<KEY_VALUE>`

    Note: Modules also support the 'ibmcloud_api_key' parameter, but it is
    recommended to only use this when encrypting your API key value.

3. Export desired IBM Cloud region to the 'IC_REGION' environment variable:

    ```
    export IC_REGION=<REGION_NAME_HERE>
    ```
    eg : `export IC_REGION="eu-de"`
    
    Note: Modules also support the 'region' parameter.

### Create

To create all resources and test public SSH connection to the VM:

1. Update variables in 'vars.yml'
2. Run the 'create' playbook:

    ```
    ansible-playbook create.yml
    ```

### Delete only floating IP and VSI

1. To destroy only floating IP and VSI resources run the 'delete-only-vsi' playbook (resource names are read
   from 'vars.yml'):

    ```
    ansible-playbook destroy.yml
    ```


### Destroy

1. To destroy all resources run the 'destroy' playbook (resource names are read
   from 'vars.yml'):

    ```
    ansible-playbook destroy.yml
    ```

### List

1. To list available VSI Images and Profiles run the 'list_vsi_images_and_profiles' playbook:

    ```
    ansible-playbook list_vsi_images_and_profiles.yml
    ```


### Usefull Links
[Retrieve available images](#list-available-vsi-images-and-profiles)

[Retrieve available profiles](#list-available-vsi-images-and-profiles)

[Ansible search path](https://docs.ansible.com/ansible/latest/dev_guide/overview_architecture.html#ansible-search-path)

[Obtain an IBM Cloud API key](https://cloud.ibm.com/docs/account?topic=account-userapikey&interface=ui)

[Ansible search path](https://docs.ansible.com/ansible/latest/dev_guide/overview_architecture.html#ansible-search-path)



# SetUp VSI For podman testing

NOTE: only needed to update this image if setup need new/updated dependency
1. run  to create vsi for creating an image for test.

    `ansible-playbook create.yml`

NOTE: check var.yml file to check name of the vsi that will be created,by default its set with `name_prefix`-vsi

2. Stop the vsi with

    `ibmcloud is instance-stop --force <NAME OF VSI CREATED>`

3. Get value of volume ID and name that will be required for creating vsi image

    `ibmcloud is instances odo-test-automation-vsi --output json | jq '.[0].boot_volume_attachment.volume.id , .[0].boot_volume_attachment.volume.name'`
```
$ ibmcloud is instances odo-test-automation-vsi --output json | jq '.[0].boot_volume_attachment.volume.id , .[0].boot_volume_attachment.volume.name'
"r010-cc2b847e-61dd-405c-afaf-2422a852b0f1"
"wildfire-camper-captain-capture"
```

Note: we will use the name recovered when running the below command

4. Create vsi-image using

    `ibmcloud is image-create --source-volume r010-cc2b847e-61dd-405c-afaf-2422a852b0f1 --output json | jq '.name'`

```
$ ibmcloud is image-create --source-volume r010-cc2b847e-61dd-405c-afaf-2422a852b0f1 --output json | jq '.name'
"riding-stratus-embellish-snippet-rename"
```

Note: it takes 5-10 mins to create the VSI-Image

5. Check for created VSI-image

    `ibmcloud is images | grep riding-stratus-embellish-snippet-rename`
```
$ ibmcloud is images | grep riding-stratus-embellish-snippet-rename
r010-f402a62c-cabb-45ff-871f-639df3cb56d3   riding-stratus-embellish-snippet-rename             available    amd64   ubuntu-20-04-amd64                   20.04 LTS Focal Fossa Minimal Install                    5               private      user         none         Default          -
```
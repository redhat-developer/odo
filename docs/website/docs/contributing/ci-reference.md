---
title:  CI Reference
sidebar_position: 8
---

### Running PR test job on PSI
PSI contains an Openshift cluster running behind firewall, we are using prow to create request for PRs, we are running rabbitmq on a public cloud to access queue for creating jobs with internal jenkins(behind a firewall). Prow uses ci-firewall within [scripts/openshiftci-presubmit-all-tests.sh](https://github.com/openshift/odo/blob/main/scripts/openshiftci-presubmit-all-tests.sh) to create request to rabbitmq.
ci-firewall creates the following json message and passes it to the rabbitmq send queue as an env variable.
```
CI_MESSAGE='{"repourl": "repourl", "kind": "PR", "target": "target", "setupscript": "setupscript", "runscript": "runscript", "rcvident": "rcvident", "runscripturl": "http://url", "mainbranch": "master"}'
```
For every message in send queue a build is triggered using a jenkins robot account, jenkins then executes the build script to start the test on the node provided in SSHNodeFile(json contains information related to node), SSHNodeFile can contain multiple node information. CI-firewall then executes the test and send back logs for tests using a receive queue.

**Jenkins build script**
```shell
rm -rf ./*
curl -kJLO https://github.com/mohammedzee1000/ci-firewall/releases/download/${CI_FIREWALL_VERSION}/ci-firewall-linux-amd64.tar.gz
tar -xzf ./ci-firewall-linux-amd64.tar.gz && rm -rf ./ci-firewall-linux-amd64.tar.gz && chmod +x ./ci-firewall
curl -kJLO  <SSHNodeFile>/jenkins-nodes.json
curl -kJLO <kube-password>
NDFILE="$(pwd)/jenkins-nodes.json"
KUBEADMIN_PASSWORD_FILE="$(pwd)/kube-password"
./ ci-firewall work --sshnodesfile ${NDFILE} --env "OCP4X_API_URL=https://<url-to-ocp-cluster>" --env "OCP4X_KUBEADMIN_PASSWORD=$(cat ${KUBEADMIN_PASSWORD_FILE})" --env "CI=openshift"
rm -rf ./*
```

**SSHNodeFile**
```json
{
    "nodes": [{
          "name": "common name of node. example -Fedora 31-",
          "user": "username to ssh into the node with",
          "address": "The address of the node, like an ip or domain name without port",
          "port": 22,
          "baseos": "linux|windows|mac",
          "arch": "arch of the system eg amd64",
          "password": "not recommended but you can provide password of target node",
          "privatekey": "Optional again but either this or password MUST be given.",
          "tags": ["optional sample tags to append to logs from ssh node. Node `name is already attached as `ssh:name`"]
  }]
}
```

### Running integration tests on Prow

Prow is the Kubernetes or OpenShift way of managing workflow, including tests. Integration and periodic test targets for odo are passed through the script scripts/openshiftci-presubmit-all-tests.sh and scripts/openshiftci-periodic-tests.sh respectively available in the [odo](https://github.com/openshift/odo/tree/main/scripts) repository. Prow uses the script through the command attribute of the odo job configuration file in [openshift/release](https://github.com/openshift/release/tree/master/ci-operator/config/openshift/odo) repository.

For running integration test on 4.x cluster, job configuration file will be as follows:

```yaml
- as: integration-e2e
steps:
  cluster_profile: aws
  test:
  - as: integration-e2e-steps
    commands: scripts/openshiftci-presubmit-all-tests.sh
    credentials:
    - mount_path: /usr/local/ci-secrets/odo-rabbitmq
      name: odo-rabbitmq
      namespace: test-credentials
    env:
    - default: /usr/local/ci-secrets/odo-rabbitmq/amqpuri
      name: ODO_RABBITMQ_AMQP_URL
    from: oc-bin-image
    resources:
      requests:
        cpu: "2"
        memory: 6Gi
  workflow: ipi-aws
```

Similarly, for running periodic test on 4.x cluster, job configuration file will be as follows:

```yaml
- as: integration-e2e-periodic
cron: 0 */6 * * *
steps:
  cluster_profile: aws
  test:
  - as: integration-e2e-periodic-steps
    commands: scripts/openshiftci-periodic-tests.sh
    from: oc-bin-image
    resources:
      requests:
        cpu: "2"
        memory: 6Gi
  workflow: ipi-aws
```

To generate the odo job file, run `make jobs` in [openshift/release](https://github.com/openshift/release) for the odo pr and periodic tests.

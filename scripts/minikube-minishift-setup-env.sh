#!/usr/bin/env bash

shout() {
  set +x
  echo -e "\n.---------------------------------------\n${1}\n'---------------------------------------\n"
  set -x
}

# This si one of the variables injected by ci-firewall. Its purpose is to allow scripts to handle uniqueness as needed
SCRIPT_IDENTITY=${SCRIPT_IDENTITY:-"def-id"}

setup_operator() {
  SETUP_OPERATORS="./scripts/configure-cluster/common/setup-operators.sh"

  # The OLM Version
  LATEST_RELEASE=$(curl -L -s -H 'Accept: application/json' https://github.com/operator-framework/operator-lifecycle-manager/releases/latest)
  OLM_VERSION=$(echo $LATEST_RELEASE | sed -e 's/.*"tag_name":"\([^"]*\)".*/\1/')
  export OLM_VERSION=${OLM_VERSION:-"v0.17.0"}
  # Enable OLM for running operator tests
  curl -sL https://github.com/operator-framework/operator-lifecycle-manager/releases/download/$OLM_VERSION/install.sh | bash -s $OLM_VERSION

  # Install Operators for Operator tests
  sh $SETUP_OPERATORS
}

setup_kubeconfig() {
  export KUBECONFIG=$HOME/.kube/config
  if [[ ! -f $KUBECONFIG ]]; then
    echo "Could not find kubeconfig file"
    exit 1
  fi
  if [[ ! -z $KUBECONFIG ]]; then
    # Copy kubeconfig to current directory, to avoid clashes with other test runs
    # Read and Write permission to current kubeconfig file
    cp $KUBECONFIG "$(pwd)/config"
    chmod 640 "$(pwd)/config"
    export KUBECONFIG="$(pwd)/config"
  fi
}

setup_minikube_developer() {
  pwd=$(pwd)
  certdir=$(mktemp -d)
  cd $certdir
  shout "Creating a minikube developer user"
  openssl genrsa -out developer.key 2048
  openssl req -new -key developer.key -out developer.csr -subj "/CN=developer/O=minikube"
  openssl x509 -req -in developer.csr -CA ~/.minikube/ca.crt -CAkey ~/.minikube/ca.key -CAcreateserial -out developer.crt -days 500
  kubectl config set-credentials developer --client-certificate=developer.crt --client-key=developer.key
  kubectl config set-context developer-minikube --cluster=minikube --user=developer
  # Create role and rolebinding to allow the user necessary access to the cluster; this does not include access to CRD
  kubectl create -f - <<EOF
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: odo-user
rules:
- apiGroups: [""] # “” indicates the core API group
  resources: ["*"]
  verbs: ["*"]
- apiGroups: ["apps"]
  resources: ["*"]
  verbs: ["*"]
- apiGroups: ["operators.coreos.com"]
  resources: ["clusterserviceversions"]
  verbs: ["*"]
- apiGroups: ["redis.redis.opstreelabs.in"]
  resources: ["*"]
  verbs: ["*"]
- apiGroups: ["networking.k8s.io", "extensions"]
  resources: ["ingresses"]
  verbs: ["*"]
- apiGroups: ["route.openshift.io"]
  resources: ["*"]
  verbs: ["*"]
- apiGroups: ["apps.openshift.io"]
  resources: ["*"]
  verbs: ["*"]
---

kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: odo-user-binding
subjects:
- kind: User
  name: developer # Name is case sensitive
  apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: ClusterRole #this must be Role or ClusterRole
  name: odo-user
  apiGroup: rbac.authorization.k8s.io
EOF
  # Go back to the pwd
  cd $pwd || return
}

case ${1} in
minikube)
  mkStatus=$(minikube status)
  shout "| Checking if Minikube needs to be started..."
  if [[ "$mkStatus" == *"host: Running"* ]] && [[ "$mkStatus" == *"kubelet: Running"* ]]; then
    if [[ "$mkStatus" == *"kubeconfig: Misconfigured"* ]]; then
      minikube update-context
    fi
    setup_kubeconfig
    kubectl config use-context minikube
  else
    minikube delete
    shout "| Start minikube"
    minikube start --vm-driver=docker --container-runtime=docker
    setup_kubeconfig
    setup_operator
    setup_minikube_developer
  fi

  minikube version
  # Setup to find necessary data from cluster setup
  ## Constants

  set +x
  # Get kubectl cluster info
  kubectl cluster-info

  set -x
  # Set kubernetes env var as true, to distinguish the platform inside the tests
  export KUBERNETES=true

  # Create a developer user if it is not created already and change the context to use it after the setup is done
  kubectl config get-contexts developer-minikube || setup_minikube_developer
  kubectl config use-context developer-minikube
  ;;
*)
  echo "<<< Need parameter set to minikube or minishift >>>"
  exit 1
  ;;
esac

# Create a bin directory whereever script runs. This will be where all binaries that need to be in PATH will reside.
export HOME=$(pwd)/home
export GOPATH="$(pwd)/home/go"
export GOBIN="$GOPATH/bin"
mkdir -p $GOBIN

# Add GOBIN which is the bin dir we created earlier to PATH so any binaries there are automatically available in PATH
export PATH=$PATH:$GOBIN

# Prep for integration/e2e
shout "Building odo binaries"
make bin

# copy built odo to GOBIN
cp -avrf ./odo $GOBIN/
#!/usr/bin/env bash

shout() {
  set +x
  echo -e "\n.---------------------------------------\n${1}\n'---------------------------------------\n"
  set -x
}

# Create a bin directory whereever script runs. This will be where all binaries that need to be in PATH will reside.
export HOME=`pwd`/home
export GOPATH="`pwd`/home/go"
export GOBIN="$GOPATH/bin"
mkdir -p $GOBIN
# This si one of the variables injected by ci-firewall. Its purpose is to allow scripts to handle uniqueness as needed
SCRIPT_IDENTITY=${SCRIPT_IDENTITY:-"def-id"}

# Add GOBIN which is the bin dir we created earlier to PATH so any binaries there are automatically available in PATH
export PATH=$PATH:$GOBIN

# Prep for integration/e2e
shout "Building odo binaries"
make bin

# copy built odo to GOBIN
cp -avrf ./odo $GOBIN/

setup_kubeconfig() {
    export KUBECONFIG=$HOME/.kube/config
    if [[ ! -f $KUBECONFIG ]]; then
        echo "Could not find kubeconfig file"
        exit 1
    fi
    if [[ ! -z $KUBECONFIG ]]; then
        # Copy kubeconfig to current directory, to avoid clashes with other test runs
        # Read and Write permission to current kubeconfig file
        cp $KUBECONFIG "`pwd`/config"
        chmod 640 "`pwd`/config"
        export KUBECONFIG="`pwd`/config"
    fi
}

setup_minikube_developer() {
  openssl genrsa -out developer.key 2048
  openssl req -new -key developer.key -out developer.csr -subj "/CN=developer/O=minikube"
  openssl x509 -req -in developer.csr -CA ~/.minikube/ca.crt -CAkey ~/.minikube/ca.key -CAcreateserial -out developer.crt -days 500
  kubectl config set-credentials developer --client-certificate=developer.crt --client-key=developer.key
  kubectl config set-context developer-context --cluster=minikube --user=developer
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
- apiGroups: ["networking.k8s.io"]
  resources: ["ingresses"]
  verbs: ["*"]
- apiGroups: ["route.openshift.io"]
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
}

case ${1} in
    minishift)
        export MINISHIFT_ENABLE_EXPERIMENTAL=y 
        export PATH="$PATH:/usr/local/go/bin/"
        sh .scripts/minishift-start-if-required.sh
        ;;
    minikube)
        mkStatus=$(minikube status)
        shout "| Checking if Minikube needs to be started..."
        if [[ "$mkStatus" == *"host: Running"* ]] && [[ "$mkStatus" == *"kubelet: Running"* ]]; then 
            if [[ "$mkStatus" == *"kubeconfig: Misconfigured"* ]]; then
                minikube update-context
            fi
            setup_kubeconfig
            setup_minikube_developer
            kubectl config use-context minikube
        else
            minikube delete
            shout "| Start minikube"
            minikube start --vm-driver=docker --container-runtime=docker
            setup_kubeconfig
        fi
        
        minikube version
        # Setup to find nessasary data from cluster setup
        ## Constants
        SETUP_OPERATORS="./scripts/configure-cluster/common/setup-operators.sh"

        # The OLM Version
        export OLM_VERSION="v0.17.0"
        # Enable OLM for running operator tests
        curl -sL https://github.com/operator-framework/operator-lifecycle-manager/releases/download/$OLM_VERSION/install.sh | bash -s $OLM_VERSION

        set +x
        # Get kubectl cluster info
        kubectl cluster-info
        
        set -x
        # Set kubernetes env var as true, to distinguish the platform inside the tests
        export KUBERNETES=true

        # Create Operators for Operator tests
        sh $SETUP_OPERATORS

        # Change the user to developer after the operators have been setup
        kubectl config use-context developer-context
        ;;
    *)
        echo "<<< Need parameter set to minikube or minishift >>>"
        exit 1
        ;;
esac

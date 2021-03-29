#!/usr/bin/env bash

shout() {
  set +x
  echo -e "\n.---------------------------------------\n${1}\n'---------------------------------------\n"
  set -x
}

# Create a bin directory whereever script runs. This will be where all binaries that need to be in PATH will reside.
mkdir bin artifacts
# Change the default location of go's bin directory (without affecting GOPATH). This is where compiled binaries will end up by default
# for eg go get ginkgo later on will produce ginkgo binary in GOBIN
export GOBIN="`pwd`/bin"

# Set kubeconfig to current dir. This ensures no clashes with other test runs
export KUBECONFIG="`pwd`/config"
export ARTIFACTS_DIR="`pwd`/artifacts"
export CUSTOM_HOMEDIR=$ARTIFACT_DIR

# This si one of the variables injected by ci-firewall. Its purpose is to allow scripts to handle uniqueness as needed
SCRIPT_IDENTITY=${SCRIPT_IDENTITY:-"def-id"}

# Add GOBIN which is the bin dir we created earlier to PATH so any binaries there are automatically available in PATH
export PATH=$PATH:$GOBIN

# Prep for integration/e2e
shout "Building odo binaries"
make bin

# copy built odo to GOBIN
cp -avrf ./odo $GOBIN/
shout "| Getting ginkgo"
make goget-ginkgo
hm = `hostname`
echo `hostname`
shout "| `hostname`"

case ${1} in
    minishift)
        shout "| Setting environemnt for minishift"
        export MINISHIFT_ENABLE_EXPERIMENTAL=y 
        export PATH="$PATH:/usr/local/go/bin/"
        export GOPATH=$HOME/go
        mkdir -p $GOPATH/bin
        export PATH="$PATH:$(pwd):$GOPATH/bin"

        #Export github API token to prevent pull limit error
        export MINISHIFT_GITHUB_API_TOKEN=${MINISHIFT_GITHUB_API_TOKEN_VALUE}

        # Verify if minishift or openshift are stopped and take appropriate actions
        shout "| Checking if minishift needs to be installed..."
        minishift version
        msStatus=$(minishift status)
        if [[ "$msStatus" == *"command not found"* ]]
        then
            shout "| Installing Minishift"
            curl -Lo minishift.tgz https://github.com/minishift/minishift/releases/download/v1.34.3/minishift-1.34.3-linux-amd64.tgz
            tar -xvzf minishift.tgz
            sudo mv minishift-1.34.3-linux-amd64/minishift /usr/local/bin
            rm minishift.tgz
        fi

        shout "| Checking if Minishift needs to be started..."
        if [[ "$msStatus" == *"Does Not Exist"* ]] || [[ "$msStatus" == *"Minishift:  Stopped"* ]]
        then 
            shout "| Starting Minishift..."
            (minishift start --vm-driver kvm --show-libmachine-logs -v 5)
        else 
            if [[ "$msStatus" == *"OpenShift:  Stopped"* ]];
                then 
                shout "| Minishift is running but Openshift is stopped, restarting minishift..."
                (minishift stop)
                (minishift start --vm-driver kvm --show-libmachine-logs -v 5)
            else
                if [[ "$msStatus" == *"Running"* ]]; 
                    then shout "| Minishift is running"
                fi
            fi
        fi

        # Check if service-catalog, automation-service-broker, and template-service-broker need to be installed
        compList=$(minishift openshift component list)
        shout "| Checking if required components need to be installed..."
        if [[ "$compList" == *"service-catalog"* ]] 
        then 
            shout "| service-catalog already installed "
        else 
                shout "| Installing service-catalog ..."
                (minishift openshift component add service-catalog)
        fi
        if [[ "$compList" == *"automation-service-broker"* ]] 
        then 
            shout "| automation-service-broker already installed "
        else 
                shout "| Installing automation-service-broker ..."
                (minishift openshift component add automation-service-broker)
        fi
        if [[ "$compList" == *"template-service-broker"* ]] 
        then 
            shout "| template-service-broker already installed "
        else 
                shout "| Installing template-service-broker ..."
                (minishift openshift component add template-service-broker)
        fi

        ;;
    minikube)
        shout "| Start minikube"
        # Delete minikube instance, if in anycase already exists
        minikube delete
        minikube start --vm-driver=docker --container-runtime=docker
        set +x
        # Get kubectl cluster info
        kubectl cluster-info

        set -x
        # Set kubernetes env var as true, to distinguish the platform inside the tests
        export KUBERNETES=true
        ;;
    *)
        echo "<<< Need (parameter) CLUSTER env. variable set to minikube or minishift >>>"
        exit 1
        ;;
esac

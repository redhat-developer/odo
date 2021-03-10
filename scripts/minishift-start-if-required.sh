#!/usr/bin/env bash

shout() {
   set +x
  echo -e "\n.---------------------------------------\n${1}\n'---------------------------------------\n"
   set -x
}

set -ex
export MINISHIFT_GITHUB_API_TOKEN=$MINISHIFT_GITHUB_API_TOKEN_VALUE
msStatus=$(minishift status)
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

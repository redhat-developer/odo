#!/usr/bin/env bash

shout() {
   set +x
  echo -e "\n.---------------------------------------\n${1}\n'---------------------------------------\n"
   set -x
}

set -ex

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

shout "| Adding required components ..."
minishift openshift component add service-catalog
minishift openshift component add automation-service-broker
minishift openshift component add template-service-broker
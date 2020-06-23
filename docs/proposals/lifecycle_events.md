# Support execution of the commands based on lifecycle events.

## Abstract
Add support for the lifecycle events which can be defined in 2.0.0 devfiles.

Our proposed solution is to perform the PreStart and PostStart events as a part of the **odo push** command, and the PreStop and Poststop events as a part of the **odo delete** command.

## Motivation
Devfile support commands that can be triggered based on dev lifecycle events. Odo will need to support/execute these commands at appropriate times within the flow.

With the implementation of livecycle events, stack creators will be able to leverage all the capabilities of devfiles when building the correct experience/features for their stacks.

## User Stories
Container initiliazation support - https://github.com/openshift/odo/issues/2936

Lifecycle bindings for specific events - https://github.com/devfile/kubernetes-api/issues/32


## Design overview

The lifecycle events that have been implemented in Devfile 2.0.0 are generic. So their effect/meaning might be slightly different when applying them to a workspace (Che) or a single project (odo).

Our proposal is that the PreStart and PostStart events are designed to run as a part of **odo push**, and that the PreStop and PostStop events run as a part of **odo delete**.



### The flow for **odo push** including the *PreStart* and *PostStart* lifecycle events will be as follows:
 - PreStart - Devfile command gets translated to entrypoint for specified init container - one time only
 - Container initialisation (as done today)
 - PostStart - Exec the specified command(s) in the container - one time only
 - Rest of the command execution (as we do today - exec commands for build/run/test/debug)

**PreStart**:
 - The PreStart command(s) that are specified in the Devfile will be translated into entry points for their specified containers, and added to the pod spec as init containers. 
 - These will only run on the first odo push, or when doing a force push (odo push -f). 
 - These are commands that you’d want to run before the main containers are created. 

    **Note: We need some good use cases to understand when the user would run commands as PreStart opposed to PostStart.**

**Component Initialization**
 - After the PreStart event has completed, the containers specified in the Devfile are initialised and created if the component doesn’t already exist.

**PostStart**
 - If the component is newly created, the PostStart events are executed sequentially within the containers given in the Devfile. 
 - These commands could all be in the same container, or all in different ones. 

**Command Execution**
 - After PostStart, we’d run the usual build/run/test/debug commands (depending on what sort of odo push parameters the user has provided) as usual.

The reason we have chosen to use init containers for PreStart rather than to be consistent, is because we don’t think there would be any difference in the PreStart and PostStart stages if they both ran in the same containers. If we were instead initialising the containers, running PreStart in those containers, and then PostStart in those containers, what would be different to running them all in PreStart, or all in PostStart? At least with init containers for PreStart, there is a definitive difference between the two, and therefore reason to include them both as separate events.



### The flow for **odo delete** including the *PreStop* and *PostStop* lifecycle events will be as follows:
**PreStop**
 - Exec the specified command(s) in their respective containers before deleting the deployment and any clean up begins.

**Clean up resources**
 - Clean up the pod and deployment etc. (as done today)

**PostStop**
 - Execute the command specified by postStop
 - Would we be spinning up a new container exclusively for the PostStop commands?
 - Would the command need to run on the host instead for local clean up? 



### Conclusions:
 - We think that the most important event is the PostStart because it has a clear, reasonable use case within odo’s flow. 
 - PreStart and PreStop could potentially be useful, but in a much more niche range. It would be important to clearly document the execution order/process for these events, because it would likely cause confusion. 
 - We aren’t fully conclusive on the necessity of PostStop, and what it would be useful for. 
 
 ## Future Evolution
 


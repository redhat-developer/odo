# Support execution of the commands based on lifecycle events.

## Abstract
Add support for the lifecycle events which can be defined in 2.0.0 devfiles.

Our proposed solution is to perform the preStart and postStart events as a part of the **odo push** command, and the preStop and Poststop events as a part of the **odo delete** command.

Lifecycle bindings for specific events - https://github.com/devfile/kubernetes-api/issues/32


## Motivation
Devfile support commands that can be triggered based on dev lifecycle events. Odo will need to support/execute these commands at appropriate times within the flow.

With the implementation of livecycle events, stack creators will be able to leverage all the capabilities of devfiles when building the correct experience/features for their stacks.

## User Stories
Each lifecycle event would be suited to individual user stories. We propose that at least one issue is made for each of them for tracking and implementation purposes.

An issue has already been created for postStart: [container initiliazation support](https://github.com/openshift/odo/issues/2936)

## Design overview

The lifecycle events that have been implemented in Devfile 2.0.0 are generic. So their effect/meaning might be slightly different when applying them to a workspace (Che) or a single project (odo). However, this proposal will aim to provide a similar user experience within the scope of a single project.

Our proposal is that the preStart and postStart events are designed to run as a part of **odo push**, and that the preStop and postStop events run as a part of **odo delete**.



### The flow for **odo push** including the *preStart* and *postStart* lifecycle events will be as follows:
 - preStart - Devfile command gets translated to entrypoint for specified init container - one time only
 - Container initialisation (as done today)
 - postStart - Exec the specified command(s) in the container - one time only
 - Rest of the command execution (as we do today - exec commands for build/run/test/debug)

**preStart**:
 - The preStart command(s) that are specified in the Devfile will be translated into entry points for their specified containers, and added to the pod spec as init containers. 
 - These will only run on the first odo push, or when doing a force push (odo push -f). 
 - These are commands that you’d want to run before the main containers are created. 

    **Note: We need some good use cases to understand when the user would run commands as preStart opposed to postStart.**

**Component Initialization**
 - After the preStart event has completed, the containers specified in the Devfile are initialised and created if the component doesn’t already exist.

**postStart**
 - If the component is newly created, the postStart events are executed sequentially within the containers given in the Devfile. 
 - These commands could all be in the same container, or all in different ones. 

**Command Execution**
 - After postStart, we’d run the usual build/run/test/debug commands (depending on what sort of odo push parameters the user has provided) as usual.

The reason we have chosen to use init containers for preStart rather than to be consistent, is because we don’t think there would be any difference in the preStart and postStart stages if they both ran in the same containers. If we were instead initialising the containers, running preStart in those containers, and then postStart in those containers, what would be different to running them all in preStart, or all in postStart? At least with init containers for preStart, there is a definitive difference between the two, and therefore reason to include them both as separate events.



### The flow for **odo delete** including the *preStop* and *postStop* lifecycle events will be as follows:
**preStop**
 - Exec the specified command(s) in their respective containers before deleting the deployment and any clean up begins.

**Clean up resources**
 - Clean up the pod and deployment etc. (as done today)

**postStop**
 - Execute the command specified by postStop
 - Would we be spinning up a new container exclusively for the postStop commands?
 - Would the command need to run on the host instead for local clean up? 



### Conclusions:
 - We think that the most important event is the postStart because it has a clear, reasonable use case within odo’s flow. 
 - preStart and preStop could potentially be useful, but in a much more niche range. It would be important to clearly document the execution order/process for these events, because it would likely cause confusion. 
 - We aren’t fully conclusive on the necessity of postStop, and what it would be useful for. 
 
 ## Future Evolution
 


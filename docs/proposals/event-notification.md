# ODO build and application status notification 

## Summary

With this proposal and [it's related issue](https://github.com/redhat-developer/odo/issues/2550) we examine how to consume build/run event output from odo in order to allow external tools (such as IDEs) to determine the application/build status of an odo-managed application.

In short, the proposal is:
- New flag for push command, `odo push -o json`: Outputs JSON events that correspond to what devfile actions/commands are being executed by push.
- New odo command `odo component status -o json`: Calls `supervisord ctl status` to check the running container processes, checks the container/pod status every X seconds (and/or a watch, for Kubernetes case), and sends an HTTP/S request to the application URL. The result of those is output as JSON to be used by external tools to determine if the application is running.
- A standardized markup format for communicating detailed build status, something like: `#devfile-status# {"buildStatus":"Compiling application"}`, to be optionally included in devfile scripts that wish to provide detailed build status to consuming tools.

## Terminology

With this issue, we are looking at adding odo support for allowing external tools to gather build status and application status. We further divide both statuses into detailed and non-detailed, with detailed principally being related to statuses that can be determine by looking at container logs.

**Build status**: A simple build status indicating whether the build is running (y/n), and whether the last build succeeded or failed. This can be determined based on whether odo is running a dev file action/command, and the process error code (0 = success, non-0 = failed) of that action/command.

**Detailed build status**: An indication of which step in the build process the build is in. For example: are we 'Compiling application', or are we 'Running unit tests'?

**App status**: Determine if an application is running, using container status (for both local and Kubernetes, various status: containercreating, started, restarting, etc), `supervisord ctl status`, or whether an HTTP/S request to the root application URL returns an HTTP response with any status code.

**Detailed application status**: 
- While app status works well for standalone application frameworks (Go, Node Express, Spring Boot), it works less well for full server runtimes such as Java EE application servers like OpenLiberty/WildFly that may begin responding to Web requests before the user's deployed WAR/EAR application has finished startup. 
- Since these application servers are built to serve multiple concurrently-deployed applications, it is more difficult to determine the status of any specific application running on them. The lifecycle of the application server differs from the lifecycle of the application running inside the application server. 
- Fortunately, in these cases the IDE/consuming tool can use the console logs (from `odo log`) from the runtime container to determine a more detailed application status.
- For example, OpenLiberty (as an example of an application-server-style container) prints a specific code when an application is starting  `CWWKZ0018I: Starting application {0}.`, and another when it has started. `CWWKZ0001I: Application {0} started in {1} seconds.`
- Odo itself should NOT know anything about these specific application codes; knowing how these translate into a detailed application status would be the responsibility of the IDE/consuming tool. Odo's role here is only to provide the console output log. 
- In the future, we could add these codes into the devfile to give Odo itself some agency over determining the detailed application status, but for this proposal responsibility is left with the consuming tool.

**Devfile writer**: A devfile writer may be a runtime developer (for example, a Red-Hatter working on WildFly or an IBMer working on OpenLibery) creating a devfile for their organization's runtime (for example 'OpenLiberty w/ Maven' dev file), or an application developer creating/customizing a dev file for use with their own application. In either case, the devfile writer must be familiar with the semantics of both odo and the devfile.

## JSON-based odo command behaviours to detect app and build status

New odo commands and flags:
- `odo push -o json`
- `odo component status -o json`

With these two additions, an IDE or similar external tool can detect build running/succeeded/failed, application starting/started/stopping/stopped, and (in many cases) get a 'detailed app status' and/or 'detailed build status'.

### Build status notification via `odo push -o json`

`odo push -o json` is like standard `odo push`, but instead it outputs JSON events (including action console output) instead of text. This allows the internal state of the odo push process to be more easily consumed by external tools.

Several different types of JSON-formatted events would be output, which correspond to odo container command/action executions:
- Dev file command execution begun *(command name, start timestamp)*
- Dev file command execution completed *(error code, end timestamp)*
- Dev file action execution begun *(action name, parent command name, start timestamp)*
- Dev file action execution completed *(action name, error code, end timestamp)*
- Log/console stdout from the actions, one line at a time *(for example, `mvn build` output).* (timestamp)
  
(Exact details for which fields are included with events are TBD)

In addition, `odo push -o json` should return a non-zero error code if one of the actions returned a non-zero error code, otherwise zero is returned.

### `odo push -o json` example output

This is what an `odo push -o json` command invocation would look like:
```
odo push -o json
{ "devFileCommandExecutionBegun": { "commandName" : "build", "timestamp" : "(UTC unix epoch seconds.microseconds)" } }
{ "devFileActionExecutionBegun" : { "commandName" : "build", "actionName" : "run-build-script", "timestamp" : "(...)" } }
{ "logText" : { "text:" "one line of text received\\n another line of text received", "timestamp" : "(...)" } }
{ "devFileActionExecutionComplete" : { "errorCode" : 1, ( same as above )} }
{ "logText" : { "text": (... ), "timestamp" : "(...)" } } # Log text is interleaved with events 
{ "devFileCommandExecutionComplete": { "success" : false, (same as above) } }

(Exact details on event name, and JSON format are TBD; feedback welcome!)
```

These events allow an external odo-consuming tool to determine the build status of an application (build succeeded, build failed, build not running).

Note that unlike other machine-readable outputs used in odo, each individual line is a fully complete and parseable JSON document, allowing events to be streamed and processed by the consuming tool one-at-a-time, rather than waiting for all the events to be received before being parseable (which would be required if the entire console output was one single JSON document, as is the case for other odo machine-readable outputs.)

### Detailed build status via JSON+custom markup

For detailed build status, it is proposed that devfile writers may *optionally* include custom markup in their devfile actions which indicate a detailed build status:
- If a dev file writer wanted to communicate that the current command/action were compiling the application, they would insert a specific markup string (`#devfile-status#`) at the beginning of a console-outputted line, and then between those two fields would be a JSON object with a single field `buildStatus`:
  - For example: `#devfile-status# {"buildStatus":"Compiling application"}` would then communicate that the detailed build status should be set to `Compiling application`.
- Since this line would be output as container stdout, it would be included as a `logText` JSON event, and the consuming tool can look for this markup string and parse the simple JSON to extract the detailed build status.
- Feedback welcome around exact markup text format.

The build step (running as a bash script, for example, invoked via an action) of a devfile might then look like this:
```
#!/bin/bash
(...)
echo "#devfile-status# {'buildStatus':'Compiling application'}
mvn compile
echo "#devfile-status# {'buildStatus':'Running unit tests'}
mvn test
```

This 'detailed build status' markup text is *entirely optional*: if this markup is not present, the odo tool can still determine build succeeded/failed and build running/not-running using the other `odo push -o json` JSON events. 

### App status notification via `odo component status -o json` and `odo log --follow`

In general, within the execution context that odo operates, there are a few ways for us to determine the application status:
1) Will the application respond to an HTTP/S request sent to its exposed URL? 
2) What state is the container in? (running/container creating/restarting/etc -- different statuses between local and Kube but same general idea)
3) Are the container processes running that are managed by supervisord? We check this by calling `supervisord ctl status`.
4) In the application log, specific hardcoded text strings can be searched for (for example, OpenLiberty outputs defined status codes to its log to indicate that an app started.) But, note that we definitely don't want to hardcode specific text strings into ODO: instead, this proposal leaves it up to the IDE to process the output from the `odo log` command. Since the `odo log` command output would contain the application text, IDEs can provide their own mechanism to determine status for supported devfiles (and in the future we may wish to add new devfile elements for these strings, to allow odo to do this as well).

Ideally, we would like for odo to provide consuming tools with all 4 sets of data. Thus, as proposed:
- 1, 2 and 3 are handled by a new `odo component status -o json` command, described here.
- 4 is handled by the existing unmodified `odo log --follow` command.

The new proposed `odo component status -o json` command will:
- Be a *long-running* command that will continue outputing status until it is aborted by the parent process.
- Every X seconds, send an HTTP/S request to the URLs/routes of the application as they existed when the command was first executed. Output the result as a JSON string.
- Every X seconds (or using a Kubernetes watch, where appropriate), check the container status for the application, based on the application data that was present when the command was first issued. Output the result as a JSON string.
- Every X seconds call `supervisord ctl status` within the container and report the status of supervisord's managed processes.

**Note**: This command will NOT, by design, respond to route/application changes that occur during or after it is first invoked. It is up to consuming tools to ensure that the `odo component status` command is stopped/restarted as needed. 
  - For example, if the user tells the IDE to delete their application with the IDE UI, the IDE will call `odo delete (component)`; at the same time, the IDE should also abort any existing `odo component status` commands that are running (as these are no longer guaranteed to return a valid status now that the application itself no longer exists). `odo component status` will not automatically abort when the application is deleted (because it has no reliable way to detect this in all cases).
  - Another example: if the IDE adds a new URL via `odo url create [...]`, any existing `odo component status` commands that are running should be aborted, as these commands would still only be checking the URLs that existed when the command was first invoked (eg there is intentionally no cross-process notification mechanism for created/updated/deleted URLs implemented as part of this command.)
  - See discussion of this design decision in 'Other solutions considered' below

This is an example an `odo component status -o json` command invocation look like:
```
odo component status -o json

{ "componentURLStatus" : { "url" : "https://(...)", "response" : "true", "responseCode" : 200, "timestamp" : (UTC unix epoch seconds.microseconds) } }
{ "componentURLStatus" : { "url" : "https://(...)", "response" : "false", error: "host unreachable", "timestamp" : (...) } }
{ "containerStatus" : { "status" : "containercreating", "timestamp" : (...)} }
{ "containerStatus" : { "status" : "running", "timestamp" : (...)} }
{ "supervisordCtlStatus" : { "name": "devrun", "status" : "STARTED", "timestamp" : (...)} }
(...)

(Exact details on event name, and JSON format are TBD; feedback welcome!)
```

To keep from overwhelming the output, only state changes would be printed (after an initial state output), rather than every specific event.

## Consumption of odo via external tools, such as IDEs

Based on our existing knowledge from previously building similar application/build-status tracking systems in Eclipse Codewind, we believe the above described commands should allow any external tool to provide a detailed status for odo-managed applications.

The proposed changes ensure that the the high-level logic around tracking application changes across time can be managed by external tools (such as IDEs) as desired, without the need to leak/"pollute" odo with any of these details. These changes give consuming tools all the data they need ensure fast, reliable, up-to-date and (where possible) detailed build/application status.


### What happens if the network connection is lost while executing these commands?

One potential challenge is how to handle network connection instability when the push/log/status commands are actively running. Both odo, and any external consuming tools, should be able to ensure that the odo-managed application can be returned to a stable state once the connection is re-established.

We can look at how each command should handle a temporary network disconnection:
- If network connection is dropped during *push*: consuming tool can restart the push command from scratch. Well-written dev files should be nuking any existing build processes (for example, when running a 'build' action, that build action should look for any old maven processes and kill them, if there are any that are already running; or said another way, it is up to the build action of a devfile to ensure that container state is consistent before starting a new build)
- If connection is dropped during *logs*: start a new tail, and then do a full 'get logs' to make sure we didn't miss anything; match up the two (the full log and the tail) as best as possible, to prevent duplicates. (The Kubernetes API may already have a better way of handling this; this is the "naive" algorithm)
- If connection is dropped during *status*: no special behaviour is needed here.

## Other solutions considered

Fundamentally, this proposal needs to find a solution to this scenario:

1) IDE creates a URL (calls `odo url create`) and pushes the user's code (calls `odo push`)
2) To get the status of that app, the IDE runs `odo component status -o json` to start the long-running odo process. The status command then helpfully reports the pod/url/supervisord container status, which allows the IDE to determine when the app process is up.
3) *[some time passes]*
4) IDE creates a new URL (or performs some other action that invalidates the existing `odo status` state, such as `odo component delete`) by calling `odo url create`.
5) The long-running `odo status` process is still running, but somehow needs to know about the new URL from step 4 (or other events).

Thus, in some way, that existing long-running `odo status` process needs to be informed of the new event (a new url event, a component deletion event, etc). Since these events are generated across independent OS processes, this requires some form of [IPC](https://en.wikipedia.org/wiki/Inter-process_communication). 

### Some options in how to communicate these data across independent odo processes (in ascending order of complexity)

#### 1) Get the IDE/consuming tool to manage the lifeycle of `odo component status`

This is the solution proposed in this proposal, and is included for contrast.

Since the IDE has a lifecycle that is greater than each of the individual calls to `odo`, and the IDE is directly and solely responsible for calling odo (when the user is interacting with the IDE), it is a good fit to ensure the state of `odo component status` is up-to-date and consistent.

But this option is by no means a perfect solution: 
- This does introduce complexity on the IDE side, as the IDE needs to keep track of which `odo` processes are running for each component, and it needs to know when/how to respond to actions (delete/url create/etc). But since the IDE is a monolithic process, this is at least straightforward (I mocked up the algorithm that the IDE will use in each case, which I can share if useful.)
- This introduces complexity for EVERY new IDE/consuming tool that uses this mechanism; rather than solving it once in ODO, it needs to be solved X times for X IDEs.
- Requires multiple concurrent long-running odo processes per odo-managed component

#### 2) `odo component status` could monitor files under the `.odo` directory in order to detect changes; for example, if a new URL is added to `.odo/env/env.yaml`, `odo component status` would detect that and update the URLs it is checking

This sounds simple, but is surprisingly difficult:
- No way to detect a delete operation just by watching `.odo` directory: at present, `odo delete` does not delete/change any of the files under `.odo`
- Partial file writes/atomicity/file locking: How to ensure that when `odo component status` reads a file that it has been fully written by the producing process? One way is to use file locks, but that means using/testing each supported platform's file locking mechanisms. Then need to implement a cross-process polling mechanism.
- Or, need to implement a cross-platform [filewatching mechanism](https://github.com/fsnotify/fsnotify): We need a way to watch the `.odo` directory and respond to I/O events to the files, either by modification. 
- Windows: Unlike other supported platforms, Windows has a number of quirky file-system behaviours that need to be individually handled. The most relevant one here is that Windows will not let you delete/modify a file in one process if another process is holding it open (we have been bitten by this a number of times in Codewind)
- Need to support all filesystems: some filesystems have different file write/locking/atomicity guarantees for various operations.


#### 3) Convert odo into a multi-process client-server architecture

Fundamentally this problem is about how to share state between odo processes; if odo instead used a client-server architure, odo state could be centrally/consistently managed via a single server process, and communicated piecemeal to odo clients.

As one example of this, we could create a new odo process/command (`odo status --daemon`?) that would listen on some IPC mechanism (TCP/IP sockets/named pipes/etc) for events from individual odo commands:
1) IDE runs `odo status --daemon --port=32272` as a long-running process; the daemon listens on localhost:32272 for events. The daemon will output component/build status to stdout, which the IDE can provide back to the user.
2) IDE calls `odo url create` to create a new URL, but includes the daemon information in the request: `odo url create --port=32272 (...)`
3) The `odo url create` process updates the `env.yaml` to include the URL, then connects to the daemon on `localhost:32272` and informs the daemon of the new url.
4) The daemon receives the new URL event, and reconciles it with its existing state for the application, and begins watching the new URL.
(This would be need to be implemented for every odo change event)


Drawbacks:
- Odo's code currently assumes that commands are short-lived, mostly single-threaded, and compartmentalized; switching to a server would fundamentally alter this presumption of existing code
- Much more complex to implement versus other options: requires changing the architecture of the odo tool into a multithreaded client-server model, meaning many more moving parts, and [the perils of distributed computing](https://en.wikipedia.org/wiki/Fallacies_of_distributed_computing).
- Most be cross-platform; IPC mechanisms/behaviour are VERY platform-specific, so we probably need to use TCP/IP sockets. 
- But, if using HTTP/S over TCP/IP socket, we need to secure endpoints; just listening on localhost [is not necessarily enough to ensure local-only access](https://bugs.chromium.org/p/project-zero/issues/detail?id=1524).
- Plus some corporate developer environments may use strict firewall rules that prevent server sockets, even on localhost ports.

Variants on this idea: 1) a new odo daemon/LSP-style server process that was responsible for running ALL odo commands; calls to the `odo` CLI would just initiate a request to the server, and the server would be responsible for performing the action and monitoring the status

### Proposed option vs options 2/3

Hopefully the inherent complexity of options 2-3 is fully conveyed above, but if you all have another fourth option, let me know.

Ultimately, this proposal (option 1) cleanly solves the problem, puts the complexity in the right place (the IDE), is straight-forward to implement, is not time consuming to implement, and does not fundamentally alter the odo architecture.

And this option definitely does not in any way tie our hands in implementing a more complex solution in the future if/when we our requirements demand it.

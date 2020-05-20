# ODO build and application status notification 

## Summary

With this proposal and [it's related issue](https://github.com/openshift/odo/issues/2550) we examine how to consume build/run event output from odo in order to allow external tools (such as IDEs) to determine the application/build status of an odo-managed application.

In short, the proposal is:
- New flag for push command, `odo push -o json`: Outputs JSON events that correspond to what devfile actions/commands are being executed by push.
- New odo command `odo component status -o json`: Sends an HTTP/S request to the application URL, and checks the container/pod status every X seconds (and/or a watch, for Kubernetes case). The result of both is output as JSON to be used by external tools to determine if the application is running.
- A standardized markup format for communicating detailed build status, something like: `#devfile-status# {"buildStatus":"Compiling application"}`, to be optionally included in devfile scripts that wish to provide detailed build status to consuming tools.

## Terminology

With this issue, we are looking at adding odo support for allowing external tools to gather build status and application status. We further divide both statuses into detailed and non-detailed, with detailed principally being related to statuses that can be determine by looking at container logs.

**Build status**: A simple build status indicating whether the build is running (y/n), and whether the last build succeeded or failed. This can be determined based on whether odo is running a dev file action/command, and the process error code (0 = success, non-0 = failed) of that action/command.

**Detailed build status**: An indication of which step in the build process the build is in. For example: are we 'Compiling application', or are we 'Running unit tests'?

**App status**: Determine if an application is running, using container status (for both local and Kubernetes, various status: containercreating, started, restarting, etc), or whether an HTTP/S request to the root application URL returns an HTTP response with any status code.

**Detailed application status**: 
- While app status works well for standalone application frameworks (Go, Node Express, Spring Boot), it works less well for full server runtimes such as Java EE application servers like OpenLiberty/WildFly that may begin responding to Web requests before the user's deployed WAR/EAR application has finished startup. 
- Since these application servers are built to serve multiple concurrently-deployed applications, it is more difficult to determine the status of any specific application running on them. The lifecycle of the application server differs from the lifecycle of the application running inside the application server. 
- Fortunately, in these cases the IDE/consuming tool can use the console logs (from `odo log`) from the runtime container to determine a more detailed application status.
- For example, OpenLiberty (as an example of an application-server-style container) prints a specific code when an application is starting  `CWWKZ0018I: Starting application {0}.`, and another when it has started. `CWWKZ0001I: Application {0} started in {1} seconds.`
- Odo itself should NOT know anything about these specific application codes; knowing how these translate into a detailed application status would be the responsibility of the IDE/consuming tool. Odo's role here is only to provide the console output log. 
- In the future, we could add these codes into the devfile to give Odo itself some agency over determining the detailed application status, but for this proposal I am leaving this responsibility with the consuming tool.

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
3) In the application log, specific hardcoded text strings can be searched for (for example, OpenLiberty outputs defined status codes to its log to indicate that an app started.)

Ideally, we would like for odo to provide consuming tools with all 3 sets of data. Thus, as proposed:
- 1 and 2 are handled by a new `odo component status -o json` command, described here.
- 3 is handled by the existing unmodified `odo log --follow` command.

The new proposed `odo component status -o json` command will:
- Be a *long-running* command that will continue outputing status until it is aborted by the parent process.
- Every X seconds, send an HTTP/S request to the URLs/routes of the application as they existed when the command was first executed. Output the result as a JSON string.
- Every X seconds (or using a Kubernetes watch, where appropriate), check the container status for the application, based on the application data that was present when the command was first issued. Output the result as a JSON string.

**Note**: This command will NOT, by design, respond to route/application changes that occur during or after it is first invoked. It is up to consuming tools to ensure that the `odo component status` command is stopped/restarted as needed. 
  - For example, if the user tells the IDE to delete their application with the IDE UI, the IDE will call `odo delete (component)`; at the same time, the IDE should also abort any existing `odo component status` commands that are running (as these are no longer guaranteed to return a valid status now that the application itself no longer exists). `odo component status` will not automatically abort when the application is deleted (because it has no reliable way to detect this in all cases).
  - Another example: if the IDE adds a new URL via `odo url create [...]`, any existing `odo component status` commands that are running should be aborted, as these commands would still only be checking the URLs that existed when the command was first invoked (eg there is intentionally no cross-process notification mechanism for created/updated/deleted URLs implemented as part of this command.)

This is an example an `odo component status -o json` command invocation look like:
```
odo component status -o json

{ "componentURLStatus" : { "url" : "https://(...)", "response" : "true", "responseCode" : 200, "timestamp" : (UTC unix epoch seconds.microseconds) } }
{ "componentURLStatus" : { "url" : "https://(...)", "response" : "false", error: "host unreachable", "timestamp" : (...) } }
{ "containerStatus" : { "status" : "containercreating", "timestamp" : (...)} }
{ "containerStatus" : { "status" : "running", "timestamp" : (...)} }
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

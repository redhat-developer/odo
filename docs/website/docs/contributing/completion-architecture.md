---
title:  Completion Architecture
sidebar_position: 8
---

odo provides smart completion of command names, arguments and flags when typing commands. The completion architecture allows for completion handlers (i.e. pieces of code that provide the suggestions to the shell running the command) to be written entirely in go. This document describes the architecture, how to develop new completion handlers and how to activate/deactivate completions for odo.

### Architecture
The completion architecture relies on the [posener/complete](https://github.com/posener/complete) project. _posener/complete_ relies on providing implementations of their `Predictor` interface:
```go
type Predictor interface {
    Predict(Args) []string
}
```

These implementations are then bound to a `complete.Command` which describes how commands are supposed to be completed by the shell.

While it is possible to create an external application simply for the purpose of providing completions, it made more sense to have odo itself deal with exposing possible completions. In this case, we need to hook into the `cobra` command lifecycle and expose completion information before `cobra` takes over. This happens in the main entry point for the odo application. Completion information is provided by the `createCompletion` function which walks the `cobra` command tree and creates the `complete.Command` tree as it goes along, attaching completion handler to commands (for arguments) and flags.

In order to provide this information, odo allows commands to register completion handlers for their arguments and flags using `completion.RegisterCommandHandler` and `completion.RegisterCommandFlagHandler` respectively. These functions will adapt the contextualized predictor that we expose to commands to the _posener/complete_ `Predictor` interface internally.

### Create a new completion handler

[//]: # (Add more information to this.)

`completion.ContextualizedPredictor` is a function which should return an array of possible completion strings based on the given arguments. It is defined as:
```go
type ContextualizedPredictor func(cmd *cobra.Command, args parsedArgs, context *genericclioptions.Context) []string
```

This function should be put in [pkg/odo/util/completion/completionhandlers.go](https://github.com/openshift/odo/blob/main/pkg/odo/util/completion/completionhandlers.go) so that it can be reused across commands.

While filtering is done by `posener/complete` itself, automatically removing all completions not prefixed by what you’ve typed already, it might be useful to use the values provided by `parsedArgs` to help optimize things to avoid repeating possible completions for example.

### Register the completion handler
The command should register the appropriate completion handler.
- For argument completion handler: 
  ```go
  completion.RegisterCommandHandler(command *cobra.Command, predictor ContextualizedPredictor)
  ```

- For flag completion handler:  
  ```go
  completion.RegisterCommandFlagHandler(command *cobra.Command, flag string, predictor ContextualizedPredictor)
  ```

Registering the completion handler will make it available for `main.createCompletion` which will then automatically create the completion information from it.

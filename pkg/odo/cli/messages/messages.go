// Package messages contains the various "outputs" that we use in both the CLI and test cases.
package messages

const InteractiveModeEnabled = "Interactive mode enabled, please answer the following questions:"
const InitializingNewComponent = "Initializing a new component"
const SourceCodeDetected = "Files: Source code detected, a Devfile will be determined based upon source code autodetection"
const NoSourceCodeDetected = "Files: No source code detected, a starter project will be created in the current directory"
const DevInitializeExistingComponent = "Dev mode ran, but no Devfile was found. Initializing a component in the current directory"
const DeployInitializeExistingComponent = "Deploy mode ran, but no Devfile was found. Initializing a component in the current directory"

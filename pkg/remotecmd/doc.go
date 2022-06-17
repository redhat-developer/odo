// Package remotecmd manages commands that are intended to be executed remotely, in Kubernetes/OpenShift containers.
// It essentially provides a generic interface allowing to manage processes spawned for executing commands defined in
// Devfiles.
// It also provides package-level functions to execute any command in a given container in a given pod.
package remotecmd

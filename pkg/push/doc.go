// Package push supports the odo push operation.
//
// Pusher walks the local files and sends actions (copy or remove)
// on a channel.
//
// Remote sets up the remote 'tar' and 'rm' commands and executes
// actions from the Pusher.
//
// Pusher and Remote both start working as soon as they are created
// but no data is sent till Pusher.Push(Remote). That means they can
// get started early and work concurrently which reduces overall
// delays.
//
package push

package plugins

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
)

// HandleCommand receives a PluginHandler and command-line arguments and attempts to find
// a plugin executable on the PATH that satisfies the given arguments.
func HandleCommand(handler PluginHandler, args []string) error {
	foundBinary, remaining := findBinary(handler, args)
	if foundBinary == "" {
		return nil
	}

	if err := handler.Execute(foundBinary, args[len(remaining):], os.Environ()); err != nil {
		return err
	}
	return nil
}

type execFunc func(string, []string, []string) (err error)

// NewExecHandler creates and returns a new ExecHandler configured with
// the prefix.
func NewExecHandler(prefix string) *ExecHandler {
	return &ExecHandler{
		Prefix: prefix,
		Exec:   syscall.Exec,
	}
}

// PluginHandler provides functionality for finding and executing external
// plugins.
type PluginHandler interface {
	// Lookup should return the full path to an executable for the provided
	// command, or "" if no matching command can be found.
	Lookup(command string) string

	// Execute should execute the provided path, passing in the args and env.
	Execute(filename string, args, env []string) error
}

// ExecHandler implements PluginHandler using the "os/exec" package.
type ExecHandler struct {
	Prefix string
	Exec   execFunc
}

var _ PluginHandler = (*ExecHandler)(nil)

// Lookup implements PluginHandler, using
// https://golang.org/pkg/os/exec/#LookPath to search for the command.
func (h *ExecHandler) Lookup(command string) string {
	if runtime.GOOS == "windows" {
		command = command + ".exe"
	}
	path, err := exec.LookPath(fmt.Sprintf("%s-%s", h.Prefix, command))
	if err == nil && len(path) != 0 {
		return path
	}
	return ""
}

// Execute implements PluginHandler.Execute
func (h *ExecHandler) Execute(filename string, args, env []string) error {
	// Windows does not support exec syscall.
	if runtime.GOOS == "windows" {
		cmd := exec.Command(filename, args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.Env = env
		err := cmd.Run()
		if err == nil {
			os.Exit(0)
		}
		return err
	}
	return h.Exec(filename, append([]string{filename}, args...), env)
}

func findBinary(handler PluginHandler, args []string) (string, []string) {
	found := ""
	remaining := []string{}
	for idx := range args {
		if strings.HasPrefix(args[idx], "-") {
			break
		}
		remaining = append(remaining, strings.Replace(args[idx], "-", "_", -1))
	}
	for len(remaining) > 0 {
		path := handler.Lookup(strings.Join(remaining, "-"))
		if path == "" {
			remaining = remaining[:len(remaining)-1]
			continue
		}

		found = path
		break
	}
	return found, remaining
}

package completion

import (
	"github.com/posener/complete"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
)

type completionHandler struct {
	cmd       *cobra.Command
	ctxLoader contextLoader
	predictor ContextualizedPredictor
}

type handlerKey struct {
	cmd  *cobra.Command
	flag string
}

type contextLoader func(command *cobra.Command) *genericclioptions.Context

// ContextualizedPredictor predicts completion based on specified arguments, potentially using the context provided by the
// specified client to resolve the entities to be completed
type ContextualizedPredictor func(cmd *cobra.Command, args complete.Args, context *genericclioptions.Context) []string

func (ch completionHandler) Predict(args complete.Args) []string {
	return ch.predictor(ch.cmd, args, ch.ctxLoader(ch.cmd))
}

// completionHandlers records available completion handlers for commands and flags
var completionHandlers = make(map[handlerKey]completionHandler)

// getCommandCompletionHandlerKey retrieves the completion handler identifier associated with the specified command. The associated
// handler should provide completions for valid values for the specified command's arguments.
func getCommandCompletionHandlerKey(command *cobra.Command) handlerKey {
	return handlerKey{
		cmd: command,
	}
}

// getCommandFlagCompletionHandlerKey retrieves the completion handler identifier associated with the specified command and flag name. The
// associated handler should provide completion for valid values for the specified command's flag.
func getCommandFlagCompletionHandlerKey(command *cobra.Command, flag string) handlerKey {
	return handlerKey{
		cmd:  command,
		flag: flag,
	}
}

func newHandler(cmd *cobra.Command, predictor ContextualizedPredictor) completionHandler {
	return completionHandler{
		cmd:       cmd,
		ctxLoader: genericclioptions.NewContext,
		predictor: predictor,
	}
}

// RegisterCommandHandler registers the provided ContextualizedPredictor as a completion handler for the specified command
func RegisterCommandHandler(command *cobra.Command, predictor ContextualizedPredictor) {
	completionHandlers[getCommandCompletionHandlerKey(command)] = newHandler(command, predictor)
}

// RegisterCommandFlagHandler registers the provided ContextualizedPredictor as a completion handler for the specified flag
// of the specified command
func RegisterCommandFlagHandler(command *cobra.Command, flag string, predictor ContextualizedPredictor) {
	completionHandlers[getCommandFlagCompletionHandlerKey(command, flag)] = newHandler(command, predictor)
}

// GetCommandHandler retrieves the command handler associated with the specified command or nil otherwise
func GetCommandHandler(command *cobra.Command) (predictor complete.Predictor, ok bool) {
	predictor, ok = completionHandlers[getCommandCompletionHandlerKey(command)]
	return
}

// GetCommandFlagHandler retrieves the command handler associated with the specified flag of the specified command or nil otherwise
func GetCommandFlagHandler(command *cobra.Command, flag string) (predictor complete.Predictor, ok bool) {
	predictor, ok = completionHandlers[getCommandFlagCompletionHandlerKey(command, flag)]
	return
}

// getCommandsAndFlags returns the commands and flags from the given input
func getCommandsAndFlags(args []string, c *cobra.Command) (map[string]bool, map[string]string) {
	strippedCommandsMap := make(map[string]bool)
	setFlags := make(map[string]string)

	if len(args) == 0 {
		return strippedCommandsMap, setFlags
	}
	err := c.ParseFlags(args)
	if err != nil {
		return strippedCommandsMap, setFlags
	}

	flags := c.Flags()

	cmds := flags.Args()
	flags.Visit(func(i *flag.Flag) {
		if i.Value.Type() != "bool" {
			setFlags[i.Name] = i.Value.String()
		}
	})

	// send a map of commands for faster searching
	for _, strippedCommand := range cmds {
		strippedCommandsMap[strippedCommand] = true
	}

	return strippedCommandsMap, setFlags
}

// getUserTypedCommands returns only the user typed entities by excluding the cobra predefined commands
func getUserTypedCommands(args complete.Args, command *cobra.Command) []string {
	var commands []string

	// get only the user typed commands/flags and remove the cobra defined commands
	found := false
	for _, arg := range args.Completed {
		if arg == command.Name() && !found {
			found = true
			continue
		}
		if found {
			commands = append(commands, arg)
		}
	}

	return commands
}

package completion

import (
	"github.com/posener/complete"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
)

// completionHandler wraps a ContextualizedPredictor providing needed information for its invocation during the Predict function
type completionHandler struct {
	cmd       *cobra.Command
	ctxLoader contextLoader
	predictor ContextualizedPredictor
}

// handlerKey provides a key value to record and identify completion handlers
type handlerKey struct {
	cmd  *cobra.Command
	flag string
}

// parsedArgs provides easier to deal with information about what the command line looks like during a completion call
type parsedArgs struct {
	// original records the original arguments provided by posener/complete
	original complete.Args
	// typed returns what the user typed minus the command triggering the completion
	typed []string
	// commands lists parsed commands from the typed portion of the command line
	commands map[string]bool
	// flagValues provides a map associating parsed flag name and its value as string
	flagValues map[string]string
}

type contextLoader func(command *cobra.Command) *genericclioptions.Context

// ContextualizedPredictor predicts completion based on specified arguments, potentially using the context provided by the
// specified client to resolve the entities to be completed
type ContextualizedPredictor func(cmd *cobra.Command, args parsedArgs, context *genericclioptions.Context) []string

// Predict is called by the posener/complete code when the shell asks for completion of a given argument
func (ch completionHandler) Predict(args complete.Args) []string {
	return ch.predictor(ch.cmd, NewParsedArgs(args, ch.cmd), ch.ctxLoader(ch.cmd))
}

// NewParsedArgs creates a parsed representation of the provided arguments for the specified command. Mostly exposed for tests.
func NewParsedArgs(args complete.Args, cmd *cobra.Command) parsedArgs {
	typed := getUserTypedCommands(args, cmd)
	commands, flagValues := getCommandsAndFlags(typed, cmd)

	complete.Log("Parsed commands values for full input: %v", commands)
	complete.Log("Parsed flag values for full input: %v", flagValues)

	// the parsing above won't work properly when we have previously supplied flags and we are trying to complete
	// another flag
	// For example when the user has typed "odo service create --plan prod -p "
	// the flagValues var above would not contains the plan flag and value
	// So in the following part of the code we strip the last thing the user added and parse again
	// any new parsed flags will be added
	// There is no need to do this extra parsing if the user has not already added a flag. So as a heuristic
	// we only perform this parsing when at the user has added at least 2 tokens
	if len(typed) > 2 {
		typedWithoutLast := make([]string, len(typed))
		copy(typedWithoutLast, typed)
		typedWithoutLast = typedWithoutLast[:len(typedWithoutLast)-1]
		_, flagValuesWithoutLast := getCommandsAndFlags(typedWithoutLast, cmd)

		complete.Log("Parsed flag values for input without last: %v", flagValuesWithoutLast)

		// now we merge the two maps
		for k, v := range flagValuesWithoutLast {
			flagValues[k] = v
		}
	}

	parsed := parsedArgs{
		original:   args,
		typed:      typed,
		commands:   commands,
		flagValues: flagValues,
	}
	_ = cmd.ParseFlags(typed)

	return parsed
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

// newHandler wraps a ContextualizedPredictor into a completionHandler
func newHandler(cmd *cobra.Command, predictor ContextualizedPredictor) completionHandler {
	return completionHandler{
		cmd:       cmd,
		ctxLoader: genericclioptions.NewContextCompletion,
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
		setFlags[i.Name] = i.Value.String()
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

package completion

import (
	"github.com/posener/complete"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"strings"
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

func getCommandsAndFlags(args []string, c *cobra.Command) ([]string, map[string]string) {
	if len(args) == 0 {
		return args, nil
	}
	c.ParseFlags(args)
	flags := c.Flags()

	var strippedCommands []string
	strippedFlags := make(map[string]string)

	for i := 0; i < len(args); i++ {
		currentArgs := args[i]

		nextArgs := ""
		if i < len(args)-1 {
			nextArgs = args[i+1]
		}

		if strings.HasPrefix(currentArgs, "--") && !strings.Contains(currentArgs, "=") && !isBooleanFlag(currentArgs[2:], flags, false) {
			// set and increment because the next word is the value of the flag
			strippedFlags[currentArgs[2:]] = nextArgs
			i++
		} else if len(currentArgs) > 1 && strings.HasPrefix(currentArgs, "-") && !strings.Contains(currentArgs, "=") && !isBooleanFlag(currentArgs[2:], flags, true) {
			// set and increment because the next word is the value of the flag
			strippedFlags[currentArgs[1:]] = nextArgs
			i++
		} else if !strings.HasPrefix(currentArgs, "-") && !strings.HasPrefix(currentArgs, "--") {
			// it's not a flag or a shorthand flag, so append
			strippedCommands = append(strippedCommands, currentArgs)
		}
	}
	return strippedCommands, strippedFlags
}

func isBooleanFlag(name string, fs *flag.FlagSet, isShort bool) bool {
	var flag *flag.Flag
	if isShort {
		flag = fs.ShorthandLookup(name)
	} else {
		flag = fs.Lookup(name)
	}
	if flag == nil || flag.Value != nil {
		return false
	}
	return flag.Value.Type() == "bool"
}

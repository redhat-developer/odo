package completion

import (
	"github.com/posener/complete"
	"github.com/redhat-developer/odo/pkg/occlient"
	"github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/spf13/cobra"
)

type completionHandler struct {
	client    clientLoader
	predictor ContextualizedPredictor
}

type handlerKey struct {
	cmd  *cobra.Command
	flag string
}

type clientLoader func() *occlient.Client

// ContextualizedPredictor predicts completion based on specified arguments, potentially using the context provided by the
// specified client to resolve the entities to be completed
type ContextualizedPredictor func(args complete.Args, client *occlient.Client) []string

func (ch completionHandler) Predict(args complete.Args) []string {
	return ch.predictor(args, ch.client())
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

func newHandler(predictor ContextualizedPredictor) completionHandler {
	return completionHandler{
		client:    util.GetOcClient,
		predictor: predictor,
	}
}

// RegisterCommandHandler registers the provided ContextualizedPredictor as a completion handler for the specified command
func RegisterCommandHandler(command *cobra.Command, predictor ContextualizedPredictor) {
	completionHandlers[getCommandCompletionHandlerKey(command)] = newHandler(predictor)
}

// RegisterCommandFlagHandler registers the provided ContextualizedPredictor as a completion handler for the specified flag
// of the specified command
func RegisterCommandFlagHandler(command *cobra.Command, flag string, predictor ContextualizedPredictor) {
	completionHandlers[getCommandFlagCompletionHandlerKey(command, flag)] = newHandler(predictor)
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

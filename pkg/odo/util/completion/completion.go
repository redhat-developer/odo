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

type clientLoader func() *occlient.Client

// ContextualizedPredictor predicts completion based on specified arguments, potentially using the context provided by the
// specified client to resolve the entities to be completed
type ContextualizedPredictor func(args complete.Args, client *occlient.Client) []string

func (ch completionHandler) Predict(args complete.Args) []string {
	return ch.predictor(args, ch.client())
}

// completionHandlers records available completion handlers for commands and flags
var completionHandlers = make(map[string]completionHandler)

// getCommandSuggesterName retrieves the completion handler identifier associated with the specified command. The associated
// handler should provide completions for valid values for the specified command's arguments.
func getCommandSuggesterName(command *cobra.Command) string {
	return getCommandSuggesterNameFrom(command.Name())
}

func getCommandSuggesterNameFrom(commandName string) string {
	return commandName
}

// getFlagSuggesterName retrieves the completion handler identifier associated with the specified command and flag name. The
// associated handler should provide completion for valid values for the specified command's flag.
func getFlagSuggesterName(command *cobra.Command, flag string) string {
	return getCommandSuggesterNameFrom(command.Name()) + "_" + flag
}

func newHandler(predictor ContextualizedPredictor) completionHandler {
	return completionHandler{
		client:    util.GetOcClient,
		predictor: predictor,
	}
}

// RegisterCommandHandler registers the provided ContextualizedPredictor as a completion handler for the specified command
func RegisterCommandHandler(command *cobra.Command, predictor ContextualizedPredictor) {
	completionHandlers[getCommandSuggesterName(command)] = newHandler(predictor)
}

// RegisterCommandFlagHandler registers the provided ContextualizedPredictor as a completion handler for the specified flag
// of the specified command
func RegisterCommandFlagHandler(command *cobra.Command, flag string, predictor ContextualizedPredictor) {
	completionHandlers[getFlagSuggesterName(command, flag)] = newHandler(predictor)
}

// GetCommandHandler retrieves the command handler associated with the specified command or nil otherwise
func GetCommandHandler(command *cobra.Command) (predictor complete.Predictor, ok bool) {
	predictor, ok = completionHandlers[getCommandSuggesterName(command)]
	return
}

// GetCommandFlagHandler retrieves the command handler associated with the specified flag of the specified command or nil otherwise
func GetCommandFlagHandler(command *cobra.Command, flag string) (predictor complete.Predictor, ok bool) {
	predictor, ok = completionHandlers[getFlagSuggesterName(command, flag)]
	return
}

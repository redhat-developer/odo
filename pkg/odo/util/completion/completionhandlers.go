package completion

import (
	"github.com/posener/complete"
	"github.com/spf13/cobra"

	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
)

// FileCompletionHandler provides suggestions for files and directories
var FileCompletionHandler = func(cmd *cobra.Command, args parsedArgs, context *genericclioptions.Context) (completions []string) {
	completions = append(completions, complete.PredictFiles("*").Predict(args.original)...)
	return
}

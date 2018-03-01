package cmd

import (
	"fmt"
	"os"

	"github.com/redhat-developer/ocdev/pkg/notify"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Global variables
var (
	GlobalVerbose bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ocdev",
	Short: "OpenShift CLI for Developers",
	Long:  `-`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {

		// Add extra logging when verbosity is passed
		if GlobalVerbose {
			//TODO
			log.SetLevel(log.DebugLevel)
		}

	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {

	updateInfo := make(chan string)
	go getLatestReleaseInfo(updateInfo)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	select {
	case message := <-updateInfo:
		fmt.Printf(message)
	default:
		log.Debug("Could not get the latest release information in time. Never mind, exiting gracefully :)")
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.ocdev.yaml)")

	rootCmd.PersistentFlags().BoolVarP(&GlobalVerbose, "verbose", "v", false, "verbose output")
}

func getLatestReleaseInfo(info chan<- string) {
	newTag, err := notify.CheckLatestReleaseTag(VERSION)
	if err != nil {
		// The error is intentionally not being handled because we don't want
		// to stop the execution of the program because of this failure
		log.Debugf("Error checking if newer ocdev release is available: %v", err)
	}
	if len(newTag) > 0 {
		info <- "---\n" +
			"A newer version of ocdev (version: " + fmt.Sprint(newTag) + ") is available.\n" +
			"Update using your package manager, or run\n" +
			"curl " + notify.InstallScriptURL + " | sh\n" +
			"to update manually, or visit https://github.com/redhat-developer/ocdev/releases\n" +
			"---\n"
	}
}

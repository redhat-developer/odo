package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joshdk/go-junit"
	"github.com/redhat-developer/odo/junit-collector/pkg/db"
)

func usage(msg string) {
	fmt.Printf("%s\n\nUsage: %s --sheetId <spreadsheetId> --junit <junit-file> --pr <pr-number> --test <test title> --logfile <logfile>\n", msg, os.Args[0])
	os.Exit(1)
}

/*
GOOGLE_APPLICATION_CREDENTIALS must point to an existing GCP JSON Service account file.
The service account does not need any extra role.
The service account must have an Editor access to the Sheet (use the Share button on the Sheet UI to add this permission)
*/
func main() {
	var (
		fSheetId   = flag.String("sheetId", "", "spreadsheetId")
		fJunitFile = flag.String("junit", "", "junit file")
		fPrNumber  = flag.String("pr", "", "PR number")
		fTestTile  = flag.String("test", "", "Test title")
		fLogFile   = flag.String("logfile", "", "Log file including base")
	)

	flag.Parse()

	if *fSheetId == "" {
		usage("--sheetId is missing")
	}
	spreadsheetId := *fSheetId

	if *fJunitFile == "" {
		usage("--junit is missing")
	}
	junitFile := *fJunitFile

	if *fPrNumber == "" {
		usage("--pr is missing")
	}
	prNumber := *fPrNumber

	if *fTestTile == "" {
		usage("--test is missing")
	}
	testTitle := *fTestTile

	if *fLogFile == "" {
		usage("--logfile is missing")
	}
	logFile := *fLogFile

	ctx := context.Background()

	suites, err := junit.IngestFile(junitFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fmt.Printf("junit file %s not found. Exiting\n", junitFile)
			return
		}
		panic(err)
	}

	for _, suite := range suites {
		for _, test := range suite.Tests {
			if test.Error == nil {
				continue
			}
			currentTime := time.Now()
			data := []interface{}{
				// Date must be set in this format, to be correctly comparable as strings
				currentTime.Format("2006-01-02"),
				test.Name,
				test.Error.Error(),
				prNumber,
				testTitle,
				logFile,
			}
			err = db.SaveToSheet(ctx, spreadsheetId, data)
			if err != nil {
				log.Fatalf("Unable to add data to sheet: %v", err)
			}
		}
	}
}

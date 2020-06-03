package utils

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/openshift/odo/pkg/machineoutput"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// AnalyzePushConsoleOutput analyzes the output of 'odo push -o json' for the machine readable event push test above.
func AnalyzePushConsoleOutput(pushConsoleOutput string) {

	lines := strings.Split(strings.Replace(pushConsoleOutput, "\r\n", "\n", -1), "\n")

	var entries []machineoutput.MachineEventLogEntry

	// Ensure that all lines can be correctly parsed into their expected JSON structure
	for _, line := range lines {

		if len(strings.TrimSpace(line)) == 0 {
			continue
		}

		// fmt.Println("Processing output line: " + line)

		lineWrapper := machineoutput.MachineEventWrapper{}

		err := json.Unmarshal([]byte(line), &lineWrapper)
		Expect(err).NotTo(HaveOccurred())

		entry, err := lineWrapper.GetEntry()
		Expect(err).NotTo(HaveOccurred())

		entries = append(entries, entry)

	}

	if len(entries) < 4 {
		Fail("Expected at least 4 entries, corresponding to command/action execution.")
	}

	// Timestamps should be monotonically increasing
	mostRecentTimestamp := float64(-1)
	for _, entry := range entries {
		timestamp, err := strconv.ParseFloat(entry.GetTimestamp(), 64)
		Expect(err).NotTo(HaveOccurred())

		if timestamp < mostRecentTimestamp {
			Fail("Timestamp was not monotonically increasing " + entry.GetTimestamp() + " " + strconv.FormatFloat(mostRecentTimestamp, 'E', -1, 64))
		}

		mostRecentTimestamp = timestamp
	}

	// First look for the expected devbuild events, then look for the expected devrun events.
	expectedEventOrder := []struct {
		entryType   machineoutput.MachineEventLogEntryType
		commandName string
	}{
		// first the devbuild command (and its action) should run
		{
			machineoutput.TypeDevFileCommandExecutionBegin,
			"devbuild",
		},
		// {
		// 	machineoutput.TypeDevFileActionExecutionBegin,
		// 	"devbuild",
		// },
		{
			// at least one logged line of text
			machineoutput.TypeLogText,
			"",
		},
		// {
		// 	machineoutput.TypeDevFileActionExecutionComplete,
		// 	"devbuild",
		// },
		{
			machineoutput.TypeDevFileCommandExecutionComplete,
			"devbuild",
		},
		// next the devbuild command (and its action) should run
		{
			machineoutput.TypeDevFileCommandExecutionBegin,
			"devrun",
		},
		// ,
		// {
		// 	machineoutput.TypeDevFileActionExecutionBegin,
		// 	"devrun",
		// },
		{
			// at least one logged line of text
			machineoutput.TypeLogText,
			"",
		},
		// ,
		// {
		// 	machineoutput.TypeDevFileActionExecutionComplete,
		// 	"devrun",
		// },
		{
			machineoutput.TypeDevFileCommandExecutionComplete,
			"devrun",
		},
	}

	currIndex := -1
	for _, nextEventOrder := range expectedEventOrder {
		entry, newIndex := machineoutput.FindNextEntryByType(currIndex, nextEventOrder.entryType, entries)
		Expect(entry).NotTo(BeNil())
		Expect(newIndex).To(BeNumerically(">=", 0))
		Expect(newIndex).To(BeNumerically(">", currIndex)) // monotonically increasing index

		// We should see devbuild for the first set of events, then devrun
		commandName := machineoutput.GetCommandName(entry)
		Expect(commandName).To(Equal(nextEventOrder.commandName))

		currIndex = newIndex
	}

}

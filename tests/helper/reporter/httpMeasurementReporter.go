package reporter

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/types"

	. "github.com/onsi/ginkgo"
)

type HTTPMeasurementReporter struct {
	url string
}

func NewHTTPMeasurementReporter(url string) *HTTPMeasurementReporter {
	r := HTTPMeasurementReporter{
		url: url,
	}
	return &r
}

func (r *HTTPMeasurementReporter) SpecSuiteWillBegin(config config.GinkgoConfigType, summary *types.SuiteSummary) {
}

func (r *HTTPMeasurementReporter) BeforeSuiteDidRun(setupSummary *types.SetupSummary) {}

func (r *HTTPMeasurementReporter) SpecWillRun(specSummary *types.SpecSummary) {}

func (r *HTTPMeasurementReporter) SpecDidComplete(specSummary *types.SpecSummary) {
	if specSummary.Passed() && specSummary.IsMeasurement {

		var pr int
		pr, err := getPRNumber()
		if err != nil {
			fmt.Fprintf(GinkgoWriter, "WARNING: unable to get PR number (%v)\n", err)
		}

		for k, v := range specSummary.Measurements {
			output := map[string]string{}
			output["Number of Samples"] = strconv.Itoa(specSummary.NumberOfSamples)
			output["Measurement"] = k
			output["PR"] = strconv.Itoa(pr)
			output[v.SmallestLabel] = strconv.FormatFloat(v.Smallest, 'f', -1, 64)
			output[v.LargestLabel] = strconv.FormatFloat(v.Largest, 'f', -1, 64)
			output[v.AverageLabel] = strconv.FormatFloat(v.Average, 'f', -1, 64)
			output["Test"] = strings.Join(specSummary.ComponentTexts, "/")
			err := r.SubmitMeasurement(output)
			if err != nil {
				// Just printing info about error. Error while submiting measurement should cause any failures
				fmt.Fprintf(GinkgoWriter, "WARNING: error in SubmitMeasurement (%v)\n", err)
			}
		}
	}
}

func (r *HTTPMeasurementReporter) AfterSuiteDidRun(setupSummary *types.SetupSummary) {}

func (r *HTTPMeasurementReporter) SpecSuiteDidEnd(summary *types.SuiteSummary) {}

func (r *HTTPMeasurementReporter) SubmitMeasurement(data map[string]string) error {
	client := &http.Client{}

	req, err := http.NewRequest("GET", r.url, nil)
	if err != nil {
		return err
	}

	q := req.URL.Query()
	for k, v := range data {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("error while submiting measurement (StatusCode: %d)", resp.StatusCode)
	}

	return nil

}

// getPrNumber returns PR number from json in CLONEREFS_OPTIONS environment variable
// this env variable is specific to Prow jobs
func getPRNumber() (int, error) {
	jsonData := os.Getenv("CLONEREFS_OPTIONS")

	type pulls struct {
		Number int `json:"number"`
	}

	type refs struct {
		BaseSha string  `json:"base_sha"`
		Pulls   []pulls `json:"pulls"`
	}

	type clonerefsOptions struct {
		Refs    []refs `json:"refs"`
		SrcRoot string `json:"src_root"`
	}

	var data clonerefsOptions
	err := json.Unmarshal([]byte(jsonData), &data)
	if err != nil {
		return 0, fmt.Errorf("error in unmarshalling json")
	}

	if len(data.Refs) < 1 {
		return 0, fmt.Errorf("no refs in the input json")
	}

	if len(data.Refs[0].Pulls) < 1 {
		return 0, fmt.Errorf("no refs[0].pulls in the input json")
	}

	return data.Refs[0].Pulls[0].Number, nil

}

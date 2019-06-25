package reporter

import (
	"fmt"
	"net/http"
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
		for k, v := range specSummary.Measurements {
			output := map[string]string{}
			output["Number of Samples"] = strconv.Itoa(specSummary.NumberOfSamples)
			output["Measurement"] = k
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

package ui

import (
	"fmt"
	"github.com/Netflix/go-expect"
	"github.com/hinshun/vt10x"
	scv1beta1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/redhat-developer/odo/pkg/testingutil"
	"github.com/stretchr/testify/require"
	"gopkg.in/AlecAivazis/survey.v1/core"
	"gopkg.in/AlecAivazis/survey.v1/terminal"
	"k8s.io/apimachinery/pkg/runtime"
	"testing"
)

func stdio(c *expect.Console) terminal.Stdio {
	return terminal.Stdio{In: c.Tty(), Out: c.Tty(), Err: c.Tty()}
}

func init() {
	// disable color output for all prompts to simplify testing
	core.DisableColor = true
}

func TestEnterServicePropertiesInteractively(t *testing.T) {
	planExternalMetaDataRaw, err := testingutil.FakePlanExternalMetaDataRaw()
	if err != nil {
		fmt.Printf("error occured %v during marshalling", err)
		return
	}

	planServiceInstanceCreateParameterSchemasRaw, err := testingutil.FakePlanServiceInstanceCreateParameterSchemasRaw()
	if err != nil {
		fmt.Printf("error occured %v during marshalling", err)
		return
	}

	tests := []struct {
		name           string
		servicePlan    scv1beta1.ClusterServicePlan
		expectedValues map[string]string
	}{
		{
			name: "test 1 : with correct values",
			servicePlan: scv1beta1.ClusterServicePlan{
				Spec: scv1beta1.ClusterServicePlanSpec{
					ClusterServiceClassRef: scv1beta1.ClusterObjectReference{
						Name: "1dda1477cace09730bd8ed7a6505607e",
					},
					CommonServicePlanSpec: scv1beta1.CommonServicePlanSpec{
						ExternalName:                         "dev",
						Description:                          "this is a example description 1",
						ExternalMetadata:                     &runtime.RawExtension{Raw: planExternalMetaDataRaw[0]},
						ServiceInstanceCreateParameterSchema: &runtime.RawExtension{Raw: planServiceInstanceCreateParameterSchemasRaw[0]},
					},
				},
			},
			expectedValues: map[string]string{
				"PLAN_DATABASE_URI":      "someuri",
				"PLAN_DATABASE_USERNAME": "default",
				"PLAN_DATABASE_PASSWORD": "foo",
			},
		},
	}

	for _, tt := range tests {
		plan := tt.servicePlan

		// Multiplex stdin/stdout to a virtual terminal to respond to ANSI escape
		// sequences (i.e. cursor position report).
		c, state, err := vt10x.NewVT10XConsole()
		require.Nil(t, err)
		defer c.Close()

		donec := make(chan struct{})
		go func() {
			defer close(donec)
			c.ExpectString("Enter a value for string property PLAN_DATABASE_PASSWORD:")
			c.SendLine("foo")
			c.ExpectString("Enter a value for string property PLAN_DATABASE_URI:")
			c.SendLine("")
			c.ExpectString("Enter a value for string property PLAN_DATABASE_USERNAME:")
			c.SendLine("")
			c.ExpectString("Provide values for non-required properties")
			c.SendLine("")
			c.ExpectEOF()
		}()

		values := enterServicePropertiesInteractively(plan, stdio(c))

		require.Equal(t, tt.expectedValues, values)

		// Close the slave end of the pty, and read the remaining bytes from the master end.
		c.Tty().Close()
		<-donec

		// Dump the terminal's screen.
		t.Log(expect.StripTrailingEmptyLines(state.String()))
	}
}

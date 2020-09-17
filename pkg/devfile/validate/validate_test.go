package validate

import (
	"testing"

	"github.com/openshift/odo/pkg/devfile/parser/data"
	versionsCommon "github.com/openshift/odo/pkg/devfile/parser/data/common"
	"github.com/openshift/odo/pkg/testingutil"
)

func TestValidateContatinerName(t *testing.T) {

	tests := []struct {
		name        string
		devfileData data.DevfileData
		wantErr     bool
	}{
		{
			name: "Case 1: Valid container name",
			devfileData: &testingutil.TestDevfileData{
				Components: []versionsCommon.DevfileComponent{
					{
						Name: "runtime",
						Container: &versionsCommon.Container{
							Image: "quay.io/nodejs-12",
							Endpoints: []versionsCommon.Endpoint{
								{
									Name:       "port-3000",
									TargetPort: 3000,
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Case 2: long container name",
			devfileData: &testingutil.TestDevfileData{
				Components: []versionsCommon.DevfileComponent{
					{
						Name: "runtimeruntimeruntimeruntimeruntimeruntimeruntimeruntimeruntimeruntime",
						Container: &versionsCommon.Container{
							Image: "quay.io/nodejs-12",
							Endpoints: []versionsCommon.Endpoint{
								{
									Name:       "port-3000",
									TargetPort: 3000,
								},
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Case 3: special character in container name",
			devfileData: &testingutil.TestDevfileData{
				Components: []versionsCommon.DevfileComponent{
					{
						Name: "run@time",
						Container: &versionsCommon.Container{
							Image: "quay.io/nodejs-12",
							Endpoints: []versionsCommon.Endpoint{
								{
									Name:       "port-3000",
									TargetPort: 3000,
								},
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Case 4: numeric container name",
			devfileData: &testingutil.TestDevfileData{
				Components: []versionsCommon.DevfileComponent{
					{
						Name: "12345",
						Container: &versionsCommon.Container{
							Image: "quay.io/nodejs-12",
							Endpoints: []versionsCommon.Endpoint{
								{
									Name:       "port-3000",
									TargetPort: 3000,
								},
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Case 5: container name with capitalised character",
			devfileData: &testingutil.TestDevfileData{
				Components: []versionsCommon.DevfileComponent{
					{
						Name: "runTime",
						Container: &versionsCommon.Container{
							Image: "quay.io/nodejs-12",
							Endpoints: []versionsCommon.Endpoint{
								{
									Name:       "port-3000",
									TargetPort: 3000,
								},
							},
						},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateContatinerName(tt.devfileData)
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

		})
	}

}

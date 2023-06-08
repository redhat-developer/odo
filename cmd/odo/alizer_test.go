package main

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/redhat-developer/odo/pkg/alizer"
	"github.com/redhat-developer/odo/pkg/config"
	envcontext "github.com/redhat-developer/odo/pkg/config/context"
	"github.com/redhat-developer/odo/pkg/odo/cli"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
)

func TestOdoAlizer(t *testing.T) {

	for _, tt := range []struct {
		name       string
		clientset  func() clientset.Clientset
		args       []string
		wantErr    string
		wantStdout string
		wantStderr string
	}{
		{
			name: "analyze without json output",
			clientset: func() clientset.Clientset {
				return clientset.Clientset{}
			},
			args:    []string{"analyze"},
			wantErr: "this command can be run with json output only, please use the flag: -o json",
		},
		{
			name: "analyze with json output",
			clientset: func() clientset.Clientset {
				ctrl := gomock.NewController(t)
				fs := filesystem.NewFakeFs()
				alizerClient := alizer.NewMockClient(ctrl)
				alizerClient.EXPECT().DetectFramework(gomock.Any(), gomock.Any())
				alizerClient.EXPECT().DetectPorts(gomock.Any())
				alizerClient.EXPECT().DetectName(gomock.Any())
				return clientset.Clientset{
					FS:           fs,
					AlizerClient: alizerClient,
				}
			},
			args: []string{"analyze", "-o", "json"},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			envConfig, err := config.GetConfiguration()
			if err != nil {
				t.Fatal(err)
			}
			ctx = envcontext.WithEnvConfig(ctx, *envConfig)

			resetGlobalFlags()

			root := cli.NewCmdOdo(ctx, cli.OdoRecommendedName, cli.OdoRecommendedName, tt.clientset())

			var stdoutB, stderrB bytes.Buffer
			root.SetOut(&stdoutB)
			root.SetErr(&stderrB)

			root.SetArgs(tt.args)

			err = root.ExecuteContext(ctx)

			if (err != nil) != (tt.wantErr != "") {
				t.Fatalf("errWanted: %v\nGot: %v", tt.wantErr != "", err != nil)
			}

			if tt.wantErr != "" {
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("%q\nerror does not contain:\n%q", err.Error(), tt.wantErr)
				}
			}

			stdout := stdoutB.String()
			stderr := stderrB.String()

			if tt.wantStdout != "" {
				if !strings.Contains(stdout, tt.wantStdout) {
					t.Fatalf("%q\nstdout does not contain:\n%q", stdout, tt.wantStdout)
				}
			}

			if tt.wantStderr != "" {
				if !strings.Contains(stderr, tt.wantStderr) {
					t.Fatalf("%q\nstderr does not contain:\n%q", stderr, tt.wantStderr)
				}
			}
		})
	}
}

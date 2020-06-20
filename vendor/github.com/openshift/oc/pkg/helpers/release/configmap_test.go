package release

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"golang.org/x/crypto/openpgp"
)

type VerifierAccessor interface {
	Verifiers() map[string]openpgp.EntityList
}

func Test_loadReleaseVerifierFromConfigMap(t *testing.T) {
	redhatData, err := ioutil.ReadFile(filepath.Join("testdata", "keyrings", "redhat.txt"))
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name          string
		data          map[string]string
		want          bool
		wantErr       bool
		wantVerifiers int
	}{
		{
			name:    "requires data",
			data:    nil,
			wantErr: true,
		},
		{
			name: "requires stores",
			data: map[string]string{
				"verifier-public-key-redhat": string(redhatData),
			},
			wantErr: true,
		},
		{
			name: "requires verifiers",
			data: map[string]string{
				"store-local": "file://../testdata/signatures",
			},
			wantErr: true,
		},
		{
			name: "loads valid configuration",
			data: map[string]string{
				"verifier-public-key-redhat": string(redhatData),
				"store-local":                "file://../testdata/signatures",
			},
			want:          true,
			wantVerifiers: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewFromConfigMapData("from_test", tt.data, DefaultClient)
			if (err != nil) != tt.wantErr {
				t.Fatalf("loadReleaseVerifierFromPayload() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != nil != tt.want {
				t.Fatal(got)
			}
			if err != nil {
				return
			}
			if got == nil {
				return
			}
			if len(got.Verifiers()) != tt.wantVerifiers {
				t.Fatalf("unexpected release verifier: %#v", got)
			}
		})
	}
}

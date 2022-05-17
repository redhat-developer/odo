package vars

import (
	"reflect"
	"testing"

	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
)

func Test_parseKeyValueFile(t *testing.T) {
	type args struct {
		fileContent string
		lookupEnv   func(string) (string, bool)
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]string
		wantErr bool
	}{
		{
			name: "no error",
			args: args{
				lookupEnv: func(s string) (string, bool) {
					res, ok := map[string]string{
						"F": "a value for F from env",
					}[s]
					return res, ok
				},
				fileContent: `A=aze
# a comment

B=zerty
# a line beginning with spaces
  C=cvb
# a value beginning and ending with spaces
  D=  dfg  

# an empty value
E=

# a key with no value
F

# not defined in environment
G
`,
			},
			want: map[string]string{
				"A": "aze",
				"B": "zerty",
				"C": "cvb",
				"D": "  dfg  ",
				"E": "",
				"F": "a value for F from env",
			},
		},
		{
			name: "works with Windows EOL",
			args: args{
				lookupEnv: func(s string) (string, bool) {
					res, ok := map[string]string{
						"F": "a value for F from env",
					}[s]
					return res, ok
				},
				fileContent: "A=aze\r\nB=qsd",
			},
			want: map[string]string{
				"A": "aze",
				"B": "qsd",
			},
		},
		{
			name: "line without key",
			args: args{
				fileContent: `# a comment
=aze`,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filename := "vars.txt"
			fs := filesystem.NewFakeFs()
			fs.WriteFile(filename, []byte(tt.args.fileContent), 0444)

			got, err := parseKeyValueFile(fs, filename, tt.args.lookupEnv)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseKeyValueFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseKeyValueFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseKeyValueStrings(t *testing.T) {
	type args struct {
		strs      []string
		lookupEnv func(string) (string, bool)
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]string
		wantErr bool
	}{
		{
			name: "no error",
			args: args{
				strs: []string{
					"A=aze",
					"# a comment",
					"",
					"B=zerty",
					"# a line beginning with spaces",
					"  C=cvb",
					"# a value beginning and ending with spaces",
					"  D=  dfg  ",
					"# an empty value",
					"E=",
					"# a key with no value",
					"F",
					"# not defined in environment",
					"G",
				},
				lookupEnv: func(s string) (string, bool) {
					res, ok := map[string]string{
						"F": "a value for F from env",
					}[s]
					return res, ok
				},
			},
			want: map[string]string{
				"A": "aze",
				"B": "zerty",
				"C": "cvb",
				"D": "  dfg  ",
				"E": "",
				"F": "a value for F from env",
			},
		},
		{
			name: "string without key",
			args: args{
				strs: []string{
					"# a comment",
					"=aze",
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseKeyValueStrings(tt.args.strs, tt.args.lookupEnv)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseKeyValueStrings() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseKeyValueStrings() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetVariables(t *testing.T) {
	type args struct {
		fileContent string
		override    []string
		lookupEnv   func(string) (string, bool)
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]string
		wantErr bool
	}{
		{
			name: "overrides file",
			args: args{
				fileContent: `A=aze`,

				override: []string{
					"A=qsd",
				},

				lookupEnv: func(s string) (string, bool) {
					res, ok := map[string]string{
						"F": "a value for F from env",
					}[s]
					return res, ok
				},
			},
			want: map[string]string{
				"A": "qsd",
			},
		},
		{
			name: "no override between file and override",
			args: args{
				fileContent: `A=aze`,

				override: []string{
					"B=qsd",
				},

				lookupEnv: func(s string) (string, bool) {
					res, ok := map[string]string{
						"F": "a value for F from env",
					}[s]
					return res, ok
				},
			},
			want: map[string]string{
				"A": "aze",
				"B": "qsd",
			},
		},
		{
			name: "override file with env var",
			args: args{
				fileContent: `A=aze`,

				override: []string{
					"A",
				},

				lookupEnv: func(s string) (string, bool) {
					res, ok := map[string]string{
						"A": "a value for A from env",
					}[s]
					return res, ok
				},
			},
			want: map[string]string{
				"A": "a value for A from env",
			},
		},
		{
			name: "no override file with not defined env var",
			args: args{
				fileContent: `A=aze`,

				override: []string{
					"A",
				},

				lookupEnv: func(s string) (string, bool) {
					res, ok := map[string]string{}[s]
					return res, ok
				},
			},
			want: map[string]string{
				"A": "aze",
			},
		},
		{
			name: "override file with empty defined env var",
			args: args{
				fileContent: `A=aze`,

				override: []string{
					"A",
				},

				lookupEnv: func(s string) (string, bool) {
					res, ok := map[string]string{
						"A": "",
					}[s]
					return res, ok
				},
			},
			want: map[string]string{
				"A": "",
			},
		},
		{
			name: "error parsing file",
			args: args{
				fileContent: `=aze`,
			},
			wantErr: true,
		},
		{
			name: "error parsing override strings",
			args: args{
				fileContent: `A=aze`,
				override: []string{
					"=aze",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filename := "vars.txt"
			fs := filesystem.NewFakeFs()
			fs.WriteFile(filename, []byte(tt.args.fileContent), 0444)

			got, err := GetVariables(fs, filename, tt.args.override, tt.args.lookupEnv)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetVariables() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetVariables() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetVariablesEmptyFilename(t *testing.T) {
	type args struct {
		override  []string
		lookupEnv func(string) (string, bool)
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]string
		wantErr bool
	}{
		{
			name: "get override value",
			args: args{
				override: []string{
					"A=qsd",
				},

				lookupEnv: func(s string) (string, bool) {
					res, ok := map[string]string{
						"F": "a value for F from env",
					}[s]
					return res, ok
				},
			},
			want: map[string]string{
				"A": "qsd",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetVariables(filesystem.NewFakeFs(), "", tt.args.override, tt.args.lookupEnv)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetVariables() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetVariables() = %v, want %v", got, tt.want)
			}
		})
	}
}

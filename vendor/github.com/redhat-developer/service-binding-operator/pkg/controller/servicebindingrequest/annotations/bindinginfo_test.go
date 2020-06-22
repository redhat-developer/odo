package annotations

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func prefix(s string) string {
	return fmt.Sprintf("%s%s", ServiceBindingOperatorAnnotationPrefix, s)
}

// TestNewBindingInfo exercises annotation binding information parsing.
func TestNewBindingInfo(t *testing.T) {
	type args struct {
		name  string
		value string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    *BindingInfo
	}{
		{
			args: args{name: prefix("status.configMapRef-password"), value: "binding"},
			want: &BindingInfo{
				ResourceReferencePath: "status.configMapRef",
				Descriptor:            "binding:password",
				SourcePath:            "password",
				Value:                 "binding",
			},
			name:    "{fieldPath}-{path} annotation",
			wantErr: false,
		},
		{
			args: args{name: prefix("status.connectionString"), value: "binding"},
			want: &BindingInfo{
				Descriptor:            "binding:status.connectionString",
				ResourceReferencePath: "status.connectionString",
				SourcePath:            "status.connectionString",
				Value:                 "binding",
			},
			name:    "{path} annotation",
			wantErr: false,
		},
		{
			args: args{name: prefix(""), value: "binding"},
			want: &BindingInfo{
				Descriptor:            "binding:status.connectionString",
				ResourceReferencePath: "status.connectionString",
				SourcePath:            "status.connectionString",
				Value:                 "binding",
			},
			name:    "empty annotation name",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := NewBindingInfo(tt.args.name, tt.args.value)
			if err != nil && !tt.wantErr {
				t.Errorf("NewBindingInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else if err == nil {
				require.Equal(t, tt.want, b)
			}
		})
	}
}

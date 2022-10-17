package context

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/pflag"
)

func TestGetContextProperties(t *testing.T) {
	ckey, value := "preferenceKey", "consenttelemetry"
	ctx := NewContext(context.Background())
	setContextProperty(ctx, ckey, value)

	got := GetContextProperties(ctx)
	want := map[string]interface{}{ckey: value}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("want: %q got: %q", want, got)
	}
}

func TestSetComponentType(t *testing.T) {
	want := "java"
	for _, value := range []string{"java", "java:8", "myproject/java:8"} {
		ctx := NewContext(context.Background())
		SetComponentType(ctx, value)

		if got, contains := GetContextProperties(ctx)[ComponentType]; !contains || got != want {
			t.Errorf("component type was not set. Got: %q, Want: %q", got, want)
		}
	}
}

// TODO(feloy) test with fake kclient implementation
//func TestSetClusterType(t *testing.T) {
//	tests := []struct {
//		want   string
//		groups []string
//	}{
//		{
//			want:   "openshift3",
//			groups: []string{"project.openshift.io/v1"},
//		},
//		{
//			want:   "openshift4",
//			groups: []string{"project.openshift.io/v1", "operators.coreos.com/v1alpha1"},
//		},
//		{
//			want: "kubernetes",
//		},
//		{
//			want: NOTFOUND,
//		},
//	}
//
//	for _, tt := range tests {
//		var fakeClient *kclient.Client
//		if tt.want != NOTFOUND {
//			fakeClient, _ = kclient.FakeNew()
//		}
//		if tt.groups != nil {
//			setupCluster(fakeClient, tt.groups)
//		}
//
//		ctx := NewContext(context.Background())
//		SetClusterType(ctx, fakeClient)
//
//		got := GetContextProperties(ctx)[ClusterType]
//		if got != tt.want {
//			t.Errorf("got: %q, want: %q", got, tt.want)
//		}
//	}
//}

func TestGetTelemetryStatus(t *testing.T) {
	want := true
	ctx := NewContext(context.Background())
	setContextProperty(ctx, TelemetryStatus, want)
	got := GetTelemetryStatus(ctx)
	if got != want {
		t.Errorf("got: %v, want: %v", got, want)
	}
}

func TestSetTelemetryStatus(t *testing.T) {
	want := false
	ctx := NewContext(context.Background())
	SetTelemetryStatus(ctx, want)
	got := GetContextProperties(ctx)[TelemetryStatus]
	if got != want {
		t.Errorf("got: %v, want: %v", got, want)
	}
}

//var apiResourceList = map[string]*metav1.APIResourceList{
//	"operators.coreos.com/v1alpha1": {
//		GroupVersion: "operators.coreos.com/v1alpha1",
//		APIResources: []metav1.APIResource{{
//			Name:         "clusterserviceversions",
//			SingularName: "clusterserviceversion",
//			Namespaced:   false,
//			Kind:         "ClusterServiceVersion",
//			ShortNames:   []string{"csv", "csvs"},
//		}},
//	},
//	"project.openshift.io/v1": {
//		GroupVersion: "project.openshift.io/v1",
//		APIResources: []metav1.APIResource{{
//			Name:         "projects",
//			SingularName: "project",
//			Namespaced:   false,
//			Kind:         "Project",
//			ShortNames:   []string{"proj"},
//		}},
//	},
//}

// setupCluster adds resource groups to the client
//func setupCluster(fakeClient kclient.ClientInterface, groupVersion []string) {
//	fd := odoFake.NewFakeDiscovery()
//	for _, group := range groupVersion {
//		fd.AddResourceList(group, apiResourceList[group])
//	}
//	fakeClient.SetDiscoveryInterface(fd)
//}

func TestSetFlags(t *testing.T) {
	type args struct {
		ctx   context.Context
		flags *pflag.FlagSet
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "no flags",
			args: args{
				ctx:   NewContext(context.Background()),
				flags: &pflag.FlagSet{},
			},
			want: "",
		},
		{
			name: "one changed flag",
			args: args{
				ctx: NewContext(context.Background()),
				flags: func() *pflag.FlagSet {
					f := &pflag.FlagSet{}
					f.String("flag1", "", "")
					//nolint:errcheck
					f.Set("flag1", "value1")
					return f
				}(),
			},
			want: "flag1",
		},
		{
			name: "one changed flag, one unchanged flag",
			args: args{
				ctx: NewContext(context.Background()),
				flags: func() *pflag.FlagSet {
					f := &pflag.FlagSet{}
					f.String("flag1", "", "")
					f.String("flag2", "", "")
					//nolint:errcheck
					f.Set("flag2", "value1")
					return f
				}(),
			},
			want: "flag2",
		},
		{
			name: "two changed flags",
			args: args{
				ctx: NewContext(context.Background()),
				flags: func() *pflag.FlagSet {
					f := &pflag.FlagSet{}
					f.String("flag1", "", "")
					f.String("flag2", "", "")
					//nolint:errcheck
					f.Set("flag1", "value1")
					//nolint:errcheck
					f.Set("flag2", "value1")
					return f
				}(),
			},
			want: "flag1 flag2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetFlags(tt.args.ctx, tt.args.flags)
			got := GetContextProperties(tt.args.ctx)[Flags]
			if got != tt.want {
				t.Errorf("SetFlags() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetCaller(t *testing.T) {
	type testScope struct {
		name       string
		callerType string
		wantErr    bool
		want       interface{}
	}

	tests := []testScope{
		{
			name:       "empty caller",
			callerType: "",
			want:       "",
		},
		{
			name:       "unknown caller",
			callerType: "an-unknown-caller",
			wantErr:    true,
			want:       "an-unknown-caller",
		},
		{
			name:       "case-insensitive caller",
			callerType: strings.ToUpper(IntelliJ),
			want:       IntelliJ,
		},
		{
			name:       "trimming space from caller",
			callerType: fmt.Sprintf("   %s\t", VSCode),
			want:       VSCode,
		},
	}
	for _, c := range []string{VSCode, IntelliJ, JBoss} {
		tests = append(tests, testScope{
			name:       fmt.Sprintf("valid caller: %s", c),
			callerType: c,
			want:       c,
		})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewContext(context.Background())

			err := SetCaller(ctx, tt.callerType)

			if !tt.wantErr == (err != nil) {
				t.Errorf("unexpected error %v, wantErr %v", err, tt.wantErr)
			}

			got := GetContextProperties(ctx)[Caller]
			if got != tt.want {
				t.Errorf("SetCaller() = %v, want %v", got, tt.want)
			}
		})
	}
}

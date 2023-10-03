package auth

import (
	"reflect"
	"testing"
)

func Test_filteredInformation(t *testing.T) {
	type args struct {
		s []byte
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "login with no projects",
			args: args{
				s: []byte(`Logged into "https://api.crc.testing:6443" as "developer" using existing credentials.

			You don't have any projects. You can try to create a new project, by running
			
				oc new-project <projectname>`),
			},
			want: []byte(`Logged into "https://api.crc.testing:6443" as "developer" using existing credentials.

			You don't have any projects. You can try to create a new project, by running
			
				odo create project <projectname>`),
		},
		{
			name: "login with only 1 project",
			args: args{
				s: []byte(`Logged into "https://api.crc.testing:6443" as "developer" using existing credentials.

			You have one project on this server: "test1"
			
			Using project "test1".`),
			},
			want: []byte(`Logged into "https://api.crc.testing:6443" as "developer" using existing credentials.

			You have one project on this server: "test1"
			
			Using project "test1".`),
		},
		{
			name: "login with more than one project",
			args: args{
				s: []byte(`Logged into "https://api.crc.testing:6443" as "developer" using existing credentials.

			You have access to the following projects and can switch between them with 'oc project <projectname>':
			
				test1
				test2
			  * test3
			
			Using project "test3".`),
			},
			want: []byte(`Logged into "https://api.crc.testing:6443" as "developer" using existing credentials.

			You have access to the following projects and can switch between them with 'odo set project <projectname>':
			
				test1
				test2
			  * test3
			
			Using project "test3".`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := filteredInformation(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("filteredInformation() = %v\nwant = %v", string(got), string(tt.want))
			}
		})
	}
}

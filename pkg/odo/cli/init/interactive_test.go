package init

import (
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/redhat-developer/odo/pkg/catalog"
)

func TestInteractiveBuilder_ParamsBuild(t *testing.T) {
	type fields struct {
		buildAsker         func(ctrl *gomock.Controller) asker
		buildCatalogClient func(ctrl *gomock.Controller) catalog.Client
	}
	tests := []struct {
		name    string
		fields  fields
		want    initParams
		wantErr bool
	}{
		{
			fields: fields{
				buildAsker: func(ctrl *gomock.Controller) asker {
					client := NewMockasker(ctrl)
					client.EXPECT().askLanguage(gomock.Any()).Return("java", nil)
					client.EXPECT().askType(gomock.Any()).Return(catalog.DevfileComponentType{
						Name: "a-devfile-name",
						Registry: catalog.Registry{
							Name: "MyRegistry1",
						},
					}, nil)
					client.EXPECT().askStarterProject(gomock.Any()).Return("starter1", nil)
					client.EXPECT().askName(gomock.Any()).Return("a-name", nil)
					return client
				},
				buildCatalogClient: func(ctrl *gomock.Controller) catalog.Client {
					client := catalog.NewMockClient(ctrl)
					client.EXPECT().ListDevfileComponents(gomock.Any())
					client.EXPECT().GetStarterProjectsNames(gomock.Any()).Return([]string{"starter1", "starter2"}, nil)
					return client
				},
			},
			want: initParams{
				name:            "a-name",
				devfile:         "a-devfile-name",
				devfileRegistry: "MyRegistry1",
				starter:         "starter1",
			},
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			o := &InteractiveBuilder{
				asker:         tt.fields.buildAsker(ctrl),
				catalogClient: tt.fields.buildCatalogClient(ctrl),
			}
			got, err := o.ParamsBuild()
			if (err != nil) != tt.wantErr {
				t.Errorf("InteractiveBuilder.ParamsBuild() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InteractiveBuilder.ParamsBuild() = %v, want %v", got, tt.want)
			}
		})
	}
}

package params

import (
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/redhat-developer/odo/pkg/catalog"
	"github.com/redhat-developer/odo/pkg/odo/cli/init/asker"
)

func TestInteractiveBuilder_ParamsBuild(t *testing.T) {
	type fields struct {
		buildAsker         func(ctrl *gomock.Controller) asker.Asker
		buildCatalogClient func(ctrl *gomock.Controller) catalog.Client
	}
	tests := []struct {
		name    string
		fields  fields
		want    InitParams
		wantErr bool
	}{
		{
			name: "direct selection",
			fields: fields{
				buildAsker: func(ctrl *gomock.Controller) asker.Asker {
					client := asker.NewMockAsker(ctrl)
					client.EXPECT().AskLanguage(gomock.Any()).Return("java", nil)
					client.EXPECT().AskType(gomock.Any()).Return(false, catalog.DevfileComponentType{
						Name: "a-devfile-name",
						Registry: catalog.Registry{
							Name: "MyRegistry1",
						},
					}, nil)
					client.EXPECT().AskStarterProject(gomock.Any()).Return(false, "starter1", nil)
					client.EXPECT().AskName(gomock.Any()).Return("a-name", nil)
					return client
				},
				buildCatalogClient: func(ctrl *gomock.Controller) catalog.Client {
					client := catalog.NewMockClient(ctrl)
					client.EXPECT().ListDevfileComponents(gomock.Any())
					client.EXPECT().GetStarterProjectsNames(gomock.Any()).Return([]string{"starter1", "starter2"}, nil)
					return client
				},
			},
			want: InitParams{
				Name:            "a-name",
				Devfile:         "a-devfile-name",
				DevfileRegistry: "MyRegistry1",
				Starter:         "starter1",
			},
		},
		{
			name: "selection with back",
			fields: fields{
				buildAsker: func(ctrl *gomock.Controller) asker.Asker {
					client := asker.NewMockAsker(ctrl)
					client.EXPECT().AskLanguage(gomock.Any()).Return("java", nil)
					client.EXPECT().AskType(gomock.Any()).Return(true, catalog.DevfileComponentType{}, nil)
					client.EXPECT().AskLanguage(gomock.Any()).Return("go", nil)
					client.EXPECT().AskType(gomock.Any()).Return(false, catalog.DevfileComponentType{
						Name: "a-devfile-name",
						Registry: catalog.Registry{
							Name: "MyRegistry1",
						},
					}, nil)
					client.EXPECT().AskStarterProject(gomock.Any()).Return(true, "", nil)
					client.EXPECT().AskType(gomock.Any()).Return(false, catalog.DevfileComponentType{
						Name: "another-devfile-name",
						Registry: catalog.Registry{
							Name: "MyRegistry2",
						},
					}, nil)
					client.EXPECT().AskStarterProject(gomock.Any()).Return(false, "starter1", nil)
					client.EXPECT().AskName(gomock.Any()).Return("a-name", nil)
					return client
				},
				buildCatalogClient: func(ctrl *gomock.Controller) catalog.Client {
					client := catalog.NewMockClient(ctrl)
					client.EXPECT().ListDevfileComponents(gomock.Any())
					client.EXPECT().GetStarterProjectsNames(gomock.Any()).Return([]string{"starter1", "starter2"}, nil).AnyTimes()
					return client
				},
			},
			want: InitParams{
				Name:            "a-name",
				Devfile:         "another-devfile-name",
				DevfileRegistry: "MyRegistry2",
				Starter:         "starter1",
			},
		},
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

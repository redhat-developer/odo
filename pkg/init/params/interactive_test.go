package params

import (
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/redhat-developer/odo/pkg/catalog"
	"github.com/redhat-developer/odo/pkg/init/asker"
)

func TestInteractiveBuilder_ParamsBuild(t *testing.T) {
	type fields struct {
		buildAsker         func(ctrl *gomock.Controller) asker.Asker
		buildCatalogClient func(ctrl *gomock.Controller) catalog.Client
	}
	tests := []struct {
		name    string
		fields  fields
		want    *DevfileLocation
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
					return client
				},
				buildCatalogClient: func(ctrl *gomock.Controller) catalog.Client {
					client := catalog.NewMockClient(ctrl)
					client.EXPECT().ListDevfileComponents(gomock.Any())
					return client
				},
			},
			want: &DevfileLocation{
				Devfile:         "a-devfile-name",
				DevfileRegistry: "MyRegistry1",
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
					return client
				},
				buildCatalogClient: func(ctrl *gomock.Controller) catalog.Client {
					client := catalog.NewMockClient(ctrl)
					client.EXPECT().ListDevfileComponents(gomock.Any())
					return client
				},
			},
			want: &DevfileLocation{
				Devfile:         "a-devfile-name",
				DevfileRegistry: "MyRegistry1",
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

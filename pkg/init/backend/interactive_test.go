package backend

import (
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/redhat-developer/odo/pkg/catalog"
	"github.com/redhat-developer/odo/pkg/init/asker"
)

func TestDetectFramework(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name             string
		args             args
		wantedLanguage   string
		wantedTools      []string
		wantedFrameworks []string
		wantErr          bool
	}{
		{
			name: "Case 1 - Detect Node.JS example",
			args: args{
				path: GetTestProjectPath("nodejs"),
			},
			wantedLanguage:   "javascript",
			wantedTools:      []string{"nodejs"},
			wantedFrameworks: []string{},
			wantErr:          false,
		},
		{
			name: "Case 2 - Detect java openjdk example",
			args: args{
				path: GetTestProjectPath("openjdk"),
			},
			wantedLanguage:   "java",
			wantedTools:      []string{"maven"},
			wantedFrameworks: []string{},
			wantErr:          false,
		},
		{
			name: "Case 3 - Detect python example",
			args: args{
				path: GetTestProjectPath("python"),
			},
			wantedLanguage:   "python",
			wantedTools:      []string{},
			wantedFrameworks: []string{},
			wantErr:          false,
		},
		{
			name: "Case 3 - Detect java wildfly example",
			args: args{
				path: GetTestProjectPath("wildfly"),
			},
			wantedLanguage:   "java",
			wantedTools:      []string{},
			wantedFrameworks: []string{},
			wantErr:          false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Run function DetectFramework
			languages, err := DetectFramework(tt.args.path)

			if !tt.wantErr == (err != nil) {
				t.Errorf("unexpected error %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !hasWantedLanguage(languages, tt.wantedLanguage, tt.wantedTools, tt.wantedFrameworks) {
				t.Errorf("Expected: Language: %s, Tools: %s, Framework: %s\nGot: %+v", tt.wantedLanguage, tt.wantedTools, tt.wantedFrameworks, languages)
			}

		})
	}
}

func TestInteractiveBackend_SelectDevfile(t *testing.T) {
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
			o := &InteractiveBackend{
				asker:         tt.fields.buildAsker(ctrl),
				catalogClient: tt.fields.buildCatalogClient(ctrl),
			}
			_, got, err := o.SelectDevfile(map[string]string{})
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

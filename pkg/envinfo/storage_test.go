package envinfo

import (
	"reflect"
	"testing"

	"github.com/devfile/library/pkg/devfile/parser/data"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/devfile/library/pkg/testingutil"
	"github.com/kylelemons/godebug/pretty"
	"github.com/openshift/odo/pkg/localConfigProvider"
)

func TestGetVolumeMountPath(t *testing.T) {

	tests := []struct {
		name        string
		volumeMount devfilev1.VolumeMount
		wantPath    string
	}{
		{
			name: "Case 1: Mount Path is present",
			volumeMount: devfilev1.VolumeMount{
				Name: "name1",
				Path: "/path1",
			},
			wantPath: "/path1",
		},
		{
			name: "Case 2: Mount Path is absent",
			volumeMount: devfilev1.VolumeMount{
				Name: "name1",
			},
			wantPath: "/name1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := GetVolumeMountPath(tt.volumeMount)

			if path != tt.wantPath {
				t.Errorf("TestGetVolumeMountPath error: mount path mismatch, expected: %v got: %v", tt.wantPath, path)
			}
		})
	}

}

func TestEnvInfo_CompleteStorage(t *testing.T) {
	type args struct {
		storage *localConfigProvider.LocalStorage
	}
	tests := []struct {
		name string
		args args
		want *localConfigProvider.LocalStorage
	}{
		{
			name: "case 1: size is empty",
			args: args{
				storage: &localConfigProvider.LocalStorage{
					Name: "storage-0",
					Path: "/data",
				},
			},
			want: &localConfigProvider.LocalStorage{
				Name: "storage-0",
				Path: "/data",
				Size: "1Gi",
			},
		},
		{
			name: "case 2: path is empty",
			args: args{
				storage: &localConfigProvider.LocalStorage{
					Name: "storage-0",
				},
			},
			want: &localConfigProvider.LocalStorage{
				Name: "storage-0",
				Path: "/storage-0",
				Size: "1Gi",
			},
		},
		{
			name: "case 3: all the writable fields are filled",
			args: args{
				storage: &localConfigProvider.LocalStorage{
					Name: "storage-0",
					Path: "/data",
					Size: "1Gi",
				},
			},
			want: &localConfigProvider.LocalStorage{
				Name: "storage-0",
				Path: "/data",
				Size: "1Gi",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ei := &EnvInfo{}
			ei.CompleteStorage(tt.args.storage)

			if !reflect.DeepEqual(tt.want, tt.args.storage) {
				t.Errorf("got storage is different: %v", pretty.Compare(tt.want, tt.args.storage))
			}
		})
	}
}

func TestEnvInfo_ListStorage(t *testing.T) {
	type fields struct {
		devfileObj parser.DevfileObj
	}
	tests := []struct {
		name    string
		fields  fields
		want    []localConfigProvider.LocalStorage
		wantErr bool
	}{
		{
			name: "case 1: list all the volumes in the devfile along with their respective size and containers",
			fields: fields{
				devfileObj: parser.DevfileObj{
					Data: func() data.DevfileData {
						devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
						if err != nil {
							t.Error(err)
						}
						err = devfileData.AddComponents([]devfilev1.Component{
							{
								Name: "container-0",
								ComponentUnion: devfilev1.ComponentUnion{
									Container: &devfilev1.ContainerComponent{
										Container: devfilev1.Container{
											VolumeMounts: []devfilev1.VolumeMount{
												{
													Name: "volume-0",
													Path: "/path",
												},
												{
													Name: "volume-1",
													Path: "/data",
												},
											},
										},
									},
								},
							},
							{
								Name: "container-1",
								ComponentUnion: devfilev1.ComponentUnion{
									Container: &devfilev1.ContainerComponent{
										Container: devfilev1.Container{
											VolumeMounts: []devfilev1.VolumeMount{
												{
													Name: "volume-1",
													Path: "/data",
												},
											},
										},
									},
								},
							},
							testingutil.GetFakeVolumeComponent("volume-0", "5Gi"),
							testingutil.GetFakeVolumeComponent("volume-1", "10Gi"),
						})
						if err != nil {
							t.Error(err)
						}
						return devfileData
					}(),
				},
			},
			want: []localConfigProvider.LocalStorage{
				{
					Name:      "volume-0",
					Size:      "5Gi",
					Path:      "/path",
					Container: "container-0",
				},
				{
					Name:      "volume-1",
					Size:      "10Gi",
					Path:      "/data",
					Container: "container-0",
				},
				{
					Name:      "volume-1",
					Size:      "10Gi",
					Path:      "/data",
					Container: "container-1",
				},
			},
		},
		{
			name: "case 2: list all the volumes in the devfile with the default size when no size is mentioned",
			fields: fields{
				devfileObj: parser.DevfileObj{
					Data: func() data.DevfileData {
						devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
						if err != nil {
							t.Error(err)
						}
						err = devfileData.AddComponents([]devfilev1.Component{
							{
								Name: "container-0",
								ComponentUnion: devfilev1.ComponentUnion{
									Container: &devfilev1.ContainerComponent{
										Container: devfilev1.Container{
											VolumeMounts: []devfilev1.VolumeMount{
												{
													Name: "volume-0",
													Path: "/path",
												},
												{
													Name: "volume-1",
													Path: "/data",
												},
											},
										},
									},
								},
							},
							testingutil.GetFakeVolumeComponent("volume-0", ""),
							testingutil.GetFakeVolumeComponent("volume-1", "10Gi"),
						})
						if err != nil {
							t.Error(err)
						}
						return devfileData
					}(),
				},
			},
			want: []localConfigProvider.LocalStorage{
				{
					Name:      "volume-0",
					Size:      "1Gi",
					Path:      "/path",
					Container: "container-0",
				},
				{
					Name:      "volume-1",
					Size:      "10Gi",
					Path:      "/data",
					Container: "container-0",
				},
			},
		},
		{
			name: "case 3: list all the volumes in the devfile with the default mount path when no path is mentioned",
			fields: fields{
				devfileObj: parser.DevfileObj{
					Data: func() data.DevfileData {
						devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
						if err != nil {
							t.Error(err)
						}
						err = devfileData.AddComponents([]devfilev1.Component{
							{
								Name: "container-0",
								ComponentUnion: devfilev1.ComponentUnion{
									Container: &devfilev1.ContainerComponent{
										Container: devfilev1.Container{
											VolumeMounts: []devfilev1.VolumeMount{
												{
													Name: "volume-0",
												},
											},
										},
									},
								},
							},
							testingutil.GetFakeVolumeComponent("volume-0", ""),
						})
						if err != nil {
							t.Error(err)
						}
						return devfileData
					}(),
				},
			},
			want: []localConfigProvider.LocalStorage{
				{
					Name:      "volume-0",
					Size:      "1Gi",
					Path:      "/volume-0",
					Container: "container-0",
				},
			},
		},
		{
			name: "case 4: return empty when no volumes is mounted",
			fields: fields{
				devfileObj: parser.DevfileObj{
					Data: func() data.DevfileData {
						devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
						if err != nil {
							t.Error(err)
						}
						err = devfileData.AddComponents([]devfilev1.Component{
							{
								Name: "container-0",
								ComponentUnion: devfilev1.ComponentUnion{
									Container: &devfilev1.ContainerComponent{
										Container: devfilev1.Container{},
									},
								},
							},
							testingutil.GetFakeVolumeComponent("volume-0", ""),
							testingutil.GetFakeVolumeComponent("volume-1", "10Gi"),
						})
						if err != nil {
							t.Error(err)
						}
						return devfileData
					}(),
				},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ei := &EnvInfo{
				devfileObj: tt.fields.devfileObj,
			}
			got, err := ei.ListStorage()
			if (err != nil) != tt.wantErr {
				t.Errorf("ListStorage() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ListStorage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnvInfo_ValidateStorage(t *testing.T) {
	type fields struct {
		devfileObj parser.DevfileObj
	}
	type args struct {
		storage localConfigProvider.LocalStorage
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "case 1: storage with the same name doesn't exist",
			fields: fields{
				devfileObj: parser.DevfileObj{
					Data: func() data.DevfileData {
						devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
						if err != nil {
							t.Error(err)
						}
						err = devfileData.AddComponents([]devfilev1.Component{
							{
								Name: "container-0",
								ComponentUnion: devfilev1.ComponentUnion{
									Container: &devfilev1.ContainerComponent{
										Container: devfilev1.Container{},
									},
								},
							},
						})
						if err != nil {
							t.Error(err)
						}
						return devfileData
					}(),
				},
			},
			args: args{
				storage: localConfigProvider.LocalStorage{
					Name: "volume-0",
					Size: "10Gi",
					Path: "/data",
				},
			},
		},
		{
			name: "case 2: storage with same name exists",
			fields: fields{
				devfileObj: parser.DevfileObj{
					Data: func() data.DevfileData {
						devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
						if err != nil {
							t.Error(err)
						}
						err = devfileData.AddComponents([]devfilev1.Component{
							{
								Name: "container-0",
								ComponentUnion: devfilev1.ComponentUnion{
									Container: &devfilev1.ContainerComponent{
										Container: devfilev1.Container{
											VolumeMounts: []devfilev1.VolumeMount{
												{
													Name: "volume-0",
													Path: "/path",
												},
												{
													Name: "volume-1",
													Path: "/data",
												},
											},
										},
									},
								},
							},
							{
								Name: "container-1",
								ComponentUnion: devfilev1.ComponentUnion{
									Container: &devfilev1.ContainerComponent{
										Container: devfilev1.Container{
											VolumeMounts: []devfilev1.VolumeMount{
												{
													Name: "volume-1",
													Path: "/data",
												},
											},
										},
									},
								},
							},
							testingutil.GetFakeVolumeComponent("volume-0", "5Gi"),
						})
						if err != nil {
							t.Error(err)
						}
						return devfileData
					}(),
				},
			},
			args: args{
				storage: localConfigProvider.LocalStorage{
					Name: "volume-0",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ei := &EnvInfo{
				devfileObj: tt.fields.devfileObj,
			}
			if err := ei.ValidateStorage(tt.args.storage); (err != nil) != tt.wantErr {
				t.Errorf("ValidateStorage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnvInfo_GetStorage(t *testing.T) {
	type fields struct {
		devfileObj parser.DevfileObj
	}
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *localConfigProvider.LocalStorage
		wantErr bool
	}{
		{
			name: "case 1: storage with the given name doesn't exist",
			fields: fields{
				devfileObj: parser.DevfileObj{
					Data: func() data.DevfileData {
						devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
						if err != nil {
							t.Error(err)
						}
						err = devfileData.AddComponents([]devfilev1.Component{
							{
								Name: "container-0",
								ComponentUnion: devfilev1.ComponentUnion{
									Container: &devfilev1.ContainerComponent{
										Container: devfilev1.Container{},
									},
								},
							},
						})
						if err != nil {
							t.Error(err)
						}
						return devfileData
					}(),
				},
			},
			args: args{
				name: "volume-0",
			},
			want: nil,
		},
		{
			name: "case 2: storage with the given name exists",
			fields: fields{
				devfileObj: parser.DevfileObj{
					Data: func() data.DevfileData {
						devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
						if err != nil {
							t.Error(err)
						}
						err = devfileData.AddComponents([]devfilev1.Component{
							{
								Name: "container-0",
								ComponentUnion: devfilev1.ComponentUnion{
									Container: &devfilev1.ContainerComponent{
										Container: devfilev1.Container{
											VolumeMounts: []devfilev1.VolumeMount{
												{
													Name: "volume-0",
													Path: "/path",
												},
												{
													Name: "volume-1",
													Path: "/data",
												},
											},
										},
									},
								},
							},
							{
								Name: "container-1",
								ComponentUnion: devfilev1.ComponentUnion{
									Container: &devfilev1.ContainerComponent{
										Container: devfilev1.Container{
											VolumeMounts: []devfilev1.VolumeMount{
												{
													Name: "volume-1",
													Path: "/data",
												},
											},
										},
									},
								},
							},
							testingutil.GetFakeVolumeComponent("volume-0", "5Gi"),
							testingutil.GetFakeVolumeComponent("volume-1", "10Gi"),
						})
						if err != nil {
							t.Error(err)
						}
						return devfileData
					}(),
				},
			},
			args: args{
				name: "volume-0",
			},
			want: &localConfigProvider.LocalStorage{
				Name:      "volume-0",
				Size:      "5Gi",
				Path:      "/path",
				Container: "container-0",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ei := &EnvInfo{
				devfileObj: tt.fields.devfileObj,
			}

			got, err := ei.GetStorage(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetStorage() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetStorage() = %v, want %v", got, tt.want)
			}
		})
	}
}

package storage

import (
	"testing"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/v2/pkg/devfile/parser"
	"github.com/devfile/library/v2/pkg/devfile/parser/data"
	"github.com/google/go-cmp/cmp"

	"github.com/redhat-developer/odo/pkg/testingutil"
)

func TestEnvInfo_ListStorage(t *testing.T) {
	type fields struct {
		devfileObj parser.DevfileObj
	}
	tests := []struct {
		name    string
		fields  fields
		want    []LocalStorage
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
			want: []LocalStorage{
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
			want: []LocalStorage{
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
			want: []LocalStorage{
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
			got, err := ListStorage(tt.fields.devfileObj)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListStorage() error = %v, wantErr %v", err, tt.wantErr)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ListStorage() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

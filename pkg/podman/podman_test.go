package podman

import (
	"os"
	"reflect"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
)

func TestPodmanCli_PodLs(t *testing.T) {
	type fields struct {
		podmanCmd                   string
		podmanCmdInitTimeout        time.Duration
		containerRunGlobalExtraArgs []string
		containerRunExtraArgs       []string
	}
	tests := []struct {
		name       string
		fields     fields
		populateFS func()
		want       map[string]bool
		wantErr    bool
	}{
		{
			name: "command fails",
			fields: fields{
				podmanCmd: "false",
			},
			wantErr: true,
		},
		{
			name: "command works, returns nothing",
			fields: fields{
				podmanCmd: "true",
			},
			wantErr: false,
			want:    map[string]bool{},
		},
		{
			name: "command works, returns pods",
			fields: fields{
				podmanCmd: "./podman.fake.sh",
			},
			populateFS: func() {
				script := []byte(`#!/bin/sh
case "$*" in
	"pod list --format {{.Name}} --noheading")
		echo name1
		echo name2
		echo name3
		;;
esac`)
				err := os.WriteFile("podman.fake.sh", script, 0755)
				if err != nil {
					t.Fatal(err)
				}
			},
			wantErr: false,
			want: map[string]bool{
				"name1": true,
				"name2": true,
				"name3": true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.populateFS != nil {
				originWd, err := os.Getwd()
				if err != nil {
					t.Fatal(err)
				}
				defer func() {
					_ = os.Chdir(originWd)
				}()
				cwd := t.TempDir()
				err = os.Chdir(cwd)
				if err != nil {
					t.Fatal(err)
				}
				tt.populateFS()
			}

			o := &PodmanCli{
				podmanCmd:                   tt.fields.podmanCmd,
				podmanCmdInitTimeout:        tt.fields.podmanCmdInitTimeout,
				containerRunGlobalExtraArgs: tt.fields.containerRunGlobalExtraArgs,
				containerRunExtraArgs:       tt.fields.containerRunExtraArgs,
			}
			got, err := o.PodLs()
			if (err != nil) != tt.wantErr {
				t.Errorf("PodmanCli.PodLs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PodmanCli.PodLs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPodmanCli_KubeGenerate(t *testing.T) {
	type fields struct {
		podmanCmd                   string
		podmanCmdInitTimeout        time.Duration
		containerRunGlobalExtraArgs []string
		containerRunExtraArgs       []string
	}
	type args struct {
		name string
	}
	tests := []struct {
		name        string
		fields      fields
		populateFS  func()
		args        args
		checkResult func(*corev1.Pod)
		wantErr     bool
	}{
		{
			name: "pod not found",
			args: args{
				name: "pod-not-found",
			},
			fields: fields{
				podmanCmd: "./podman.fake.sh",
			},
			populateFS: func() {
				script := []byte(`#!/bin/sh
case "$*" in
	"generate kube pod-not-found")
		exit 125
		;;
esac`)
				err := os.WriteFile("podman.fake.sh", script, 0755)
				if err != nil {
					t.Fatal(err)
				}
			},
			wantErr: true,
		},
		{
			name: "command works, returns pod",
			args: args{
				name: "my-pod",
			},
			fields: fields{
				podmanCmd: "./podman.fake.sh",
			},
			populateFS: func() {
				script := []byte(`#!/bin/sh
case "$*" in
	"generate kube my-pod")
		cat <<EOF
apiVersion: v1
kind: Pod
metadata:
  name: my-pod		
EOF
		;;
esac`)
				err := os.WriteFile("podman.fake.sh", script, 0755)
				if err != nil {
					t.Fatal(err)
				}
			},
			wantErr: false,
			checkResult: func(pod *corev1.Pod) {
				podName := pod.GetName()
				if podName != "my-pod" {
					t.Errorf("pod name should be %q but is %q", "my-pod", podName)
				}
			},
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.populateFS != nil {
				originWd, err := os.Getwd()
				if err != nil {
					t.Fatal(err)
				}
				defer func() {
					_ = os.Chdir(originWd)
				}()
				cwd := t.TempDir()
				err = os.Chdir(cwd)
				if err != nil {
					t.Fatal(err)
				}
				tt.populateFS()
			}

			o := &PodmanCli{
				podmanCmd:                   tt.fields.podmanCmd,
				podmanCmdInitTimeout:        tt.fields.podmanCmdInitTimeout,
				containerRunGlobalExtraArgs: tt.fields.containerRunGlobalExtraArgs,
				containerRunExtraArgs:       tt.fields.containerRunExtraArgs,
			}
			got, err := o.KubeGenerate(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("PodmanCli.KubeGenerate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.checkResult != nil {
				tt.checkResult(got)
			}
		})
	}
}

func TestPodmanCli_CleanupPodResources(t *testing.T) {
	type fields struct {
		podmanCmd                   string
		podmanCmdInitTimeout        time.Duration
		containerRunGlobalExtraArgs []string
		containerRunExtraArgs       []string
	}
	type args struct {
		pod            func() *corev1.Pod
		cleanupVolumes bool
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		populateFS  func()
		wantErr     bool
		checkResult func()
	}{
		{
			name: "cleanup pod, not volumes",
			fields: fields{
				podmanCmd: "./podman.fake.sh",
			},
			args: args{
				pod: func() *corev1.Pod {
					pod := corev1.Pod{}
					pod.SetName("my-pod")
					pod.Spec.Volumes = []corev1.Volume{
						{
							Name: "vol1",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "volume1",
								},
							},
						},
						{
							Name: "vol2",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "volume2",
								},
							},
						},
					}
					return &pod
				},
				cleanupVolumes: false,
			},
			populateFS: func() {
				script := []byte(`#!/bin/sh
case "$*" in
	"pod stop my-pod")
		touch stop
		echo my-pod
		;;
	"pod rm my-pod")
		touch rm
		echo my-pod	
		;;
	"volume rm volume1")
		touch volume1
		;;
	"volume rm volume2")
		touch volume2
		;;
esac`)
				err := os.WriteFile("podman.fake.sh", script, 0755)
				if err != nil {
					t.Fatal(err)
				}
			},
			checkResult: func() {
				_, err := os.Stat("stop")
				if err != nil {
					t.Errorf("podman stop has not been called")
				}
				_, err = os.Stat("rm")
				if err != nil {
					t.Errorf("podman rm has not been called")
				}
				_, err = os.Stat("volume1")
				if err == nil {
					t.Errorf("podman rm volume volume1 has been called, it should not")
				}
				_, err = os.Stat("volume2")
				if err == nil {
					t.Errorf("podman rm volume volume2 has been called, it should not")
				}
			},
		},
		{
			name: "cleanup pod and volumes",
			fields: fields{
				podmanCmd: "./podman.fake.sh",
			},
			args: args{
				pod: func() *corev1.Pod {
					pod := corev1.Pod{}
					pod.SetName("my-pod")
					pod.Spec.Volumes = []corev1.Volume{
						{
							Name: "vol1",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "volume1",
								},
							},
						},
						{
							Name: "vol2",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "volume2",
								},
							},
						},
					}
					return &pod
				},
				cleanupVolumes: true,
			},
			populateFS: func() {
				script := []byte(`#!/bin/sh
case "$*" in
	"pod stop my-pod")
		touch stop
		echo my-pod
		;;
	"pod rm my-pod")
		touch rm
		echo my-pod	
		;;
	"volume rm volume1")
		touch volume1
		;;
	"volume rm volume2")
		touch volume2
		;;
esac`)
				err := os.WriteFile("podman.fake.sh", script, 0755)
				if err != nil {
					t.Fatal(err)
				}
			},
			checkResult: func() {
				_, err := os.Stat("stop")
				if err != nil {
					t.Errorf("podman stop has not been called")
				}
				_, err = os.Stat("rm")
				if err != nil {
					t.Errorf("podman rm has not been called")
				}
				_, err = os.Stat("volume1")
				if err != nil {
					t.Errorf("podman rm volume volume1 has not been called")
				}
				_, err = os.Stat("volume2")
				if err != nil {
					t.Errorf("podman rm volume volume2 has not been called")
				}
			},
		}, // TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.populateFS != nil {
				originWd, err := os.Getwd()
				if err != nil {
					t.Fatal(err)
				}
				defer func() {
					_ = os.Chdir(originWd)
				}()
				cwd := t.TempDir()
				err = os.Chdir(cwd)
				if err != nil {
					t.Fatal(err)
				}
				tt.populateFS()
			}

			o := &PodmanCli{
				podmanCmd:                   tt.fields.podmanCmd,
				podmanCmdInitTimeout:        tt.fields.podmanCmdInitTimeout,
				containerRunGlobalExtraArgs: tt.fields.containerRunGlobalExtraArgs,
				containerRunExtraArgs:       tt.fields.containerRunExtraArgs,
			}
			if err := o.CleanupPodResources(tt.args.pod(), tt.args.cleanupVolumes); (err != nil) != tt.wantErr {
				t.Errorf("PodmanCli.CleanupPodResources() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.checkResult != nil {
				tt.checkResult()
			}
		})
	}
}

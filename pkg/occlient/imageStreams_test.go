package occlient

import (
	"fmt"
	"reflect"
	"testing"

	imagev1 "github.com/openshift/api/image/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ktesting "k8s.io/client-go/testing"
)

func TestGetImageStream(t *testing.T) {
	tests := []struct {
		name           string
		imageNS        string
		imageName      string
		imageTag       string
		wantErr        bool
		want           *imagev1.ImageStream
		wantActionsCnt int
	}{
		{
			name:           "Case: Valid request for imagestream of latest version and not namespace qualified",
			imageNS:        "",
			imageName:      "foo",
			imageTag:       "latest",
			want:           fakeImageStream("foo", "testing", []string{"latest"}),
			wantActionsCnt: 1,
		},
		{
			name:           "Case: Valid explicit request for specific namespace qualified imagestream of specific version",
			imageNS:        "openshift",
			imageName:      "foo",
			imageTag:       "latest",
			want:           fakeImageStream("foo", "openshift", []string{"latest", "3.5"}),
			wantActionsCnt: 1,
		},
		{
			name:           "Case: Valid request for specific imagestream of specific version not in current namespace",
			imageNS:        "",
			imageName:      "foo",
			imageTag:       "3.5",
			want:           fakeImageStream("foo", "openshift", []string{"latest", "3.5"}),
			wantActionsCnt: 1, // Ideally supposed to be 2 but bcoz prependreactor is not parameter sensitive, the way it is mocked makes it 1
		},
		{
			name:           "Case: Invalid request for non-current and non-openshift namespace imagestream/Non-existant imagestream",
			imageNS:        "foo",
			imageName:      "bar",
			imageTag:       "3.5",
			wantErr:        true,
			wantActionsCnt: 1,
		},
		{
			name:           "Case: Request for non-existant tag",
			imageNS:        "",
			imageName:      "foo",
			imageTag:       "3.6",
			wantErr:        true,
			wantActionsCnt: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fkclient, fkclientset := FakeNew()
			fkclient.Namespace = "testing"
			openshiftIS := fakeImageStream(tt.imageName, "openshift", []string{"latest", "3.5"})
			currentNSIS := fakeImageStream(tt.imageName, "testing", []string{"latest"})

			fkclientset.ImageClientset.PrependReactor("get", "imagestreams", func(action ktesting.Action) (bool, runtime.Object, error) {
				if tt.imageNS == "" {
					if isTagInImageStream(*fakeImageStream("foo", "testing", []string{"latest"}), tt.imageTag) {
						return true, currentNSIS, nil
					} else if isTagInImageStream(*fakeImageStream("foo", "openshift", []string{"latest", "3.5"}), tt.imageTag) {
						return true, openshiftIS, nil
					}
					return true, nil, fmt.Errorf("Requested imagestream %s with tag %s not found", tt.imageName, tt.imageTag)
				}
				if tt.imageNS == "testing" {
					return true, currentNSIS, nil
				}
				if tt.imageNS == "openshift" {
					return true, openshiftIS, nil
				}
				return true, nil, fmt.Errorf("Requested imagestream %s with tag %s not found", tt.imageName, tt.imageTag)
			})

			got, err := fkclient.GetImageStream(tt.imageNS, tt.imageName, tt.imageTag)
			if len(fkclientset.ImageClientset.Actions()) != tt.wantActionsCnt {
				t.Errorf("expected %d ImageClientset.Actions() in GetImageStream, got %v", tt.wantActionsCnt, fkclientset.ImageClientset.Actions())
			}
			if !tt.wantErr == (err != nil) {
				t.Errorf("\nclient.GetImageStream(imageNS, imageName, imageTag) unexpected error %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetImageStream() = %#v, want %#v and the current project name is %s\n\n", got, tt, fkclient.GetCurrentProjectName())
			}
		})
	}
}

func TestListImageStreams(t *testing.T) {

	type args struct {
		name      string
		namespace string
	}

	tests := []struct {
		name    string
		args    args
		want    []imagev1.ImageStream
		wantErr bool
	}{
		{
			name: "case 1: testing a valid imagestream",
			args: args{
				name:      "ruby",
				namespace: "testing",
			},
			want: []imagev1.ImageStream{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ruby",
						Namespace: "testing",
					},
					Status: imagev1.ImageStreamStatus{
						Tags: []imagev1.NamedTagEventList{
							{
								Tag: "latest",
								Items: []imagev1.TagEvent{
									{
										DockerImageReference: "example/ruby:latest",
										Generation:           1,
										Image:                "sha256:9579a93ee",
									},
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},

		// TODO: Currently fails. Enable once fixed
		// {
		//         name: "case 2: empty namespace",
		//         args: args{
		//                 name:      "ruby",
		//                 namespace: "",
		//         },
		//         wantErr: true,
		// },

		// {
		// 	name: "case 3: empty name",
		// 	args: args{
		// 		name:      "",
		// 		namespace: "testing",
		// 	},
		// 	wantErr: true,
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			client, fkclientset := FakeNew()

			fkclientset.ImageClientset.PrependReactor("list", "imagestreams", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, fakeImageStreams(tt.args.name, tt.args.namespace), nil
			})

			got, err := client.ListImageStreams(tt.args.namespace)

			if (err != nil) != tt.wantErr {
				t.Errorf("ListImageStreams() error = %#v, wantErr %#v", err, tt.wantErr)
				return
			}

			if len(fkclientset.ImageClientset.Actions()) != 1 {
				t.Errorf("expected 1 action in ListImageStreams got: %v", fkclientset.ImageClientset.Actions())
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ListImageStreams() = %#v, want %#v", got, tt.want)
			}

		})
	}
}

func TestGetPortsFromBuilderImage(t *testing.T) {

	type args struct {
		componentType string
	}
	tests := []struct {
		name           string
		imageNamespace string
		args           args
		want           []string
		wantErr        bool
	}{
		{
			name:           "component type: nodejs",
			imageNamespace: "openshift",
			args:           args{componentType: "nodejs"},
			want:           []string{"8080/TCP"},
			wantErr:        false,
		},
		{
			name:           "component type: php",
			imageNamespace: "openshift",
			args:           args{componentType: "php"},
			want:           []string{"8080/TCP", "8443/TCP"},
			wantErr:        false,
		},
		{
			name:           "component type: is empty",
			imageNamespace: "openshift",
			args:           args{componentType: ""},
			want:           []string{},
			wantErr:        true,
		},
		{
			name:           "component type: is invalid",
			imageNamespace: "openshift",
			args:           args{componentType: "abc"},
			want:           []string{},
			wantErr:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fkclient, fkclientset := FakeNew()
			fkclient.Namespace = "testing"
			// Fake getting image stream
			fkclientset.ImageClientset.PrependReactor("get", "imagestreams", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, fakeImageStream(tt.args.componentType, tt.imageNamespace, []string{"latest"}), nil
			})

			fkclientset.ImageClientset.PrependReactor("get", "imagestreamimages", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, fakeImageStreamImage(tt.args.componentType, tt.want, ""), nil
			})
			got, err := fkclient.GetPortsFromBuilderImage(tt.args.componentType)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.GetPortsFromBuilderImage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !sliceEqual(got, tt.want) {
				t.Errorf("Client.GetPortsFromBuilderImage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getS2IMetaInfoFromBuilderImg(t *testing.T) {
	tests := []struct {
		name             string
		imageStreamImage *imagev1.ImageStreamImage
		want             S2IPaths
		wantErr          bool
	}{
		{
			name: "Case 1: Valid nodejs test case with image protocol access",
			imageStreamImage: fakeImageStreamImage(
				"nodejs",
				[]string{"8080/tcp"},
				`{
					"kind": "DockerImage",
					"apiVersion": "1.0",
					"Id": "sha256:93de1230c12b512ebbaf28b159f450a44c632eda06bdc0754236f403f5876234",
					"Created": "2018-10-19T15:43:13Z",
					"ContainerConfig": {
						"Hostname": "8911994b686d",
						"User": "1001",
						"ExposedPorts": {
							"8080/tcp": {}
						},
						"Env": [
							"PATH=/opt/app-root/src/node_modules/.bin/:/opt/app-root/src/.npm-global/bin/:/opt/app-root/src/bin:/opt/app-root/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
							"SUMMARY=Platform for building and running Node.js 10.12.0 applications",
							"DESCRIPTION=Node.js  available as docker container is a base platform for building and running various Node.js  applications and frameworks. Node.js is a platform built on Chrome's JavaScript runtime for easily building fast, scalable network applications. Node.js uses an event-driven, non-blocking I/O model that makes it lightweight and efficient, perfect for data-intensive real-time applications that run across distributed devices.",
							"STI_SCRIPTS_URL=image:///usr/libexec/s2i",
							"STI_SCRIPTS_PATH=/usr/libexec/s2i",
							"APP_ROOT=/opt/app-root",
							"HOME=/opt/app-root/src",
							"BASH_ENV=/opt/app-root/etc/scl_enable",
							"ENV=/opt/app-root/etc/scl_enable",
							"PROMPT_COMMAND=. /opt/app-root/etc/scl_enable",
							"NODEJS_SCL=rh-nodejs8",
							"NPM_RUN=start",
							"NODE_VERSION=10.12.0",
							"NPM_VERSION=6.4.1",
							"NODE_LTS=false",
							"NPM_CONFIG_LOGLEVEL=info",
							"NPM_CONFIG_PREFIX=/opt/app-root/src/.npm-global",
							"NPM_CONFIG_TARBALL=/usr/share/node/node-v10.12.0-headers.tar.gz",
							"DEBUG_PORT=5858"
						],
						"Cmd": [
							"/bin/sh",
							"-c",
							"#(nop) ",
							"CMD [\"/bin/sh\" \"-c\" \"${STI_SCRIPTS_PATH}/usage\"]"
						],
						"Image": "sha256:d353b3f467c2d2ff59a1e09bb91cff1c493aedd6c8041ebb273346e892f8ee85",
						"WorkingDir": "/opt/app-root/src",
						"Entrypoint": [
							"container-entrypoint"
						],
						"Labels": {
							"com.redhat.component": "s2i-base-container",
							"com.redhat.deployments-dir": "/opt/app-root/src",
							"com.redhat.dev-mode": "DEV_MODE:false",
							"com.rehdat.dev-mode.port": "DEBUG_PORT:5858",
							"description": "Node.js  available as docker container is a base platform for building and running various Node.js  applications and frameworks. Node.js is a platform built on Chrome's JavaScript runtime for easily building fast, scalable network applications. Node.js uses an event-driven, non-blocking I/O model that makes it lightweight and efficient, perfect for data-intensive real-time applications that run across distributed devices.",
							"io.k8s.description": "Node.js  available as docker container is a base platform for building and running various Node.js  applications and frameworks. Node.js is a platform built on Chrome's JavaScript runtime for easily building fast, scalable network applications. Node.js uses an event-driven, non-blocking I/O model that makes it lightweight and efficient, perfect for data-intensive real-time applications that run across distributed devices.",
							"io.k8s.display-name": "Node.js 10.12.0",
							"io.openshift.builder-version": "\"190ef14\"",
							"io.openshift.expose-services": "8080:http",
							"io.openshift.s2i.scripts-url": "image:///usr/libexec/s2i",
							"io.openshift.tags": "builder,nodejs,nodejs-10.12.0",
							"io.s2i.scripts-url": "image:///usr/libexec/s2i",
							"maintainer": "Lance Ball \u003clball@redhat.com\u003e",
							"name": "bucharestgold/centos7-s2i-nodejs",
							"org.label-schema.build-date": "20180804",
							"org.label-schema.license": "GPLv2",
							"org.label-schema.name": "CentOS Base Image",
							"org.label-schema.schema-version": "1.0",
							"org.label-schema.vendor": "CentOS",
							"release": "1",
							"summary": "Platform for building and running Node.js 10.12.0 applications",
							"version": "10.12.0"
						}
					},
					"DockerVersion": "18.06.0-ce",
					"Config": {
						"Hostname": "8911994b686d",
						"User": "1001",
						"ExposedPorts": {
							"8080/tcp": {}
						},
						"Env": [
							"PATH=/opt/app-root/src/node_modules/.bin/:/opt/app-root/src/.npm-global/bin/:/opt/app-root/src/bin:/opt/app-root/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
							"SUMMARY=Platform for building and running Node.js 10.12.0 applications",
							"DESCRIPTION=Node.js  available as docker container is a base platform for building and running various Node.js  applications and frameworks. Node.js is a platform built on Chrome's JavaScript runtime for easily building fast, scalable network applications. Node.js uses an event-driven, non-blocking I/O model that makes it lightweight and efficient, perfect for data-intensive real-time applications that run across distributed devices.",
							"STI_SCRIPTS_URL=image:///usr/libexec/s2i",
							"STI_SCRIPTS_PATH=/usr/libexec/s2i",
							"APP_ROOT=/opt/app-root",
							"HOME=/opt/app-root/src",
							"BASH_ENV=/opt/app-root/etc/scl_enable",
							"ENV=/opt/app-root/etc/scl_enable",
							"PROMPT_COMMAND=. /opt/app-root/etc/scl_enable",
							"NODEJS_SCL=rh-nodejs8",
							"NPM_RUN=start",
							"NODE_VERSION=10.12.0",
							"NPM_VERSION=6.4.1",
							"NODE_LTS=false",
							"NPM_CONFIG_LOGLEVEL=info",
							"NPM_CONFIG_PREFIX=/opt/app-root/src/.npm-global",
							"NPM_CONFIG_TARBALL=/usr/share/node/node-v10.12.0-headers.tar.gz",
							"DEBUG_PORT=5858"
						],
						"Cmd": [
							"/bin/sh",
							"-c",
							"${STI_SCRIPTS_PATH}/usage"
						],
						"Image": "57a00ab03a3f8c3af19e91c284e0c499b16d5115c925aa845b5d14eace949c34",
						"WorkingDir": "/opt/app-root/src",
						"Entrypoint": [
							"container-entrypoint"
						],
						"Labels": {
							"com.redhat.component": "s2i-base-container",
							"com.redhat.deployments-dir": "/opt/app-root/src",
							"com.redhat.dev-mode": "DEV_MODE:false",
							"com.rehdat.dev-mode.port": "DEBUG_PORT:5858",
							"description": "Node.js  available as docker container is a base platform for building and running various Node.js  applications and frameworks. Node.js is a platform built on Chrome's JavaScript runtime for easily building fast, scalable network applications. Node.js uses an event-driven, non-blocking I/O model that makes it lightweight and efficient, perfect for data-intensive real-time applications that run across distributed devices.",
							"io.k8s.description": "Node.js  available as docker container is a base platform for building and running various Node.js  applications and frameworks. Node.js is a platform built on Chrome's JavaScript runtime for easily building fast, scalable network applications. Node.js uses an event-driven, non-blocking I/O model that makes it lightweight and efficient, perfect for data-intensive real-time applications that run across distributed devices.",
							"io.k8s.display-name": "Node.js 10.12.0",
							"io.openshift.builder-version": "\"190ef14\"",
							"io.openshift.expose-services": "8080:http",
							"io.openshift.s2i.scripts-url": "image:///usr/libexec/s2i",
							"io.openshift.tags": "builder,nodejs,nodejs-10.12.0",
							"io.s2i.scripts-url": "image:///usr/libexec/s2i",
							"maintainer": "Lance Ball \u003clball@redhat.com\u003e",
							"name": "bucharestgold/centos7-s2i-nodejs",
							"org.label-schema.build-date": "20180804",
							"org.label-schema.license": "GPLv2",
							"org.label-schema.name": "CentOS Base Image",
							"org.label-schema.schema-version": "1.0",
							"org.label-schema.vendor": "CentOS",
							"release": "1",
							"summary": "Platform for building and running Node.js 10.12.0 applications",
							"version": "10.12.0"
						}
					},
					"Architecture": "amd64",
					"Size": 221580439
				}`,
			),
			want: S2IPaths{
				ScriptsPathProtocol: "image://",
				ScriptsPath:         "/usr/libexec/s2i",
				SrcOrBinPath:        "/tmp",
				DeploymentDir:       "/opt/app-root/src",
				WorkingDir:          "/opt/app-root/src",
				SrcBackupPath:       "/opt/app-root/src-backup",
				BuilderImgName:      "bucharestgold/centos7-s2i-nodejs",
			},
			wantErr: false,
		},
		{
			name: "Case 2: Valid nodejs test case with file protocol access",
			imageStreamImage: fakeImageStreamImage(
				"nodejs",
				[]string{"8080/tcp"},
				`{
					"kind": "DockerImage",
					"apiVersion": "1.0",
					"Id": "sha256:93de1230c12b512ebbaf28b159f450a44c632eda06bdc0754236f403f5876234",
					"Created": "2018-10-19T15:43:13Z",
					"ContainerConfig": {
						"Hostname": "8911994b686d",
						"User": "1001",
						"ExposedPorts": {
							"8080/tcp": {}
						},
						"Env": [
							"PATH=/opt/app-root/src/node_modules/.bin/:/opt/app-root/src/.npm-global/bin/:/opt/app-root/src/bin:/opt/app-root/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
							"SUMMARY=Platform for building and running Node.js 10.12.0 applications",
							"DESCRIPTION=Node.js  available as docker container is a base platform for building and running various Node.js  applications and frameworks. Node.js is a platform built on Chrome's JavaScript runtime for easily building fast, scalable network applications. Node.js uses an event-driven, non-blocking I/O model that makes it lightweight and efficient, perfect for data-intensive real-time applications that run across distributed devices.",
							"STI_SCRIPTS_URL=file:///usr/libexec/s2i",
							"STI_SCRIPTS_PATH=/usr/libexec/s2i",
							"APP_ROOT=/opt/app-root",
							"HOME=/opt/app-root/src",
							"BASH_ENV=/opt/app-root/etc/scl_enable",
							"ENV=/opt/app-root/etc/scl_enable",
							"PROMPT_COMMAND=. /opt/app-root/etc/scl_enable",
							"NODEJS_SCL=rh-nodejs8",
							"NPM_RUN=start",
							"NODE_VERSION=10.12.0",
							"NPM_VERSION=6.4.1",
							"NODE_LTS=false",
							"NPM_CONFIG_LOGLEVEL=info",
							"NPM_CONFIG_PREFIX=/opt/app-root/src/.npm-global",
							"NPM_CONFIG_TARBALL=/usr/share/node/node-v10.12.0-headers.tar.gz",
							"DEBUG_PORT=5858"
						],
						"Cmd": [
							"/bin/sh",
							"-c",
							"#(nop) ",
							"CMD [\"/bin/sh\" \"-c\" \"${STI_SCRIPTS_PATH}/usage\"]"
						],
						"Image": "sha256:d353b3f467c2d2ff59a1e09bb91cff1c493aedd6c8041ebb273346e892f8ee85",
						"WorkingDir": "/opt/app-root/src",
						"Entrypoint": [
							"container-entrypoint"
						],
						"Labels": {
							"com.redhat.component": "s2i-base-container",
							"com.redhat.deployments-dir": "/opt/app-root/src",
							"com.redhat.dev-mode": "DEV_MODE:false",
							"com.rehdat.dev-mode.port": "DEBUG_PORT:5858",
							"description": "Node.js  available as docker container is a base platform for building and running various Node.js  applications and frameworks. Node.js is a platform built on Chrome's JavaScript runtime for easily building fast, scalable network applications. Node.js uses an event-driven, non-blocking I/O model that makes it lightweight and efficient, perfect for data-intensive real-time applications that run across distributed devices.",
							"io.k8s.description": "Node.js  available as docker container is a base platform for building and running various Node.js  applications and frameworks. Node.js is a platform built on Chrome's JavaScript runtime for easily building fast, scalable network applications. Node.js uses an event-driven, non-blocking I/O model that makes it lightweight and efficient, perfect for data-intensive real-time applications that run across distributed devices.",
							"io.k8s.display-name": "Node.js 10.12.0",
							"io.openshift.builder-version": "\"190ef14\"",
							"io.openshift.expose-services": "8080:http",
							"io.openshift.s2i.scripts-url": "file:///usr/libexec/s2i",
							"io.openshift.tags": "builder,nodejs,nodejs-10.12.0",
							"io.s2i.scripts-url": "file:///usr/libexec/s2i",
							"maintainer": "Lance Ball \u003clball@redhat.com\u003e",
							"name": "bucharestgold/centos7-s2i-nodejs",
							"org.label-schema.build-date": "20180804",
							"org.label-schema.license": "GPLv2",
							"org.label-schema.name": "CentOS Base Image",
							"org.label-schema.schema-version": "1.0",
							"org.label-schema.vendor": "CentOS",
							"release": "1",
							"summary": "Platform for building and running Node.js 10.12.0 applications",
							"version": "10.12.0"
						}
					},
					"DockerVersion": "18.06.0-ce",
					"Config": {
						"Hostname": "8911994b686d",
						"User": "1001",
						"ExposedPorts": {
							"8080/tcp": {}
						},
						"Env": [
							"PATH=/opt/app-root/src/node_modules/.bin/:/opt/app-root/src/.npm-global/bin/:/opt/app-root/src/bin:/opt/app-root/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
							"SUMMARY=Platform for building and running Node.js 10.12.0 applications",
							"DESCRIPTION=Node.js  available as docker container is a base platform for building and running various Node.js  applications and frameworks. Node.js is a platform built on Chrome's JavaScript runtime for easily building fast, scalable network applications. Node.js uses an event-driven, non-blocking I/O model that makes it lightweight and efficient, perfect for data-intensive real-time applications that run across distributed devices.",
							"STI_SCRIPTS_URL=image:///usr/libexec/s2i",
							"STI_SCRIPTS_PATH=/usr/libexec/s2i",
							"APP_ROOT=/opt/app-root",
							"HOME=/opt/app-root/src",
							"BASH_ENV=/opt/app-root/etc/scl_enable",
							"ENV=/opt/app-root/etc/scl_enable",
							"PROMPT_COMMAND=. /opt/app-root/etc/scl_enable",
							"NODEJS_SCL=rh-nodejs8",
							"NPM_RUN=start",
							"NODE_VERSION=10.12.0",
							"NPM_VERSION=6.4.1",
							"NODE_LTS=false",
							"NPM_CONFIG_LOGLEVEL=info",
							"NPM_CONFIG_PREFIX=/opt/app-root/src/.npm-global",
							"NPM_CONFIG_TARBALL=/usr/share/node/node-v10.12.0-headers.tar.gz",
							"DEBUG_PORT=5858"
						],
						"Cmd": [
							"/bin/sh",
							"-c",
							"${STI_SCRIPTS_PATH}/usage"
						],
						"Image": "57a00ab03a3f8c3af19e91c284e0c499b16d5115c925aa845b5d14eace949c34",
						"WorkingDir": "/opt/app-root/src",
						"Entrypoint": [
							"container-entrypoint"
						],
						"Labels": {
							"com.redhat.component": "s2i-base-container",
							"com.redhat.deployments-dir": "/opt/app-root/src",
							"com.redhat.dev-mode": "DEV_MODE:false",
							"com.rehdat.dev-mode.port": "DEBUG_PORT:5858",
							"description": "Node.js  available as docker container is a base platform for building and running various Node.js  applications and frameworks. Node.js is a platform built on Chrome's JavaScript runtime for easily building fast, scalable network applications. Node.js uses an event-driven, non-blocking I/O model that makes it lightweight and efficient, perfect for data-intensive real-time applications that run across distributed devices.",
							"io.k8s.description": "Node.js  available as docker container is a base platform for building and running various Node.js  applications and frameworks. Node.js is a platform built on Chrome's JavaScript runtime for easily building fast, scalable network applications. Node.js uses an event-driven, non-blocking I/O model that makes it lightweight and efficient, perfect for data-intensive real-time applications that run across distributed devices.",
							"io.k8s.display-name": "Node.js 10.12.0",
							"io.openshift.builder-version": "\"190ef14\"",
							"io.openshift.expose-services": "8080:http",
							"io.openshift.s2i.scripts-url": "file:///usr/libexec/s2i",
							"io.openshift.tags": "builder,nodejs,nodejs-10.12.0",
							"io.s2i.scripts-url": "image:///usr/libexec/s2i",
							"maintainer": "Lance Ball \u003clball@redhat.com\u003e",
							"name": "bucharestgold/centos7-s2i-nodejs",
							"org.label-schema.build-date": "20180804",
							"org.label-schema.license": "GPLv2",
							"org.label-schema.name": "CentOS Base Image",
							"org.label-schema.schema-version": "1.0",
							"org.label-schema.vendor": "CentOS",
							"release": "1",
							"summary": "Platform for building and running Node.js 10.12.0 applications",
							"version": "10.12.0"
						}
					},
					"Architecture": "amd64",
					"Size": 221580439
				}`,
			),
			want: S2IPaths{
				ScriptsPathProtocol: "file://",
				ScriptsPath:         "/usr/libexec/s2i",
				SrcOrBinPath:        "/tmp",
				DeploymentDir:       "/opt/app-root/src",
				WorkingDir:          "/opt/app-root/src",
				SrcBackupPath:       "/opt/app-root/src-backup",
				BuilderImgName:      "bucharestgold/centos7-s2i-nodejs",
			},
			wantErr: false,
		},
		{
			name: "Case 3: Valid nodejs test case with http(s) protocol access",
			imageStreamImage: fakeImageStreamImage(
				"nodejs",
				[]string{"8080/tcp"},
				`{
					"kind": "DockerImage",
					"apiVersion": "1.0",
					"Id": "sha256:93de1230c12b512ebbaf28b159f450a44c632eda06bdc0754236f403f5876234",
					"Created": "2018-10-19T15:43:13Z",
					"ContainerConfig": {
						"Hostname": "8911994b686d",
						"User": "1001",
						"ExposedPorts": {
							"8080/tcp": {}
						},
						"Env": [
							"PATH=/opt/app-root/src/node_modules/.bin/:/opt/app-root/src/.npm-global/bin/:/opt/app-root/src/bin:/opt/app-root/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
							"SUMMARY=Platform for building and running Node.js 10.12.0 applications",
							"DESCRIPTION=Node.js  available as docker container is a base platform for building and running various Node.js  applications and frameworks. Node.js is a platform built on Chrome's JavaScript runtime for easily building fast, scalable network applications. Node.js uses an event-driven, non-blocking I/O model that makes it lightweight and efficient, perfect for data-intensive real-time applications that run across distributed devices.",
							"STI_SCRIPTS_URL=http(s):///usr/libexec/s2i",
							"STI_SCRIPTS_PATH=/usr/libexec/s2i",
							"APP_ROOT=/opt/app-root",
							"HOME=/opt/app-root/src",
							"BASH_ENV=/opt/app-root/etc/scl_enable",
							"ENV=/opt/app-root/etc/scl_enable",
							"PROMPT_COMMAND=. /opt/app-root/etc/scl_enable",
							"NODEJS_SCL=rh-nodejs8",
							"NPM_RUN=start",
							"NODE_VERSION=10.12.0",
							"NPM_VERSION=6.4.1",
							"NODE_LTS=false",
							"NPM_CONFIG_LOGLEVEL=info",
							"NPM_CONFIG_PREFIX=/opt/app-root/src/.npm-global",
							"NPM_CONFIG_TARBALL=/usr/share/node/node-v10.12.0-headers.tar.gz",
							"DEBUG_PORT=5858"
						],
						"Cmd": [
							"/bin/sh",
							"-c",
							"#(nop) ",
							"CMD [\"/bin/sh\" \"-c\" \"${STI_SCRIPTS_PATH}/usage\"]"
						],
						"Image": "sha256:d353b3f467c2d2ff59a1e09bb91cff1c493aedd6c8041ebb273346e892f8ee85",
						"WorkingDir": "/opt/app-root/src",
						"Entrypoint": [
							"container-entrypoint"
						],
						"Labels": {
							"com.redhat.component": "s2i-base-container",
							"com.redhat.deployments-dir": "/opt/app-root/src",
							"com.redhat.dev-mode": "DEV_MODE:false",
							"com.rehdat.dev-mode.port": "DEBUG_PORT:5858",
							"description": "Node.js  available as docker container is a base platform for building and running various Node.js  applications and frameworks. Node.js is a platform built on Chrome's JavaScript runtime for easily building fast, scalable network applications. Node.js uses an event-driven, non-blocking I/O model that makes it lightweight and efficient, perfect for data-intensive real-time applications that run across distributed devices.",
							"io.k8s.description": "Node.js  available as docker container is a base platform for building and running various Node.js  applications and frameworks. Node.js is a platform built on Chrome's JavaScript runtime for easily building fast, scalable network applications. Node.js uses an event-driven, non-blocking I/O model that makes it lightweight and efficient, perfect for data-intensive real-time applications that run across distributed devices.",
							"io.k8s.display-name": "Node.js 10.12.0",
							"io.openshift.builder-version": "\"190ef14\"",
							"io.openshift.expose-services": "8080:http",
							"io.openshift.s2i.scripts-url": "http(s):///usr/libexec/s2i",
							"io.openshift.tags": "builder,nodejs,nodejs-10.12.0",
							"io.s2i.scripts-url": "image:///usr/libexec/s2i",
							"maintainer": "Lance Ball \u003clball@redhat.com\u003e",
							"name": "bucharestgold/centos7-s2i-nodejs",
							"org.label-schema.build-date": "20180804",
							"org.label-schema.license": "GPLv2",
							"org.label-schema.name": "CentOS Base Image",
							"org.label-schema.schema-version": "1.0",
							"org.label-schema.vendor": "CentOS",
							"release": "1",
							"summary": "Platform for building and running Node.js 10.12.0 applications",
							"version": "10.12.0"
						}
					},
					"DockerVersion": "18.06.0-ce",
					"Config": {
						"Hostname": "8911994b686d",
						"User": "1001",
						"ExposedPorts": {
							"8080/tcp": {}
						},
						"Env": [
							"PATH=/opt/app-root/src/node_modules/.bin/:/opt/app-root/src/.npm-global/bin/:/opt/app-root/src/bin:/opt/app-root/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
							"SUMMARY=Platform for building and running Node.js 10.12.0 applications",
							"DESCRIPTION=Node.js  available as docker container is a base platform for building and running various Node.js  applications and frameworks. Node.js is a platform built on Chrome's JavaScript runtime for easily building fast, scalable network applications. Node.js uses an event-driven, non-blocking I/O model that makes it lightweight and efficient, perfect for data-intensive real-time applications that run across distributed devices.",
							"STI_SCRIPTS_URL=image:///usr/libexec/s2i",
							"STI_SCRIPTS_PATH=/usr/libexec/s2i",
							"APP_ROOT=/opt/app-root",
							"HOME=/opt/app-root/src",
							"BASH_ENV=/opt/app-root/etc/scl_enable",
							"ENV=/opt/app-root/etc/scl_enable",
							"PROMPT_COMMAND=. /opt/app-root/etc/scl_enable",
							"NODEJS_SCL=rh-nodejs8",
							"NPM_RUN=start",
							"NODE_VERSION=10.12.0",
							"NPM_VERSION=6.4.1",
							"NODE_LTS=false",
							"NPM_CONFIG_LOGLEVEL=info",
							"NPM_CONFIG_PREFIX=/opt/app-root/src/.npm-global",
							"NPM_CONFIG_TARBALL=/usr/share/node/node-v10.12.0-headers.tar.gz",
							"DEBUG_PORT=5858"
						],
						"Cmd": [
							"/bin/sh",
							"-c",
							"${STI_SCRIPTS_PATH}/usage"
						],
						"Image": "57a00ab03a3f8c3af19e91c284e0c499b16d5115c925aa845b5d14eace949c34",
						"WorkingDir": "/opt/app-root/src",
						"Entrypoint": [
							"container-entrypoint"
						],
						"Labels": {
							"com.redhat.component": "s2i-base-container",
							"com.redhat.deployments-dir": "/opt/app-root/src",
							"com.redhat.dev-mode": "DEV_MODE:false",
							"com.rehdat.dev-mode.port": "DEBUG_PORT:5858",
							"description": "Node.js  available as docker container is a base platform for building and running various Node.js  applications and frameworks. Node.js is a platform built on Chrome's JavaScript runtime for easily building fast, scalable network applications. Node.js uses an event-driven, non-blocking I/O model that makes it lightweight and efficient, perfect for data-intensive real-time applications that run across distributed devices.",
							"io.k8s.description": "Node.js  available as docker container is a base platform for building and running various Node.js  applications and frameworks. Node.js is a platform built on Chrome's JavaScript runtime for easily building fast, scalable network applications. Node.js uses an event-driven, non-blocking I/O model that makes it lightweight and efficient, perfect for data-intensive real-time applications that run across distributed devices.",
							"io.k8s.display-name": "Node.js 10.12.0",
							"io.openshift.builder-version": "\"190ef14\"",
							"io.openshift.expose-services": "8080:http",
							"io.openshift.s2i.scripts-url": "http(s):///usr/libexec/s2i",
							"io.openshift.tags": "builder,nodejs,nodejs-10.12.0",
							"io.s2i.scripts-url": "image:///usr/libexec/s2i",
							"maintainer": "Lance Ball \u003clball@redhat.com\u003e",
							"name": "bucharestgold/centos7-s2i-nodejs",
							"org.label-schema.build-date": "20180804",
							"org.label-schema.license": "GPLv2",
							"org.label-schema.name": "CentOS Base Image",
							"org.label-schema.schema-version": "1.0",
							"org.label-schema.vendor": "CentOS",
							"release": "1",
							"summary": "Platform for building and running Node.js 10.12.0 applications",
							"version": "10.12.0"
						}
					},
					"Architecture": "amd64",
					"Size": 221580439
				}`,
			),
			want: S2IPaths{
				ScriptsPathProtocol: "http(s)://",
				ScriptsPath:         "http(s):///usr/libexec/s2i",
				SrcOrBinPath:        "/tmp",
				DeploymentDir:       "/opt/app-root/src",
				WorkingDir:          "/opt/app-root/src",
				SrcBackupPath:       "/opt/app-root/src-backup",
				BuilderImgName:      "bucharestgold/centos7-s2i-nodejs",
			},
			wantErr: false,
		},
		{
			name: "Case 4: Valid openjdk test case with image(s) protocol access",
			imageStreamImage: fakeImageStreamImage(
				"redhat-openjdk18-openshift",
				[]string{"8080/tcp"},
				`{
					"kind": "DockerImage",
					"apiVersion": "1.0",
					"Id": "sha256:93de1230c12b512ebbaf28b159f450a44c632eda06bdc0754236f403f5876234",
					"Created": "2018-10-19T15:43:13Z",
					"ContainerConfig": {
						"Hostname": "8911994b686d",
						"User": "1001",
						"ExposedPorts": {
							"8080/tcp": {}
						},
						"Env": [
							"PATH=/opt/app-root/src/node_modules/.bin/:/opt/app-root/src/.npm-global/bin/:/opt/app-root/src/bin:/opt/app-root/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
							"SUMMARY=Platform for building and running Node.js 10.12.0 applications",
							"DESCRIPTION=Node.js  available as docker container is a base platform for building and running various Node.js  applications and frameworks. Node.js is a platform built on Chrome's JavaScript runtime for easily building fast, scalable network applications. Node.js uses an event-driven, non-blocking I/O model that makes it lightweight and efficient, perfect for data-intensive real-time applications that run across distributed devices.",
							"STI_SCRIPTS_URL=http(s):///usr/libexec/s2i",
							"STI_SCRIPTS_PATH=/usr/libexec/s2i",
							"APP_ROOT=/opt/app-root",
							"HOME=/opt/app-root/src",
							"BASH_ENV=/opt/app-root/etc/scl_enable",
							"ENV=/opt/app-root/etc/scl_enable",
							"PROMPT_COMMAND=. /opt/app-root/etc/scl_enable",
							"NODEJS_SCL=rh-nodejs8",
							"NPM_RUN=start",
							"NODE_VERSION=10.12.0",
							"NPM_VERSION=6.4.1",
							"NODE_LTS=false",
							"NPM_CONFIG_LOGLEVEL=info",
							"NPM_CONFIG_PREFIX=/opt/app-root/src/.npm-global",
							"NPM_CONFIG_TARBALL=/usr/share/node/node-v10.12.0-headers.tar.gz",
							"DEBUG_PORT=5858"
						],
						"Cmd": [
							"/bin/sh",
							"-c",
							"#(nop) ",
							"CMD [\"/bin/sh\" \"-c\" \"${STI_SCRIPTS_PATH}/usage\"]"
						],
						"Image": "sha256:d353b3f467c2d2ff59a1e09bb91cff1c493aedd6c8041ebb273346e892f8ee85",
						"WorkingDir": "/opt/app-root/src",
						"Entrypoint": [
							"container-entrypoint"
						],
						"Labels": {
							"com.redhat.component": "s2i-base-container",
							"com.redhat.deployments-dir": "/opt/app-root/src",
							"com.redhat.dev-mode": "DEV_MODE:false",
							"com.rehdat.dev-mode.port": "DEBUG_PORT:5858",
							"description": "Node.js  available as docker container is a base platform for building and running various Node.js  applications and frameworks. Node.js is a platform built on Chrome's JavaScript runtime for easily building fast, scalable network applications. Node.js uses an event-driven, non-blocking I/O model that makes it lightweight and efficient, perfect for data-intensive real-time applications that run across distributed devices.",
							"io.k8s.description": "Node.js  available as docker container is a base platform for building and running various Node.js  applications and frameworks. Node.js is a platform built on Chrome's JavaScript runtime for easily building fast, scalable network applications. Node.js uses an event-driven, non-blocking I/O model that makes it lightweight and efficient, perfect for data-intensive real-time applications that run across distributed devices.",
							"io.k8s.display-name": "Node.js 10.12.0",
							"io.openshift.builder-version": "\"190ef14\"",
							"io.openshift.expose-services": "8080:http",
							"io.openshift.s2i.scripts-url": "image:///usr/libexec/s2i",
							"io.openshift.tags": "builder,nodejs,nodejs-10.12.0",
							"io.s2i.scripts-url": "image:///usr/libexec/s2i",
							"maintainer": "Lance Ball \u003clball@redhat.com\u003e",
							"name": "bucharestgold/centos7-s2i-nodejs",
							"org.label-schema.build-date": "20180804",
							"org.label-schema.license": "GPLv2",
							"org.label-schema.name": "CentOS Base Image",
							"org.label-schema.schema-version": "1.0",
							"org.label-schema.vendor": "CentOS",
							"release": "1",
							"summary": "Platform for building and running Node.js 10.12.0 applications",
							"version": "10.12.0"
						}
					},
					"DockerVersion": "18.06.0-ce",
					"Config": {
						"Hostname": "8911994b686d",
						"User": "1001",
						"ExposedPorts": {
							"8080/tcp": {}
						},
						"Env": [
							"PATH=/opt/app-root/src/node_modules/.bin/:/opt/app-root/src/.npm-global/bin/:/opt/app-root/src/bin:/opt/app-root/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
							"SUMMARY=Platform for building and running Node.js 10.12.0 applications",
							"DESCRIPTION=Node.js  available as docker container is a base platform for building and running various Node.js  applications and frameworks. Node.js is a platform built on Chrome's JavaScript runtime for easily building fast, scalable network applications. Node.js uses an event-driven, non-blocking I/O model that makes it lightweight and efficient, perfect for data-intensive real-time applications that run across distributed devices.",
							"STI_SCRIPTS_URL=image:///usr/libexec/s2i",
							"STI_SCRIPTS_PATH=/usr/libexec/s2i",
							"APP_ROOT=/opt/app-root",
							"HOME=/opt/app-root/src",
							"BASH_ENV=/opt/app-root/etc/scl_enable",
							"ENV=/opt/app-root/etc/scl_enable",
							"PROMPT_COMMAND=. /opt/app-root/etc/scl_enable",
							"NODEJS_SCL=rh-nodejs8",
							"NPM_RUN=start",
							"NODE_VERSION=10.12.0",
							"NPM_VERSION=6.4.1",
							"NODE_LTS=false",
							"NPM_CONFIG_LOGLEVEL=info",
							"NPM_CONFIG_PREFIX=/opt/app-root/src/.npm-global",
							"NPM_CONFIG_TARBALL=/usr/share/node/node-v10.12.0-headers.tar.gz",
							"DEBUG_PORT=5858"
						],
						"Cmd": [
							"/bin/sh",
							"-c",
							"${STI_SCRIPTS_PATH}/usage"
						],
						"Image": "57a00ab03a3f8c3af19e91c284e0c499b16d5115c925aa845b5d14eace949c34",
						"WorkingDir": "/opt/app-root/src",
						"Entrypoint": [
							"container-entrypoint"
						],
						"Labels": {
							"com.redhat.component": "s2i-base-container",
							"com.redhat.deployments-dir": "/opt/app-root/src",
							"com.redhat.dev-mode": "DEV_MODE:false",
							"com.rehdat.dev-mode.port": "DEBUG_PORT:5858",
							"description": "Node.js  available as docker container is a base platform for building and running various Node.js  applications and frameworks. Node.js is a platform built on Chrome's JavaScript runtime for easily building fast, scalable network applications. Node.js uses an event-driven, non-blocking I/O model that makes it lightweight and efficient, perfect for data-intensive real-time applications that run across distributed devices.",
							"io.k8s.description": "Node.js  available as docker container is a base platform for building and running various Node.js  applications and frameworks. Node.js is a platform built on Chrome's JavaScript runtime for easily building fast, scalable network applications. Node.js uses an event-driven, non-blocking I/O model that makes it lightweight and efficient, perfect for data-intensive real-time applications that run across distributed devices.",
							"io.k8s.display-name": "Node.js 10.12.0",
							"io.openshift.builder-version": "\"190ef14\"",
							"io.openshift.expose-services": "8080:http",
							"io.openshift.s2i.scripts-url": "image:///usr/libexec/s2i",
							"io.openshift.tags": "builder,nodejs,nodejs-10.12.0",
							"io.s2i.scripts-url": "image:///usr/libexec/s2i",
							"maintainer": "Lance Ball \u003clball@redhat.com\u003e",
							"name": "bucharestgold/centos7-s2i-nodejs",
							"org.label-schema.build-date": "20180804",
							"org.label-schema.license": "GPLv2",
							"org.label-schema.name": "CentOS Base Image",
							"org.label-schema.schema-version": "1.0",
							"org.label-schema.vendor": "CentOS",
							"release": "1",
							"summary": "Platform for building and running Node.js 10.12.0 applications",
							"version": "10.12.0"
						}
					},
					"Architecture": "amd64",
					"Size": 221580439
				}`,
			),
			want: S2IPaths{
				ScriptsPathProtocol: "image://",
				ScriptsPath:         "/usr/libexec/s2i",
				SrcOrBinPath:        "/tmp",
				DeploymentDir:       "/opt/app-root/src",
				WorkingDir:          "/opt/app-root/src",
				SrcBackupPath:       "/opt/app-root/src-backup",
				BuilderImgName:      "bucharestgold/centos7-s2i-nodejs",
			},
			wantErr: false,
		},
		{
			name: "Case 5: Inalid nodejs test case with invalid protocol access",
			imageStreamImage: fakeImageStreamImage(
				"nodejs",
				[]string{"8080/tcp"},
				`{
					"kind": "DockerImage",
					"apiVersion": "1.0",
					"Id": "sha256:93de1230c12b512ebbaf28b159f450a44c632eda06bdc0754236f403f5876234",
					"Created": "2018-10-19T15:43:13Z",
					"ContainerConfig": {
						"Hostname": "8911994b686d",
						"User": "1001",
						"ExposedPorts": {
							"8080/tcp": {}
						},
						"Env": [
							"PATH=/opt/app-root/src/node_modules/.bin/:/opt/app-root/src/.npm-global/bin/:/opt/app-root/src/bin:/opt/app-root/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
							"SUMMARY=Platform for building and running Node.js 10.12.0 applications",
							"DESCRIPTION=Node.js  available as docker container is a base platform for building and running various Node.js  applications and frameworks. Node.js is a platform built on Chrome's JavaScript runtime for easily building fast, scalable network applications. Node.js uses an event-driven, non-blocking I/O model that makes it lightweight and efficient, perfect for data-intensive real-time applications that run across distributed devices.",
							"STI_SCRIPTS_URL=something:///usr/libexec/s2i",
							"STI_SCRIPTS_PATH=/usr/libexec/s2i",
							"APP_ROOT=/opt/app-root",
							"HOME=/opt/app-root/src",
							"BASH_ENV=/opt/app-root/etc/scl_enable",
							"ENV=/opt/app-root/etc/scl_enable",
							"PROMPT_COMMAND=. /opt/app-root/etc/scl_enable",
							"NODEJS_SCL=rh-nodejs8",
							"NPM_RUN=start",
							"NODE_VERSION=10.12.0",
							"NPM_VERSION=6.4.1",
							"NODE_LTS=false",
							"NPM_CONFIG_LOGLEVEL=info",
							"NPM_CONFIG_PREFIX=/opt/app-root/src/.npm-global",
							"NPM_CONFIG_TARBALL=/usr/share/node/node-v10.12.0-headers.tar.gz",
							"DEBUG_PORT=5858"
						],
						"Cmd": [
							"/bin/sh",
							"-c",
							"#(nop) ",
							"CMD [\"/bin/sh\" \"-c\" \"${STI_SCRIPTS_PATH}/usage\"]"
						],
						"Image": "sha256:d353b3f467c2d2ff59a1e09bb91cff1c493aedd6c8041ebb273346e892f8ee85",
						"WorkingDir": "/opt/app-root/src",
						"Entrypoint": [
							"container-entrypoint"
						],
						"Labels": {
							"com.redhat.component": "s2i-base-container",
							"com.redhat.deployments-dir": "/opt/app-root/src",
							"com.redhat.dev-mode": "DEV_MODE:false",
							"com.rehdat.dev-mode.port": "DEBUG_PORT:5858",
							"description": "Node.js  available as docker container is a base platform for building and running various Node.js  applications and frameworks. Node.js is a platform built on Chrome's JavaScript runtime for easily building fast, scalable network applications. Node.js uses an event-driven, non-blocking I/O model that makes it lightweight and efficient, perfect for data-intensive real-time applications that run across distributed devices.",
							"io.k8s.description": "Node.js  available as docker container is a base platform for building and running various Node.js  applications and frameworks. Node.js is a platform built on Chrome's JavaScript runtime for easily building fast, scalable network applications. Node.js uses an event-driven, non-blocking I/O model that makes it lightweight and efficient, perfect for data-intensive real-time applications that run across distributed devices.",
							"io.k8s.display-name": "Node.js 10.12.0",
							"io.openshift.builder-version": "\"190ef14\"",
							"io.openshift.expose-services": "8080:http",
							"io.openshift.s2i.scripts-url": "something:///usr/libexec/s2i",
							"io.openshift.tags": "builder,nodejs,nodejs-10.12.0",
							"io.s2i.scripts-url": "image:///usr/libexec/s2i",
							"maintainer": "Lance Ball \u003clball@redhat.com\u003e",
							"name": "bucharestgold/centos7-s2i-nodejs",
							"org.label-schema.build-date": "20180804",
							"org.label-schema.license": "GPLv2",
							"org.label-schema.name": "CentOS Base Image",
							"org.label-schema.schema-version": "1.0",
							"org.label-schema.vendor": "CentOS",
							"release": "1",
							"summary": "Platform for building and running Node.js 10.12.0 applications",
							"version": "10.12.0"
						}
					},
					"DockerVersion": "18.06.0-ce",
					"Config": {
						"Hostname": "8911994b686d",
						"User": "1001",
						"ExposedPorts": {
							"8080/tcp": {}
						},
						"Env": [
							"PATH=/opt/app-root/src/node_modules/.bin/:/opt/app-root/src/.npm-global/bin/:/opt/app-root/src/bin:/opt/app-root/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
							"SUMMARY=Platform for building and running Node.js 10.12.0 applications",
							"DESCRIPTION=Node.js  available as docker container is a base platform for building and running various Node.js  applications and frameworks. Node.js is a platform built on Chrome's JavaScript runtime for easily building fast, scalable network applications. Node.js uses an event-driven, non-blocking I/O model that makes it lightweight and efficient, perfect for data-intensive real-time applications that run across distributed devices.",
							"STI_SCRIPTS_URL=image:///usr/libexec/s2i",
							"STI_SCRIPTS_PATH=/usr/libexec/s2i",
							"APP_ROOT=/opt/app-root",
							"HOME=/opt/app-root/src",
							"BASH_ENV=/opt/app-root/etc/scl_enable",
							"ENV=/opt/app-root/etc/scl_enable",
							"PROMPT_COMMAND=. /opt/app-root/etc/scl_enable",
							"NODEJS_SCL=rh-nodejs8",
							"NPM_RUN=start",
							"NODE_VERSION=10.12.0",
							"NPM_VERSION=6.4.1",
							"NODE_LTS=false",
							"NPM_CONFIG_LOGLEVEL=info",
							"NPM_CONFIG_PREFIX=/opt/app-root/src/.npm-global",
							"NPM_CONFIG_TARBALL=/usr/share/node/node-v10.12.0-headers.tar.gz",
							"DEBUG_PORT=5858"
						],
						"Cmd": [
							"/bin/sh",
							"-c",
							"${STI_SCRIPTS_PATH}/usage"
						],
						"Image": "57a00ab03a3f8c3af19e91c284e0c499b16d5115c925aa845b5d14eace949c34",
						"WorkingDir": "/opt/app-root/src",
						"Entrypoint": [
							"container-entrypoint"
						],
						"Labels": {
							"com.redhat.component": "s2i-base-container",
							"com.redhat.deployments-dir": "/opt/app-root/src",
							"com.redhat.dev-mode": "DEV_MODE:false",
							"com.rehdat.dev-mode.port": "DEBUG_PORT:5858",
							"description": "Node.js  available as docker container is a base platform for building and running various Node.js  applications and frameworks. Node.js is a platform built on Chrome's JavaScript runtime for easily building fast, scalable network applications. Node.js uses an event-driven, non-blocking I/O model that makes it lightweight and efficient, perfect for data-intensive real-time applications that run across distributed devices.",
							"io.k8s.description": "Node.js  available as docker container is a base platform for building and running various Node.js  applications and frameworks. Node.js is a platform built on Chrome's JavaScript runtime for easily building fast, scalable network applications. Node.js uses an event-driven, non-blocking I/O model that makes it lightweight and efficient, perfect for data-intensive real-time applications that run across distributed devices.",
							"io.k8s.display-name": "Node.js 10.12.0",
							"io.openshift.builder-version": "\"190ef14\"",
							"io.openshift.expose-services": "8080:http",
							"io.openshift.s2i.scripts-url": "something:///usr/libexec/s2i",
							"io.openshift.tags": "builder,nodejs,nodejs-10.12.0",
							"io.s2i.scripts-url": "image:///usr/libexec/s2i",
							"maintainer": "Lance Ball \u003clball@redhat.com\u003e",
							"name": "bucharestgold/centos7-s2i-nodejs",
							"org.label-schema.build-date": "20180804",
							"org.label-schema.license": "GPLv2",
							"org.label-schema.name": "CentOS Base Image",
							"org.label-schema.schema-version": "1.0",
							"org.label-schema.vendor": "CentOS",
							"release": "1",
							"summary": "Platform for building and running Node.js 10.12.0 applications",
							"version": "10.12.0"
						}
					},
					"Architecture": "amd64",
					"Size": 221580439
				}`,
			),
			want:    S2IPaths{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s2iPaths, err := getS2IMetaInfoFromBuilderImg(tt.imageStreamImage)
			if !reflect.DeepEqual(tt.want, s2iPaths) {
				t.Errorf("s2i paths are not matching with expected values, expected: %v, got %v", tt.want, s2iPaths)
			}
			if !tt.wantErr == (err != nil) {
				t.Errorf(" GetS2IScriptsPathFromBuilderImg() unexpected error %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

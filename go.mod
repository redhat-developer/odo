module github.com/openshift/odo

go 1.13

require (
	cloud.google.com/go v0.45.1 // indirect
	github.com/Azure/go-autorest/autorest v0.11.4 // indirect
	github.com/Microsoft/go-winio v0.4.15 // indirect
	github.com/Netflix/go-expect v0.0.0-20200312175327-da48e75238e2
	github.com/blang/semver v3.5.1+incompatible
	github.com/containerd/containerd v1.3.3 // indirect
	github.com/devfile/api/v2 v2.0.0-20210304212617-bfc3f501616b
	github.com/devfile/library v1.0.0-alpha.2.0.20210323153322-3d708859f0b5
	github.com/docker/docker v17.12.0-ce-rc1.0.20200309214505-aa6a9891b09c+incompatible
	github.com/docker/go-connections v0.4.1-0.20200120150455-7dc0a2d6ddce
	github.com/fatih/color v1.7.0
	github.com/fsnotify/fsnotify v1.4.9
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/go-git/go-git/v5 v5.2.0
	github.com/go-logr/logr v0.3.0 // indirect
	github.com/gobwas/glob v0.2.4-0.20181002190808-e7a84e9525fe
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/mock v1.5.0
	github.com/google/go-cmp v0.5.2 // indirect
	github.com/gophercloud/gophercloud v0.2.0 // indirect
	github.com/gorilla/mux v1.7.3 // indirect
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79
	github.com/hinshun/vt10x v0.0.0-20180809195222-d55458df857c
	github.com/imdario/mergo v0.3.11 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/kevinburke/ssh_config v0.0.0-20201106050909-4977a11b4351 // indirect
	github.com/kubernetes-sigs/service-catalog v0.2.2
	github.com/kylelemons/godebug v1.1.1-0.20190824192725-fa7b53cdfc91
	github.com/mattn/go-colorable v0.1.2
	github.com/mattn/go-isatty v0.0.13-0.20200128103942-cb30d6282491 // indirect
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/mgutz/ansi v0.0.0-20170206155736-9520e82c474b
	github.com/olekukonko/tablewriter v0.0.0-20180506121414-d4647c9c7a84
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.2
	github.com/opencontainers/image-spec v1.0.2-0.20200206005212-79b036d80240 // indirect
	github.com/openshift/api v3.9.1-0.20190924102528-32369d4db2ad+incompatible
	github.com/openshift/client-go v0.0.0-20200116152001-92a2713fa240
	github.com/openshift/library-go v0.0.0-20200407165825-2e79bd232e72
	github.com/openshift/oc v0.0.0-alpha.0.0.20200305142246-2576e482bf00
	github.com/operator-framework/operator-lifecycle-manager v0.0.0-20200422144016-a6acf50218ed
	github.com/pborman/uuid v1.2.0
	github.com/pkg/errors v0.9.1
	github.com/posener/complete v1.1.2
	github.com/redhat-developer/service-binding-operator v0.7.0 // indirect
	github.com/segmentio/backo-go v0.0.0-20200129164019-23eae7c10bd3 // indirect
	github.com/spf13/afero v1.2.2
	github.com/spf13/cobra v0.0.6
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.6.1
	github.com/tidwall/gjson v1.6.3
	github.com/xtgo/uuid v0.0.0-20140804021211-a0b114877d4c // indirect
	github.com/zalando/go-keyring v0.1.0
	golang.org/x/crypto v0.0.0-20201124201722-c8d3bf9c5392 // indirect
	golang.org/x/net v0.0.0-20201110031124-69a78807bb2b // indirect
	golang.org/x/sys v0.0.0-20201119102817-f84b799fce68 // indirect
	golang.org/x/term v0.0.0-20201117132131-f5c789dd3221
	gopkg.in/AlecAivazis/survey.v1 v1.8.0
	gopkg.in/segmentio/analytics-go.v3 v3.1.0
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/api v0.19.2
	k8s.io/apiextensions-apiserver v0.19.2 // indirect
	k8s.io/apimachinery v0.19.2
	k8s.io/cli-runtime v0.17.0
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/klog v1.0.0
	k8s.io/kubectl v0.0.0
	sigs.k8s.io/controller-runtime v0.6.4 // indirect
	sigs.k8s.io/yaml v1.2.0
)

replace (
	bitbucket.org/ww/goautoneg => github.com/munnerz/goautoneg v0.0.0-20120707110453-a547fc61f48d
	github.com/apcera/gssapi => github.com/openshift/gssapi v0.0.0-20161010215902-5fb4217df13b
	github.com/containers/image => github.com/openshift/containers-image v0.0.0-20190130162819-76de87591e9d
	github.com/docker/docker => github.com/docker/docker v17.12.0-ce-rc1.0.20200309214505-aa6a9891b09c+incompatible
	github.com/kr/pty => github.com/creack/pty v1.1.11
	github.com/openshift/api => github.com/openshift/api v0.0.0-20200205133042-34f0ec8dab87
	k8s.io/api => k8s.io/api v0.17.1
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.17.1
	k8s.io/apimachinery => github.com/openshift/kubernetes-apimachinery v0.0.0-20191211181342-5a804e65bdc1
	k8s.io/apiserver => k8s.io/apiserver v0.17.1
	k8s.io/cli-runtime => github.com/openshift/kubernetes-cli-runtime v0.0.0-20200114162348-c8810ef308ee
	k8s.io/client-go => github.com/openshift/kubernetes-client-go v0.0.0-20191211181558-5dcabadb2b45
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.17.1
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.17.1
	k8s.io/code-generator => k8s.io/code-generator v0.17.1
	k8s.io/component-base => k8s.io/component-base v0.17.1
	k8s.io/cri-api => k8s.io/cri-api v0.17.1
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.17.1
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.17.1
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.17.1
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.17.1
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.17.1
	k8s.io/kubectl => github.com/openshift/kubernetes-kubectl v0.0.0-20200211153013-50adac736181
	k8s.io/kubelet => k8s.io/kubelet v0.17.1
	k8s.io/kubernetes => github.com/openshift/kubernetes v1.17.0-alpha.0.0.20191216151305-079984b0a154
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.17.1
	k8s.io/metrics => k8s.io/metrics v0.17.1
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.17.1
	vbom.ml/util => github.com/fvbommel/util v0.0.0-20180919145318-efcd4e0f9787

)

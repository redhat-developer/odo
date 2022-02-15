module github.com/redhat-developer/odo

go 1.16

require (
	github.com/AlecAivazis/survey/v2 v2.0.5
	github.com/Netflix/go-expect v0.0.0-20201125194554-85d881c3777e
	github.com/Xuanwo/go-locale v1.0.0
	github.com/blang/semver v3.5.1+incompatible
	github.com/devfile/api/v2 v2.0.0-20220117162434-6e6e6a8bc14c
	github.com/devfile/library v1.2.1-0.20220217161036-0f5995513e92
	github.com/devfile/registry-support/index/generator v0.0.0-20211012185733-0a73f866043f
	github.com/devfile/registry-support/registry-library v0.0.0-20211125162259-d7edf148d3e2
	github.com/fatih/color v1.10.0
	github.com/frapposelli/wwhrd v0.4.0
	github.com/fsnotify/fsnotify v1.5.1
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/go-git/go-git/v5 v5.3.0
	github.com/go-openapi/spec v0.19.5
	github.com/golang/mock v1.5.0
	github.com/hinshun/vt10x v0.0.0-20220127042424-3ca73d0126d7
	github.com/kr/pty v1.1.5
	github.com/kubernetes-sigs/service-catalog v0.3.1
	github.com/kylelemons/godebug v1.1.0
	github.com/mattn/go-colorable v0.1.8
	github.com/mgutz/ansi v0.0.0-20200706080929-d51e80ef957d // indirect
	github.com/monochromegane/go-gitignore v0.0.0-20200626010858-205db1a8cc00
	github.com/olekukonko/tablewriter v0.0.5
	github.com/onsi/ginkgo v4.7.0-origin.0+incompatible
	github.com/onsi/gomega v1.15.0
	github.com/openshift/api v0.0.0-20210831091943-07e756545ac1
	github.com/openshift/client-go v0.0.0-20210831095141-e19a065e79f7
	github.com/openshift/library-go v0.0.0-20210923120925-caee30353c0d // indirect
	github.com/openshift/oc v0.0.0-alpha.0.0.20210902003738-96e95cef877b
	github.com/operator-framework/api v0.3.20
	github.com/operator-framework/operator-lifecycle-manager v0.17.0
	github.com/pborman/uuid v1.2.0
	github.com/pkg/errors v0.9.1
	github.com/posener/complete v1.1.1
	github.com/redhat-developer/alizer/go v0.0.0-20220215154256-33df7feef4ae
	github.com/redhat-developer/service-binding-operator v1.0.1-0.20211222115357-5b7bbba3bfb3
	github.com/securego/gosec/v2 v2.8.0
	github.com/segmentio/backo-go v0.0.0-20200129164019-23eae7c10bd3 // indirect
	github.com/spf13/afero v1.2.2
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/tidwall/gjson v1.9.3
	github.com/xtgo/uuid v0.0.0-20140804021211-a0b114877d4c // indirect
	github.com/zalando/go-keyring v0.1.1
	golang.org/x/sys v0.0.0-20220204135822-1c1b9b1eba6a // indirect
	golang.org/x/term v0.0.0-20210317153231-de623e64d2a6
	gopkg.in/AlecAivazis/survey.v1 v1.8.8
	gopkg.in/segmentio/analytics-go.v3 v3.1.0
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.22.2
	k8s.io/apimachinery v0.22.2
	k8s.io/cli-runtime v0.22.0-rc.0
	k8s.io/client-go v0.22.2
	k8s.io/klog v1.0.0
	k8s.io/klog/v2 v2.10.0
	k8s.io/kubectl v0.22.1
	k8s.io/utils v0.0.0-20210819203725-bdf08cb9a70a
	sigs.k8s.io/controller-runtime v0.10.2
	sigs.k8s.io/yaml v1.3.0 // indirect

)

replace (
	bitbucket.org/ww/goautoneg => github.com/munnerz/goautoneg v0.0.0-20120707110453-a547fc61f48d
	github.com/Microsoft/hcsshim => github.com/Microsoft/hcsshim v0.8.7
	github.com/apcera/gssapi => github.com/openshift/gssapi v0.0.0-20161010215902-5fb4217df13b
	github.com/containers/image => github.com/openshift/containers-image v0.0.0-20190130162819-76de87591e9d
	github.com/docker/docker => github.com/docker/docker v1.4.2-0.20191121165722-d1d5f6476656
	github.com/onsi/ginkgo => github.com/openshift/onsi-ginkgo v4.7.0-origin.0+incompatible
	k8s.io/api => github.com/openshift/kubernetes/staging/src/k8s.io/api v0.0.0-20210831004331-1199c36daed6
	k8s.io/apimachinery => github.com/openshift/kubernetes/staging/src/k8s.io/apimachinery v0.0.0-20210831004331-1199c36daed6
	k8s.io/cli-runtime => github.com/openshift/kubernetes/staging/src/k8s.io/cli-runtime v0.0.0-20210831004331-1199c36daed6
	k8s.io/client-go => github.com/openshift/kubernetes/staging/src/k8s.io/client-go v0.0.0-20210831004331-1199c36daed6
	k8s.io/component-helpers => k8s.io/component-helpers v0.0.0-20211006165314-dacad8cb3fcb
	k8s.io/kubectl => github.com/openshift/kubernetes/staging/src/k8s.io/kubectl v0.0.0-20210831004331-1199c36daed6
	k8s.io/metrics => k8s.io/metrics v0.0.0-20211006171351-de75bc981086
)

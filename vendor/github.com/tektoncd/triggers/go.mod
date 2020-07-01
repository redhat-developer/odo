module github.com/tektoncd/triggers

go 1.14

require (
	contrib.go.opencensus.io/exporter/ocagent v0.6.0 // indirect
	contrib.go.opencensus.io/exporter/stackdriver v0.12.9 // indirect
	github.com/GoogleCloudPlatform/cloud-builders/gcs-fetcher v0.0.0-20191203181535-308b93ad1f39
	github.com/gobuffalo/envy v1.9.0 // indirect
	github.com/golang/protobuf v1.3.4
	github.com/google/cel-go v0.4.2
	github.com/google/go-cmp v0.4.0
	github.com/google/go-github/v31 v31.0.0
	github.com/gorilla/mux v1.7.3
	github.com/grpc-ecosystem/grpc-gateway v1.13.0 // indirect
	github.com/openzipkin/zipkin-go v0.2.2 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/tektoncd/pipeline v0.11.3
	github.com/tektoncd/plumbing v0.0.0-20200430135134-e53521e1d887
	github.com/tidwall/gjson v1.3.5 // indirect
	github.com/tidwall/sjson v1.0.4
	go.uber.org/zap v1.13.0
	golang.org/x/crypto v0.0.0-20200220183623-bac4c82f6975 // indirect
	golang.org/x/xerrors v0.0.0-20191204190536-9bdfabe68543
	google.golang.org/genproto v0.0.0-20200305110556-506484158171
	gopkg.in/yaml.v2 v2.2.8
	k8s.io/api v0.18.2
	k8s.io/apimachinery v0.18.2
	k8s.io/client-go v0.18.2
	k8s.io/code-generator v0.17.1
	k8s.io/klog v1.0.0
	k8s.io/kube-openapi v0.0.0-20200121204235-bf4fb3bd569c
	k8s.io/utils v0.0.0-20200324210504-a9aa75ae1b89 // indirect
	knative.dev/caching v0.0.0-20200228235451-13d271455c74
	knative.dev/pkg v0.0.0-20200207155214-fef852970f43
	sigs.k8s.io/yaml v1.2.0 // indirect
)

// Knative deps (release-0.12)
replace (
	contrib.go.opencensus.io/exporter/stackdriver => contrib.go.opencensus.io/exporter/stackdriver v0.12.9-0.20191108183826-59d068f8d8ff
	knative.dev/caching => knative.dev/caching v0.0.0-20200116200605-67bca2c83dfa
	knative.dev/pkg => knative.dev/pkg v0.0.0-20200113182502-b8dc5fbc6d2f
	knative.dev/pkg/vendor/github.com/spf13/pflag => github.com/spf13/pflag v1.0.5
)

// Pin k8s deps to 1.16.5
replace (
	k8s.io/api => k8s.io/api v0.16.5
	k8s.io/apimachinery => k8s.io/apimachinery v0.16.5
	k8s.io/client-go => k8s.io/client-go v0.16.5
	k8s.io/code-generator => k8s.io/code-generator v0.16.5
	k8s.io/gengo => k8s.io/gengo v0.0.0-20190327210449-e17681d19d3a
)

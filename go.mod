module kraken.dev/kraken-scheduler

go 1.14

require (
	github.com/cloudevents/sdk-go/v2 v2.2.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/golang/protobuf v1.4.2 // indirect
	github.com/google/go-cmp v0.5.2 // indirect
	github.com/google/uuid v1.1.1 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/influxdata/tdigest v0.0.1 // indirect
	github.com/kelseyhightower/envconfig v1.4.0 // indirect
	github.com/slinkydeveloper/loadastic v0.0.0-20191203132749-9afe5a010a57 // indirect
	github.com/stretchr/testify v1.6.0 // indirect
	go.opencensus.io v0.22.5-0.20200716030834-3456e1d174b2 // indirect
	go.opentelemetry.io/otel v0.4.2 // indirect
	go.uber.org/zap v1.15.0
	golang.org/x/tools v0.0.0-20200916195026-c9a70fc28ce3 // indirect
	honnef.co/go/tools v0.0.1-2020.1.5 // indirect
	k8s.io/api v0.18.8
	k8s.io/apimachinery v0.18.8
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	k8s.io/code-generator v0.18.8
	k8s.io/utils v0.0.0-20200603063816-c1c6865ac451 // indirect
	knative.dev/eventing v0.17.1-0.20200925222044-b313bac67b1c // indirect
	knative.dev/pkg v0.0.0-20200922164940-4bf40ad82aab
	knative.dev/serving v0.18.0
	knative.dev/test-infra v0.0.0-20200921012245-37f1a12adbd3
)

replace (
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.3.1
	github.com/prometheus/client_golang => github.com/prometheus/client_golang v0.9.2
	gomodules.xyz/jsonpatch/v2 => github.com/gomodules/jsonpatch/v2 v2.1.0
	k8s.io/api => k8s.io/api v0.18.8
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.18.8
	k8s.io/apimachinery => k8s.io/apimachinery v0.18.8
	k8s.io/client-go => k8s.io/client-go v0.18.8
	k8s.io/code-generator => k8s.io/code-generator v0.18.8
)

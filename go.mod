module github.com/symcn/sym-ops

go 1.16

require (
	github.com/Shopify/logrus-bugsnag v0.0.0-20171204204709-577dee27f20d // indirect
	github.com/bshuster-repo/logrus-logstash-hook v1.0.0 // indirect
	github.com/bugsnag/bugsnag-go v2.1.0+incompatible // indirect
	github.com/bugsnag/panicwrap v1.3.1 // indirect
	github.com/docker/go-metrics v0.0.1 // indirect
	github.com/docker/libtrust v0.0.0-20160708172513-aabc10ec26b7 // indirect
	github.com/garyburd/redigo v1.6.2 // indirect
	github.com/gofrs/uuid v4.0.0+incompatible // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/uuid v1.2.0 // indirect
	github.com/googleapis/gnostic v0.5.4 // indirect
	github.com/gorilla/handlers v1.5.1 // indirect
	github.com/json-iterator/go v1.1.10
	github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0 // indirect
	github.com/prometheus/common v0.17.0 // indirect
	github.com/prometheus/procfs v0.6.0 // indirect
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/symcn/api v0.0.0-20210225092819-1c733cf5f3f4
	github.com/symcn/pkg v0.0.0-20210225093656-1508bc648e04
	github.com/yvasiyarov/go-metrics v0.0.0-20150112132944-c25f46c4b940 // indirect
	github.com/yvasiyarov/gorelic v0.0.7 // indirect
	github.com/yvasiyarov/newrelic_platform_go v0.0.0-20160601141957-9c099fbc30e9 // indirect
	golang.org/x/crypto v0.0.0-20210220033148-5ea612d1eb83 // indirect
	golang.org/x/net v0.0.0-20210224082022-3d97a244fca7 // indirect
	golang.org/x/oauth2 v0.0.0-20210220000619-9bb904979d93 // indirect
	golang.org/x/sys v0.0.0-20210225091947-4ada9433c6ea // indirect
	golang.org/x/term v0.0.0-20210220032956-6a3ed077a48d // indirect
	golang.org/x/text v0.3.5 // indirect
	golang.org/x/time v0.0.0-20210220033141-f8bda1e9f3ba // indirect
	google.golang.org/appengine v1.6.7 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
	helm.sh/helm/v3 v3.5.2
	k8s.io/api v0.20.4
	k8s.io/apiextensions-apiserver v0.20.4
	k8s.io/apimachinery v0.20.4
	k8s.io/client-go v1.5.1
	k8s.io/klog/v2 v2.5.0
	k8s.io/kube-openapi v0.0.0-20210216185858-15cd8face8d6 // indirect
	rsc.io/letsencrypt v0.0.3 // indirect
	sigs.k8s.io/controller-runtime v0.8.2
	sigs.k8s.io/structured-merge-diff/v4 v4.0.3 // indirect
	sigs.k8s.io/yaml v1.2.0
)

replace (
	github.com/docker/distribution => github.com/docker/distribution v2.7.1+incompatible
	github.com/docker/docker => github.com/moby/moby v17.12.0-ce-rc1.0.20200618181300-9dc6525e6118+incompatible
	k8s.io/client-go => k8s.io/client-go v0.20.4
)

module github.com/crain-cn/event-mesh

go 1.15

require (
	github.com/cenkalti/backoff/v4 v4.1.1
	github.com/cespare/xxhash v1.1.0
	github.com/crain-cn/cluster-mesh v0.0.1
	github.com/fastly/go-utils v0.0.0-20180712184237-d95a45783239 // indirect
	github.com/fsnotify/fsnotify v1.4.9
	github.com/go-kit/kit v0.9.0
	github.com/go-logr/logr v0.1.0 // indirect
	github.com/jehiah/go-strftime v0.0.0-20171201141054-1d33003b3869 // indirect
	github.com/kylelemons/godebug v1.1.0
	github.com/lestrrat/go-envload v0.0.0-20180220120943-6ed08b54a570 // indirect
	github.com/lestrrat/go-file-rotatelogs v0.0.0-20180223000712-d3151e2a480f
	github.com/lestrrat/go-strftime v0.0.0-20180220042222-ba3bf9c1d042 // indirect
	github.com/linvon/cuckoo-filter v0.4.0
	github.com/onsi/ginkgo v1.14.0 // indirect
	github.com/onsi/gomega v1.10.1 // indirect
	github.com/pkg/errors v0.9.1
	github.com/prometheus-operator/prometheus-operator/pkg/client v0.51.2
	github.com/prometheus/alertmanager v0.23.0
	github.com/prometheus/client_golang v1.11.0
	github.com/prometheus/common v0.30.0
	github.com/rifflock/lfshook v0.0.0-20180920164130-b9218ef580f5
	github.com/sirupsen/logrus v1.6.0
	github.com/stretchr/testify v1.7.0
	github.com/tebeka/strftime v0.1.5 // indirect
	gopkg.in/yaml.v2 v2.4.0
	gorm.io/driver/mysql v1.1.2
	gorm.io/gorm v1.21.16
	k8s.io/api v0.18.3
	k8s.io/apimachinery v0.22.0
	k8s.io/cli-runtime v0.22.2 // indirect
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/code-generator v0.22.2 // indirect
	sigs.k8s.io/controller-runtime v0.4.0 // indirect
	sigs.k8s.io/yaml v1.2.0
)

replace (
	github.com/crain-cn/event-mesh => /Users/edz/go/src/github.com/crain-cn/event-mesh
	k8s.io/api => k8s.io/api v0.0.0-20191016110408-35e52d86657a
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20191016113550-5357c4baaf65
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20191004115801-a2eda9f80ab8
	k8s.io/apiserver => k8s.io/apiserver v0.0.0-20191016112112-5190913f932d
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20191016114015-74ad18325ed5
	k8s.io/client-go => k8s.io/client-go v0.0.0-20191016111102-bec269661e48
	k8s.io/code-generator => k8s.io/code-generator v0.0.0-20191004115455-8e001e5d1894
	k8s.io/component-base => k8s.io/component-base v0.0.0-20191016111319-039242c015a9
)

module github.com/crain-cn/event-mesh

go 1.15

require (
	github.com/Shopify/sarama v1.27.2 // indirect
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751 // indirect
	github.com/armon/go-metrics v0.0.0-20190430140413-ec5e00d3c878 // indirect
	github.com/cenkalti/backoff/v4 v4.0.2
	github.com/cespare/xxhash v1.1.0
	github.com/crain-cn/cluster-mesh v0.0.2
	github.com/crain-cn/event-mesh/pkg/k8s/apis v0.0.11
	github.com/crain-cn/event-mesh/pkg/k8s/client v0.0.0-20211018075026-7b332612d535
	github.com/denverdino/aliyungo v0.0.0-20210318042315-546d0768f5c7 // indirect
	github.com/elastic/go-elasticsearch/v7 v7.10.0 // indirect
	github.com/fsnotify/fsnotify v1.4.10-0.20200417215612-7f4cf4dd2b52
	github.com/gin-contrib/pprof v1.3.0 // indirect
	github.com/gin-contrib/sessions v0.0.3 // indirect
	github.com/gin-gonic/gin v1.6.3 // indirect
	github.com/go-kit/kit v0.10.0
	github.com/go-openapi/spec v0.20.3 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/hashicorp/go-immutable-radix v1.1.0 // indirect
	github.com/hashicorp/go-msgpack v0.5.5 // indirect
	github.com/jpillora/backoff v0.0.0-20180909062703-3050d21c67d7 // indirect
	github.com/kylelemons/godebug v1.1.0
	github.com/lestrrat/go-file-rotatelogs v0.0.0-20180223000712-d3151e2a480f
	github.com/lestrrat/go-strftime v0.0.0-20180220042222-ba3bf9c1d042 // indirect
	github.com/linvon/cuckoo-filter v0.3.0
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mitchellh/mapstructure v1.3.2 // indirect
	github.com/mwitkow/go-conntrack v0.0.0-20190716064945-2f068394615f // indirect
	github.com/pkg/errors v0.9.1
	github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring v0.45.0 // indirect
	github.com/prometheus-operator/prometheus-operator/pkg/client v0.45.0
	github.com/prometheus/alertmanager v0.21.0
	github.com/prometheus/client_golang v1.7.1
	github.com/prometheus/common v0.10.0
	github.com/rifflock/lfshook v0.0.0-20180920164130-b9218ef580f5
	github.com/rs/zerolog v1.20.0 // indirect
	github.com/sasha-s/go-deadlock v0.2.1-0.20190427202633-1595213edefa // indirect
	github.com/sirupsen/logrus v1.7.0
	github.com/spaolacci/murmur3 v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.7.0 // indirect
	github.com/stretchr/testify v1.7.0
	github.com/swaggo/gin-swagger v1.3.0 // indirect
	github.com/swaggo/swag v1.7.0 // indirect
	golang.org/x/net v0.0.0-20210224082022-3d97a244fca7 // indirect
	golang.org/x/sync v0.0.0-20201020160332-67f06af15bc9 // indirect
	golang.org/x/sys v0.0.0-20210225091947-4ada9433c6ea // indirect
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e // indirect
	golang.org/x/tools v0.1.0 // indirect
	google.golang.org/grpc v1.34.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v2 v2.4.0
	gorm.io/driver/mysql v1.0.5
	gorm.io/gorm v1.21.3
	k8s.io/api v0.22.2
	k8s.io/apimachinery v0.22.2
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/klog/v2 v2.4.0 // indirect
	sigs.k8s.io/yaml v1.2.0
)

replace (
	github.com/crain-cn/event-mesh/pkg/k8s/apis => ./pkg/k8s/apis
	github.com/crain-cn/event-mesh/pkg/k8s/client => ./pkg/k8s/client
	k8s.io/api => k8s.io/api v0.20.0
	k8s.io/apimachinery => k8s.io/apimachinery v0.20.0
	k8s.io/client-go => k8s.io/client-go v0.20.0
	k8s.io/code-generator => k8s.io/code-generator v0.20.0
)

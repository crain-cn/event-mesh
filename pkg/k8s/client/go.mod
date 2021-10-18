module github.com/crain-cn/event-mesh/pkg/k8s/client

go 1.15

replace (
	github.com/crain-cn/event-mesh/pkg/k8s/apis => ./../apis
	k8s.io/api => k8s.io/api v0.20.0
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.20.0
	k8s.io/apimachinery => k8s.io/apimachinery v0.20.0
	k8s.io/client-go => k8s.io/client-go v0.20.0
)

require (
	github.com/crain-cn/event-mesh/pkg/k8s/apis v0.0.11
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v0.20.2
)

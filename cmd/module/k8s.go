package module

import (
	"github.com/crain-cn/event-mesh/cmd/config"
	"github.com/crain-cn/event-mesh/pkg/k8s/watcher"
	"github.com/crain-cn/event-mesh/pkg/provider"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func SetupK8s(o options, configResolver *config.ConfigResolver, alerts provider.Alerts) *rest.Config {
	// creates the connection
	log.Info("SetupK8sClient...")
	clientConfig, err := clientcmd.BuildConfigFromFlags(o.master, o.kubeConfig)
	if err != nil {
		log.Fatal(err)
	}
	k8sWatcher := watcher.NewK8sWatcher(configResolver, clientConfig)
	//k8sClient, err := kubernetes.NewForConfig(clientConfig)
	go k8sWatcher.EnableK8sWatcher(alerts)

	return clientConfig
}

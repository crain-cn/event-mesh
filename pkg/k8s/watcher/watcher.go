package watcher

import (
	"github.com/crain-cn/event-mesh/api/model"
	"github.com/crain-cn/event-mesh/cmd/config"
	eventmesh_v1 "github.com/crain-cn/event-mesh/pkg/k8s/apis/eventmesh/v1"
	"github.com/crain-cn/event-mesh/pkg/k8s/clustermesh"
	"github.com/crain-cn/event-mesh/pkg/k8s/events"
	"github.com/crain-cn/event-mesh/pkg/logging"
	"github.com/crain-cn/event-mesh/pkg/logging/logfields"
	"github.com/crain-cn/event-mesh/pkg/provider"
	"github.com/crain-cn/cluster-mesh/api/cloud.mesh/v1beta1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"sync"
)

var (
	// log is the k8s package logger object.
	log = logging.DefaultLogger.WithField(logfields.LogSubsys, "k8s_watcher")
)

var (
//k8sCM = controller.NewManager()
)

type ClusterManagerInterface interface {
	AddCluster(cluster *v1beta1.Cluster) (bool, error)
	UpdateCluster(cluster *v1beta1.Cluster) (bool, error)
	DeleteCluster(cluster *v1beta1.Cluster) error
}

type EventRouteManagerInterface interface {
	AddEventRoute(eventroute *eventmesh_v1.EventRoute) (bool, error)
	UpdateEventRoute(eventroute *eventmesh_v1.EventRoute) (bool, error)
	DeleteEventRoute(eventroute *eventmesh_v1.EventRoute) error
}

type K8sWatcher struct {
	// k8sResourceSynced maps a resource name to a channel. Once the given
	// resource name is synchronized with k8s, the channel for which that
	// resource name maps to is closed.
	// k8sAPIGroups is a set of k8s API in use. They are setup in EnableK8sWatcher,
	// and may be disabled while the agent runs.
	// This is on this object, instead of a global, because EnableK8sWatcher is
	// on Daemon.
	clientConfig      *rest.Config
	configResolver    *config.ConfigResolver
	clusterManager    *clustermesh.ClusterManager
	eventRouteManager *events.EventRouteManager
	// controllersStarted is a channel that is closed when all controllers, i.e.,
	// k8s watchers have started listening for k8s events.
	controllersStarted chan struct{}
}

func NewK8sWatcher(configResolver *config.ConfigResolver, clientConfig *rest.Config) *K8sWatcher {
	return &K8sWatcher{
		configResolver:     configResolver,
		clientConfig:       clientConfig,
		controllersStarted: make(chan struct{}),
	}
}

// k8sMetrics implements the LatencyMetric and ResultMetric interface from
// k8s client-go package
type k8sMetrics struct{}

// EnableK8sWatcher watches for eventRouter, clusterManager cluster changes on the Kubernetes
// api server defined in the receiver's daemon k8sClient.
func (k *K8sWatcher) EnableK8sWatcher(alerts provider.Alerts) error {

	log.Info("Enabling k8s event listener")

	asyncControllers := &sync.WaitGroup{}
	//swg := lock.NewStoppableWaitGroup()
	k.clusterManager = clustermesh.NewClusterManager(k.configResolver, k.clientConfig, alerts)
	k.clusterManager.Reasons = model.GetEventReasonsAll()
	k.clusterManager.Workcodes = model.GetAppUserAll()
	k.clusterManager.ClusterMeshInit(asyncControllers)
	asyncControllers.Add(1)

	k.eventRouteManager = events.NewEventRouteManager(k.clientConfig)
	k.eventRouteManager.ReceiverInit(asyncControllers)
	k.eventRouteManager.EventRouteInit(asyncControllers)
	asyncControllers.Add(1)

	asyncControllers.Wait()
	return nil
}

// GetStore returns the k8s cache store for the given resource name.
func (k *K8sWatcher) GetStore(name string) cache.Store {
	switch name {
	case "eventroute":
		return k.eventRouteManager.Store
	case "cluster":
		return k.clusterManager.Store
	default:
		return nil
	}
}

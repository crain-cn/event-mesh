package clustermesh

import (
	"github.com/crain-cn/event-mesh/cmd/config"
	"github.com/crain-cn/event-mesh/pkg/k8s/events"
	"github.com/crain-cn/event-mesh/pkg/provider"
	"github.com/crain-cn/cluster-mesh/api/cloud.mesh/v1beta1"
	"github.com/crain-cn/cluster-mesh/client/clientset/versioned"
	"github.com/crain-cn/cluster-mesh/client/informers/externalversions"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/record"
	"os"
	"reflect"
	"sync"
	"time"
)

type ClusterManager struct {
	client         *versioned.Clientset
	configResolver *config.ConfigResolver
	clusterClient  *ClusterClient
	eventWatchers  map[string]*events.EventWatcher
	eventRecorder  record.EventRecorder
	Lock           sync.Mutex
	Alerts         provider.Alerts
	Store          cache.Store
	Reasons 	   map[string]string
	Workcodes	   map[string]string
	CacheSynced    chan struct{}
	stopCh         chan struct{}
}

var (
	scheme = runtime.NewScheme()
)

func init() {
	err := v1beta1.AddToScheme(scheme)
	if err != nil {
		panic(err.Error())
	}
}

func NewClusterManager(configResolver *config.ConfigResolver, clientConfig *rest.Config, alerts provider.Alerts) *ClusterManager {
	client, err := versioned.NewForConfig(clientConfig)
	if err != nil {
		log.Error(err)
		//return nil, fmt.Errorf("unable to create k8s client: %s", err)
	}

	clusterManager := &ClusterManager{
		client:         client,
		configResolver: configResolver,
		CacheSynced:    make(chan struct{}),
		clusterClient:  NewClusterClinet(client),
		eventWatchers:  make(map[string]*events.EventWatcher),
		Alerts:         alerts,
		stopCh:         make(chan struct{}),
	}

	//list, err := client.CloudV1beta1().Clusters().List(context.TODO(), metav1.ListOptions{})
	//if err != nil {
	//	log.Error(err)
	//}
	//for _, cluster := range list.Items {
	//clusterManager.AddCluster(&cluster)
	//}

	return clusterManager
}

func (m *ClusterManager) ClusterMeshInit(asyncControllers *sync.WaitGroup) {

	sharedInformerFactory := externalversions.NewSharedInformerFactory(m.client, time.Minute*1)
	clusterInformer := sharedInformerFactory.Cloud().V1beta1().Clusters()
	clusterInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if cluster := ObjToV1beta1Cluster(obj); cluster != nil {
				clusterCpy := cluster.DeepCopy()
				m.AddCluster(clusterCpy)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			if old := ObjToV1beta1Cluster(oldObj); old != nil {
				if new := ObjToV1beta1Cluster(newObj); new != nil {
					if reflect.DeepEqual(old, new) {
						return
					}
					oldCpy := old.DeepCopy()
					newCpy := new.DeepCopy()
					m.UpdateCluster(oldCpy, newCpy)
				}
			}
		},
		DeleteFunc: func(obj interface{}) {
			cluster := ObjToV1beta1Cluster(obj)
			if cluster == nil {
				return
			}
			m.DeleteCluster(cluster)
		},
	})
	go clusterInformer.Informer().Run(m.stopCh)
	sharedInformerFactory.Start(m.stopCh)
	sharedInformerFactory.WaitForCacheSync(m.stopCh)
}

func (m *ClusterManager) GetDevK8S() *kubernetes.Clientset{
	clientConfig, err := clientcmd.BuildConfigFromFlags("", "/Users/edz/.kube/k8s-32-dev")
	if err != nil {
		log.Fatal(err)
	}

	kcClient, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		log.Error(err)
	}
	return kcClient
}

func (m *ClusterManager) AddCluster(cluster *v1beta1.Cluster) (bool, error) {
	envStr := os.Getenv("env")
	if envStr == "dev" && cluster.Name != "k8s-dev" {
		return false, nil
	}
	log.Info("add cluster:", cluster.Name)
	kcClient, err := m.clusterClient.GetClient(cluster.Name)
	if err != nil {
		log.Error(err)
	}
	if cluster.Name != "aliyun-us-rock-online" && cluster.Name != "neirongyun-aliyun-ack" && cluster.Name != "aliyun-yunxuexi-eci-online" {
		//m.eventWatchers[cluster.Name] = events.NewEventWatcher(m.configResolver, cluster.Name, kcClient, m.Alerts)
		m.eventWatchers[cluster.Name] = events.NewEventControllerWatcher(m.Workcodes,m.Reasons,m.configResolver, cluster.Name, kcClient, m.Alerts)
	}
	return false, nil
}

func (s *ClusterManager) UpdateCluster(old, new *v1beta1.Cluster) (bool, error) {
	log.WithFields(logrus.Fields{
		"cluster_name": old.ClusterName,
	}).Debug(" Update Cluster ")
	return false, nil
}

func (m *ClusterManager) DeleteCluster(cluster *v1beta1.Cluster) error {
	m.Lock.Lock()
	if watcher, ok := m.eventWatchers[cluster.ClusterName]; !ok {
		watcher.Stop()
	}
	delete(m.eventWatchers, cluster.ClusterName)
	delete(m.clusterClient.kcClients, cluster.ClusterName)
	m.Lock.Unlock()
	log.WithFields(logrus.Fields{
		"cluster_name": cluster.ClusterName,
	}).Debug(" Delete Cluster")
	return nil
}

func ObjToV1beta1Cluster(obj interface{}) *v1beta1.Cluster {
	cluster, ok := obj.(*v1beta1.Cluster)
	if ok {
		return cluster
	}
	deletedObj, ok := obj.(cache.DeletedFinalStateUnknown)
	if ok {
		// Delete was not observed by the watcher but is
		// removed from kube-apiserver. This is the last
		// known state and the object no longer exists.
		cluster, ok := deletedObj.Obj.(*v1beta1.Cluster)
		if ok {
			return cluster
		}
	}
	return nil
}

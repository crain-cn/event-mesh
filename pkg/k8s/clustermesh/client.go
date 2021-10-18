package clustermesh

import (
	"context"
	"errors"
	"github.com/crain-cn/event-mesh/pkg/k8s/kubeutil"
	"github.com/crain-cn/cluster-mesh/client/clientset/versioned"
	"github.com/crain-cn/cluster-mesh/client/informers/externalversions"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sync"
)

type ClusterClient struct {
	Client          *versioned.Clientset
	kcClients       map[string]*kubernetes.Clientset
	SharedInformers externalversions.SharedInformerFactory
	Lock            sync.Mutex
}

func (c *ClusterClient) GetClient(clusterName string) (*kubernetes.Clientset, error) {
	if _, ok := c.kcClients[clusterName]; !ok {
		kubeClient, err := c.CreateClient(clusterName)
		if kubeClient == nil || err != nil {
			log.Info("Failed to client cluster:", clusterName)
			return nil, errors.New("invalid kubeconfig")
		} else {
			c.Lock.Lock()
			c.kcClients[clusterName] = kubeClient
			c.Lock.Unlock()
			return kubeClient, nil
		}
	}
	return c.kcClients[clusterName], nil
}

func (c *ClusterClient) CreateClient(clusterName string) (*kubernetes.Clientset, error) {
	cluster, err := c.Client.CloudV1beta1().
		Clusters().
		Get(context.TODO(), clusterName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	if err != nil {
		log.Error(err, "Failed to loadConfig  Cluster:%v", clusterName)
		return nil, err
	}
	config, err := kubeutil.LoadConfig(cluster.Spec.KubeConfig)
	return kubernetes.NewForConfig(config)
}

func NewClusterClinet(client *versioned.Clientset) *ClusterClient {
	return &ClusterClient{
		Client:    client,
		kcClients: make(map[string]*kubernetes.Clientset),
	}
}

package clustermesh

import (
	"github.com/crain-cn/cluster-mesh/api/cloud.mesh/v1beta1"
	"k8s.io/client-go/tools/cache"
)

type ClusterAction string

const (
	ClusterAdd    = ClusterAction("add")
	ClusterUpdate = ClusterAction("update")
	ClusterDelete = ClusterAction("delete")
)

func ObjToV1Cluster(obj interface{}) *v1beta1.Cluster {
	cluster, ok := obj.(*v1beta1.Cluster)
	if ok {
		return cluster
	}
	deletedObj, ok := obj.(cache.DeletedFinalStateUnknown)
	if ok {
		// Delete was not observed by the watcher but is
		// removed from kube-apiserver. This is the last
		// known state and the object no longer exists.
		svc, ok := deletedObj.Obj.(*v1beta1.Cluster)
		if ok {
			return svc
		}
	}
	return nil
}

type Event struct {
	Action         ClusterAction
	Cluster        *v1beta1.Cluster
	ClusterManager *ClusterManager
}

type ClusterResult struct {
	err error
}

/*
func (e *Event) Handle(res chan interface{}) {
	switch e.Action {
	case ClusterAdd:
		e.ClusterManager.AddCluster(e.Cluster)
	case ClusterUpdate:
		e.ClusterManager.UpdateCluster(e.Cluster)
	case ClusterDelete:
		e.ClusterManager.DeleteCluster(e.Cluster)
	}

	res <- &ClusterResult{
		err: nil,
	}
}
*/

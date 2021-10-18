package events

import (
	"fmt"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
)

type EventsController struct {
	indexer  cache.Indexer  // Indexer 的引用
	informer cache.Controller // Informer 的引用
}

func NewEventsController(indexer  cache.Indexer,informer cache.Controller) *EventsController {
	return &EventsController{
		indexer:  indexer,
		informer: informer,
	}
}

func (c *EventsController) Run(stopCh chan struct{}) {
	defer runtime.HandleCrash()
	go c.informer.Run(stopCh)   // 启动 informer

	if !cache.WaitForCacheSync(stopCh, c.informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("Time out waitng for caches to sync"))
		return
	}
	<-stopCh
}

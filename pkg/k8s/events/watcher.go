package events

import (
	"github.com/crain-cn/event-mesh/cmd/config"
	"github.com/crain-cn/event-mesh/pkg/provider"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"os"
	"time"
)

const (
	resync = time.Minute * 1
)

type EventWatcher struct {
	Kc                  kubernetes.Interface
	ConfigResolver      *config.ConfigResolver
	StopCh              chan struct{}
	SharedIndexInformer cache.SharedIndexInformer
	InformerSynced      cache.InformerSynced
	Cache               map[string]corev1.Event
	StartTime           time.Time
}

func NewEventControllerWatcher(workcodes,reasons map[string]string,configResolver *config.ConfigResolver, cluster string, kc kubernetes.Interface, alerts provider.Alerts) *EventWatcher {
	// 指定 ListWatcher 在所有namespace下监听 Events 资源
	log.Info("NewEventWatcher:", cluster)

	startTime := time.Now()
	eventWatcher := &EventWatcher{
		Kc:                  kc,
		ConfigResolver:      configResolver,
		StopCh:              make(chan struct{}),
	}

	envStr := os.Getenv("env")
	eventListWatcher := cache.NewListWatchFromClient(kc.CoreV1().RESTClient(), "events", corev1.NamespaceAll, fields.Everything())
	filters := configResolver.GetEventSinksFilters(cluster)
	notFilters := configResolver.GetEventSinksNotFilters(cluster)

	eventFilter := NewEventFilter(startTime, ToFilterList(filters), ToFilterList(notFilters))
	// 创建 indexer 和 informer
	indexer, informer := cache.NewIndexerInformer(eventListWatcher, &corev1.Event{}, 0, cache.ResourceEventHandlerFuncs{
		// 当有 pod 创建时，根据 Delta queue 弹出的 object 生成对应的Key，并加入到 workqueue中。此处可以根据Object的一些属性，进行过滤
		AddFunc: func(obj interface{}) {
			if event := ObjToV1Event(obj); event != nil {
				filterResult := eventFilter.Filter(event)
				if filterResult {
					return
				}
				el := &ElEvent{
					Cluster:   cluster,
					Event:     event,
					StartTime: startTime,
					Alerts:    alerts,
				}
				el.ENV = envStr
				el.Reasons = reasons
				el.Workcodes = workcodes
				go el.Handle2()
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			if old := ObjToV1Event(oldObj); old != nil {
				if new := ObjToV1Event(newObj); new != nil {
					filterResult := eventFilter.Filter(new)
					if filterResult {
						return
					}
					log.Info("ElEvent Update: ",&ElEvent{
						Cluster:   cluster,
						Event:     new,
						StartTime: startTime,
						SlsSink:   nil,
						Alerts:    alerts,
					})
					el := &ElEvent{
						Cluster:   cluster,
						Event:     new,
						StartTime: startTime,
						Alerts:    alerts,
					}
					el.Workcodes = workcodes
					el.Reasons = reasons
					el.ENV = envStr
					go el.Handle2()
				}
			}
		},
		DeleteFunc: func(obj interface{}) {
		},
	}, cache.Indexers{})
	controller := NewEventsController(indexer, informer)
	// 启动 controller
	go controller.Run(eventWatcher.StopCh)
	return eventWatcher
}


//func NewEventWatcher2(configResolver *config.ConfigResolver, cluster string, kc kubernetes.Interface, alerts provider.Alerts) *EventWatcher {
//	log.Info("NewEventWatcher:", cluster)
//	eq := eventqueue.NewEventQueueBuffered(cluster, 10000)
//	eq.Run()
//
//	watchlist := cache.NewListWatchFromClient(kc.CoreV1().RESTClient(),"events",v1.NamespaceAll, fields.Everything())
//	sharedIndexInformer := cache.NewSharedIndexInformer(watchlist, &corev1.Event{}, 0, cache.Indexers{})
//	startTime := time.Now()
//	eventWatcher := &EventWatcher{
//		Kc:                  kc,
//		ConfigResolver:      configResolver,
//		StopCh:              make(chan struct{}),
//		SharedIndexInformer: sharedIndexInformer,
//		InformerSynced:      sharedIndexInformer.HasSynced,
//		EventQueue:          eq,
//	}
//
//	filters := configResolver.GetEventSinksFilters(cluster)
//	notFilters := configResolver.GetEventSinksNotFilters(cluster)
//	slsSink := eventWatcher.SetupSLSSink(cluster)
//	eventFilter := NewEventFilter(startTime, ToFilterList(filters), ToFilterList(notFilters))
//
//	_, controller := cache.NewInformer(watchlist,
//		&corev1.Event{},
//		0, //Duration is int64
//		cache.ResourceEventHandlerFuncs{
//			AddFunc: func(obj interface{}) {
//				if event := ObjToV1Event(obj); event != nil {
//					filterResult := eventFilter.Filter(event)
//					if filterResult {
//						return
//					}
//					eventWatcher.EventQueue.Enqueue(eventqueue.NewEvent(&ElEvent{
//						Cluster:   cluster,
//						Event:     event,
//						StartTime: startTime,
//						SlsSink:   slsSink,
//						Alerts:    alerts,
//					}))
//				}
//			},
//			DeleteFunc: func(obj interface{}) {
//			},
//			UpdateFunc: func(oldObj, newObj interface{}) {
//				if old := ObjToV1Event(oldObj); old != nil {
//					if new := ObjToV1Event(newObj); new != nil {
//						filterResult := eventFilter.Filter(new)
//						if filterResult {
//							return
//						}
//						eventWatcher.EventQueue.Enqueue(eventqueue.NewEvent(&ElEvent{
//							Cluster:   cluster,
//							Event:     new,
//							StartTime: startTime,
//							Alerts:    alerts,
//						}))
//					}
//				}
//			},
//		},
//	)
//
//
//	go controller.Run(eventWatcher.StopCh)
//	go cache.WaitForCacheSync(eventWatcher.StopCh, eventWatcher.InformerSynced)
//	return eventWatcher
//}

func (e *EventWatcher) Stop() {
	// send a finish signal
	e.StopCh <- struct{}{}
}

func ObjToV1Event(obj interface{}) *corev1.Event {
	event, ok := obj.(*corev1.Event)
	if ok {
		return event
	}
	deletedObj, ok := obj.(cache.DeletedFinalStateUnknown)
	if ok {
		// Delete was not observed by the watcher but is
		// removed from kube-apiserver. This is the last
		// known state and the object no longer exists.
		event, ok := deletedObj.Obj.(*corev1.Event)
		if ok {
			return event
		}
	}
	return nil
}

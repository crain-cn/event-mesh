package events

import (
	eventmeshv1 "github.com/crain-cn/event-mesh/pkg/k8s/apis/eventmesh/v1"
	notification_v1 "github.com/crain-cn/event-mesh/pkg/k8s/apis/notification/v1"
	e_versioned "github.com/crain-cn/event-mesh/pkg/k8s/client/clientset/versioned"
	e_externalversions "github.com/crain-cn/event-mesh/pkg/k8s/client/informers/externalversions"
	"io/ioutil"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"reflect"
	"sync"
	"time"
)

type EventRouteManager struct {
	client       *e_versioned.Clientset
	cfgGenerator *configGenerator
	Store        cache.Store
	CacheSynced  chan struct{}

	listerSyncedReceiver   bool
	listerSyncedEventRoute bool
	stopChReceiver         chan struct{}
	stopChEventRoute       chan struct{}
}

func NewEventRouteManager(clientConfig *rest.Config) *EventRouteManager {
	client, err := e_versioned.NewForConfig(clientConfig)
	if err != nil {
		//return nil, fmt.Errorf("unable to create k8s client: %s", err)
	}
	return &EventRouteManager{
		client:                 client,
		listerSyncedReceiver:   false,
		listerSyncedEventRoute: false,
		CacheSynced:            make(chan struct{}),
		stopChReceiver:         make(chan struct{}),
		stopChEventRoute:       make(chan struct{}),
		cfgGenerator:           NewConfigGenerator(),
	}
}

func ObjToV1EventRoute(obj interface{}) *eventmeshv1.EventRoute {
	k8sNP, ok := obj.(*eventmeshv1.EventRoute)
	if ok {
		return k8sNP
	}
	deletedObj, ok := obj.(cache.DeletedFinalStateUnknown)
	if ok {
		// Delete was not observed by the watcher but is
		// removed from kube-apiserver. This is the last
		// known state and the object no longer exists.
		k8sNP, ok := deletedObj.Obj.(*eventmeshv1.EventRoute)
		if ok {
			return k8sNP
		}
	}
	return nil
}

func (s *EventRouteManager) EventRouteInit(asyncControllers *sync.WaitGroup) {
	log.Info("eventRoute informer start")
	sharedInformerFactory := e_externalversions.NewSharedInformerFactory(s.client, time.Minute*1)
	eventRouteInformer := sharedInformerFactory.Eventmesh().V1().EventRoutes()
	eventRouteInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if route := ObjToV1EventRoute(obj); route != nil {
				routeCpy := route.DeepCopy()
				s.AddEventRoute(routeCpy)
				if s.listerSyncedEventRoute && s.listerSyncedReceiver {
					s.GeneratorConfig()
				}
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			if old := ObjToV1EventRoute(oldObj); old != nil {
				if new := ObjToV1EventRoute(newObj); new != nil {
					oldCpy := old.DeepCopy()
					newCpy := new.DeepCopy()
					s.UpdateEventRoute(oldCpy, newCpy)
					if s.listerSyncedEventRoute && s.listerSyncedReceiver {
						s.GeneratorConfig()
					}
				}
			}
		},
		DeleteFunc: func(obj interface{}) {
			route := ObjToV1EventRoute(obj)
			if route == nil {
				return
			}
			s.DeleteEventRoute(route)
			if s.listerSyncedEventRoute && s.listerSyncedReceiver {
				s.GeneratorConfig()
			}
		},
	},
	)

	go eventRouteInformer.Informer().Run(s.stopChEventRoute)
	sharedInformerFactory.Start(s.stopChEventRoute)
	sharedInformerFactory.WaitForCacheSync(s.stopChEventRoute)
	s.listerSyncedEventRoute = eventRouteInformer.Informer().HasSynced()
	s.GeneratorConfig()
}

func ObjToV1Receiver(obj interface{}) *notification_v1.Receiver {
	k8sNP, ok := obj.(*notification_v1.Receiver)
	if ok {
		return k8sNP
	}
	deletedObj, ok := obj.(cache.DeletedFinalStateUnknown)
	if ok {
		// Delete was not observed by the watcher but is
		// removed from kube-apiserver. This is the last
		// known state and the object no longer exists.
		k8sNP, ok := deletedObj.Obj.(*notification_v1.Receiver)
		if ok {
			return k8sNP
		}
	}
	return nil
}

func (s *EventRouteManager) ReceiverInit(asyncControllers *sync.WaitGroup) {
	log.Info("Receiver informer start")
	sharedInformerFactory := e_externalversions.NewSharedInformerFactory(s.client, time.Minute*1)
	receiverInformer := sharedInformerFactory.Notification().V1().Receivers()
	receiverInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if r := ObjToV1Receiver(obj); r != nil {
				rCpy := r.DeepCopy()
				s.AddReceiver(rCpy)
				if s.listerSyncedEventRoute && s.listerSyncedReceiver {
					s.GeneratorConfig()
				}
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			if old := ObjToV1Receiver(oldObj); old != nil {
				if new := ObjToV1Receiver(newObj); new != nil {
					oldCpy := old.DeepCopy()
					newCpy := new.DeepCopy()
					s.UpdateReceiver(oldCpy, newCpy)
					if s.listerSyncedEventRoute && s.listerSyncedReceiver {
						s.GeneratorConfig()
					}
				}
			}
		},
		DeleteFunc: func(obj interface{}) {
			receiver := ObjToV1Receiver(obj)
			if receiver == nil {
				return
			}
			s.DeleteReceiver(receiver)
			if s.listerSyncedEventRoute && s.listerSyncedReceiver {
				s.GeneratorConfig()
			}
		},
	},
	)
	go receiverInformer.Informer().Run(s.stopChReceiver)
	sharedInformerFactory.Start(s.stopChReceiver)
	sharedInformerFactory.WaitForCacheSync(s.stopChReceiver)
	s.listerSyncedReceiver = receiverInformer.Informer().HasSynced()
	s.GeneratorConfig()
}

func (s *EventRouteManager) AddReceiver(receiver *notification_v1.Receiver) (bool, error) {
	s.cfgGenerator.appendReceiver(receiver)
	return false, nil
}

func (s *EventRouteManager) UpdateReceiver(old, new *notification_v1.Receiver) (bool, error) {
	if !reflect.DeepEqual(old.Spec, new.Spec) {
		s.cfgGenerator.removeReceiver(old)
		s.cfgGenerator.appendReceiver(new)
	}
	return false, nil
}

func (s *EventRouteManager) DeleteReceiver(receiver *notification_v1.Receiver) error {
	s.cfgGenerator.removeReceiver(receiver)
	return nil
}

func (s *EventRouteManager) AddEventRoute(eventroute *eventmeshv1.EventRoute) (bool, error) {
	s.cfgGenerator.appendEventRoute(eventroute)
	return false, nil
}

func (s *EventRouteManager) UpdateEventRoute(old, new *eventmeshv1.EventRoute) (bool, error) {
	if !reflect.DeepEqual(old.Spec, new.Spec) {
		s.cfgGenerator.removeEventRoute(old)
		s.cfgGenerator.appendEventRoute(new)
	}
	return false, nil
}

func (s *EventRouteManager) DeleteEventRoute(eventroute *eventmeshv1.EventRoute) error {
	s.cfgGenerator.removeEventRoute(eventroute)
	return nil
}

func (s *EventRouteManager) GeneratorConfig() error {
	yaml, err := s.cfgGenerator.generateYaml()
	if err != nil {
		log.Info("generateYaml", err)
	}

	err = ioutil.WriteFile("./config/route.yml", yaml, 0666)
	if err != nil {
		log.Error("GeneratorConfig write file", err)
	}

	return nil
}

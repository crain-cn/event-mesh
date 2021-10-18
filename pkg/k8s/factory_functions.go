package k8s

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
)

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

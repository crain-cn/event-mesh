package events

import (
	"context"
	"fmt"
	"github.com/crain-cn/event-mesh/api/model"
	"github.com/crain-cn/event-mesh/pkg/config"
	"github.com/crain-cn/event-mesh/pkg/dispatch"
	"github.com/crain-cn/event-mesh/pkg/notify"
	"github.com/crain-cn/event-mesh/pkg/provider"
	"github.com/prometheus/alertmanager/types"
	common_model "github.com/prometheus/common/model"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"regexp"
	"strings"
	"time"
)

type EventAction string

type ElEvent struct {
	Cluster   string
	Action    EventAction
	Event     *v1.Event
	StartTime time.Time
	Alerts    provider.Alerts
	T         time.Time
	Reasons   map[string]string
	Workcodes   map[string]string
	ENV       string
}

type EventResult struct {
	err error
}

const (
	EVENT_TYPE_WARRNIGN = "Warning"
	EVENT_TYPE_NORMAL   = "Normal"
	EVENT_TYPE_ERROR    = "Error"
)

func (e *ElEvent) Handle(res chan interface{}) {
	if e.Event.FirstTimestamp.Time.IsZero() {
		e.T = e.Event.EventTime.Time.Local()
	} else {
		e.T = e.Event.LastTimestamp.Time.Local()
	}
	e.log()
	e.insertAlerts()

	e.insertMysql()
	res <- &EventResult{
		err: nil,
	}
}

func (e *ElEvent) Handle2() {
	if e.Event.FirstTimestamp.Time.IsZero() {
		e.T = e.Event.EventTime.Time.Local()
	} else {
		e.T = e.Event.LastTimestamp.Time.Local()
	}
	e.insertMysql()
	go e.insertAlerts()
}


func (e *ElEvent) log() {
	event := e.Event
	logger := log.WithTime(e.T).
		WithFields(logrus.Fields{
			"type":             event.Type,
			"count":            event.Count,
			"event_reason":     event.Reason,
			"source_component": event.Source.Component,
			"source_host":      event.Source.Host,
			"event_timestamp": time.Now().Format("2006-01-02 15:04:05"),
			"namespace": event.InvolvedObject.Namespace,
			"cluster":   e.Cluster,
			"obj_kind":  event.InvolvedObject.Kind,
			"obj_name":  event.InvolvedObject.Name,
			"message":   event.Message,
		})
	switch event.Type {
	case EVENT_TYPE_WARRNIGN:
		logger.Warning()
	case EVENT_TYPE_ERROR:
		logger.Error()
	case EVENT_TYPE_NORMAL:
		logger.Info()
	default:
		logger.Info()
	}
}

//
//func BuildReceiverIntegrations(nc *config.Receiver, tmpl *template.Template) ([]notify.Integration, error) {
//	var (
//		errs         types.MultiError
//		integrations []notify.Integration
//		add          = func(name string, i int, rs notify.ResolvedSender, f func() (notify.Notifier, error)) {
//			n, err := f()
//			if err != nil {
//				errs.Add(err)
//				return
//			}
//			integrations = append(integrations, notify.NewIntegration(n, rs, name, i))
//		}
//	)
//
//	for i, c := range nc.WebhookConfigs {
//		add("webhook", i, c, func() (notify.Notifier, error) { return webhook.New(c, tmpl) })
//	}
//	for i, c := range nc.DogConfigs {
//		add("dog", i, c, func() (notify.Notifier, error) { return dog.New(c, tmpl) })
//	}
//
//	for i, c := range nc.YachConfigs {
//
//		add("yach", i, c, func() (notify.Notifier, error) { return yach.New(c, tmpl) })
//	}
//	for i, c := range nc.WechatConfigs {
//		add("wechat", i, c, func() (notify.Notifier, error) { return wechat.New(c, tmpl) })
//	}
//	if errs.Len() > 0 {
//		return nil, &errs
//	}
//	return integrations, nil
//}

func (e *ElEvent) insertAlerts() {
	event := e.Event
	labelSet := common_model.LabelSet{
		"cluster":      common_model.LabelValue(e.Cluster),
		"namespace":    common_model.LabelValue(event.InvolvedObject.Namespace),
		"obj_kind":     common_model.LabelValue(event.InvolvedObject.Kind),
		"obj_name":     common_model.LabelValue(event.InvolvedObject.Name),
		"severity":     common_model.LabelValue(event.Type),
		"event_reason": common_model.LabelValue(event.Reason),
		"cn_reason":   common_model.LabelValue(e.Reasons[event.Reason]),
		"source_host": common_model.LabelValue(event.Source.Host),
	}

	switch event.InvolvedObject.Kind {
	case "Pod":
		labelSet["pod"] = common_model.LabelValue(event.InvolvedObject.Name)
		reg := regexp.MustCompile(`(.*)-[a-fA-F\d]{1,28}-\w{5}`)
		name := event.InvolvedObject.Name
		dy := reg.FindStringSubmatch(name)
		if dy != nil {
			labelSet["workcode"] = common_model.LabelValue(e.Workcodes[event.InvolvedObject.Namespace + "|" + dy[1]])
		}
	case "Node":
		labelSet["node"] = common_model.LabelValue(event.InvolvedObject.Name)
	case "Deployment":
		labelSet["deployment"] = common_model.LabelValue(event.InvolvedObject.Name)
		labelSet["workcode"] = common_model.LabelValue(e.Workcodes[event.InvolvedObject.Namespace + "|" + event.InvolvedObject.Name])
	case "ReplicaSet":
		labelSet["replicaSet"] = common_model.LabelValue(event.InvolvedObject.Name)
	case "HorizontalPodAutoscaler":
		name := strings.Replace(event.InvolvedObject.Name,"-hpa","",-1)
		labelSet["workcode"] = common_model.LabelValue(e.Workcodes[event.InvolvedObject.Namespace + "|" +name])
	}

	annotations := common_model.LabelSet{
		"message": common_model.LabelValue(event.Message),
	}

	typeAlert := &types.Alert{
		Alert: common_model.Alert{
			Labels:      labelSet,
			Annotations: annotations,
			StartsAt:    e.T,
			EndsAt:      e.T,
		},
		Timeout: false,
	}
	//inputAlert := []*types.Alert{}
	//inputAlert = append(inputAlert, typeAlert)
	//e.Alerts.Put(inputAlert...)

	route := config.StaticRoute
	routes := dispatch.NewRoute(route, nil)
	for _, r := range routes.Match(typeAlert.Labels) {
		nt := config.StaticPipeline.(notify.Stage)
		ctx := notify.WithReceiverName(context.Background(), r.RouteOpts.Receiver)
		ctx = notify.WithGroupKey(ctx, fmt.Sprintf("%s:%s", r.Key(), typeAlert.Labels))
		ctx = notify.WithGroupLabels(ctx, typeAlert.Labels)
		_, _, err := nt.Exec(ctx,typeAlert)
		if err != nil {
			log.WithTime(e.T).
				WithFields(logrus.Fields{
					"message":   err.Error(),
				}).Error()
		}
	}


}

func (e *ElEvent) insertMysql() {
	event := e.Event
	if e.ENV == "" || e.ENV  == "dev" {
		//fmt.Println(model.EventHistory{
		//	Severity:        strings.ToLower(event.Type),
		//	Message:         event.Message,
		//	Reason:          event.Reason,
		//	Datetime:        t,
		//	Namespace:       event.InvolvedObject.Namespace,
		//	Cluster:         e.Cluster,
		//	ObjKind:         event.InvolvedObject.Kind,
		//	ObjName:         event.InvolvedObject.Name,
		//	SourceComponent: event.Source.Component,
		//	SourceHost:      event.Source.Host,
		//})
		log.WithTime(time.Now()).
			WithFields(logrus.Fields{
				"errtype":           "local event ",
				"Severity":        strings.ToLower(event.Type),
				"Message":         event.Message,
				"Reason":          event.Reason,
				"Datetime":        e.T,
				"Namespace":       event.InvolvedObject.Namespace,
				"Cluster":         e.Cluster,
				"ObjKind":         event.InvolvedObject.Kind,
				"ObjName":         event.InvolvedObject.Name,
				"SourceComponent": event.Source.Component,
				"SourceHost":      event.Source.Host,
			}).Info()
	} else {
		err,id := model.AddEventHistory(&model.EventHistory{
			Severity:        strings.ToLower(event.Type),
			Message:         event.Message,
			Reason:          event.Reason,
			Datetime:        e.T,
			Namespace:       event.InvolvedObject.Namespace,
			Cluster:         e.Cluster,
			ObjKind:         event.InvolvedObject.Kind,
			ObjName:         event.InvolvedObject.Name,
			SourceComponent: event.Source.Component,
			SourceHost:      event.Source.Host,
		})
		if err != nil {
			log.WithTime(time.Now()).
				WithFields(logrus.Fields{
					"errtype":           "mysql insert ",
					"errinfo":            err.Error(),
					"id" : fmt.Sprintf("%d",id),
				}).Info()
		} else {
			log.WithTime(time.Now()).
				WithFields(logrus.Fields{
					"errtype":           "mysql insert ",
					"Severity":        strings.ToLower(event.Type),
					"Message":         event.Message,
					"Reason":          event.Reason,
					"Datetime":        e.T,
					"Namespace":       event.InvolvedObject.Namespace,
					"Cluster":         e.Cluster,
					"ObjKind":         event.InvolvedObject.Kind,
					"ObjName":         event.InvolvedObject.Name,
					"SourceComponent": event.Source.Component,
					"SourceHost":      event.Source.Host,
					"id" : fmt.Sprintf("%d",id),
				}).Info()
		}
	}

}

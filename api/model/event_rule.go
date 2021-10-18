package model

import (
	"context"
	"fmt"
	v1 "github.com/crain-cn/event-mesh/pkg/k8s/apis/eventmesh/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
	"time"
)

type EventRule struct {
	ID            uint      `json:"id" gorm:"primarykey"`
	ReceiverRefer uint      `json:"receiver_id" gorm:"index"`
	GroupRefer    uint      `json:"group_id"  gorm:"index"`
	Name          string    `json:"name" example:"normal event notify" gorm:"uniqueIndex"`
	Scope         string    `json:"scope" example:"app"`
	App           uint      `json:"app" example:"fend-demo"`
	Namespace     string    `json:"namespace" example:""`
	Receiver      *Receiver `json:"receiver" example:"bot1" gorm:"foreignKey:ReceiverRefer"`
	Severity      string    `json:"severity" example:"warning"`
	Group         *AppGroup `json:"group" example:"hudong"  gorm:"foreignKey:GroupRefer"`
	//	Labels        []*Label  `json:"labels,omitempty"  gorm:"many2many:notification_event_rule_labels;"`
	Datetime time.Time `json:"datetime" example:"2021-03-03 22:00:00"`
	Events   string    `json:"events"`
	Status   string    `json:"status" example:"On"`
}

func (a *EventRule) TableName() string {
	return "notification_event_rule"
}

type EventRuleListRepose struct {
	Code    int          `json:"code" example:"0"`
	Stat    int          `json:"stat" example:"0"`
	Message string       `json:"msg" example:""`
	Data    []*EventRule `json:"data"`
}

type EventRuleRepose struct {
	Code    int        `json:"code" example:"0"`
	Stat    int        `json:"stat" example:"0"`
	Message string     `json:"msg" example:""`
	Data    *EventRule `json:"data"`
}

func ListEventRule(name string, group uint) []*EventRule {
	var eventRules []*EventRule
	tx := Db.Table("notification_event_rule")
	if len(name) > 0 {
		tx = tx.Where("name like ?", "%"+name+"%")
	}
	if group > 0 {
		tx = tx.Where("group_refer =?", group)
	}
	tx.Order("id desc").Limit(100).Find(&eventRules)
	for key, r := range eventRules {
		r.Group = getAppGroup(r.GroupRefer)
		r.Receiver = GetReceiverById(r.ReceiverRefer)
		eventRules[key] = r
	}
	return eventRules
}

func GetEventRule(id uint) *EventRule {
	rule := &EventRule{}
	Db.Where(&EventRule{ID: id}).First(rule)
	return rule
}

func UpdateEventRule(old, update *EventRule) (error, *EventRule) {
	result := Db.Model(&EventRule{}).Where("id =?", update.ID).Updates(update)
	if result.Error == nil {
		deleteEventResource(old)
		if update.Status == STATUS_ON {
			createEventResource(update)
		}
	}
	return nil, GetEventRule(update.ID)
}

func DeleteEventRule(id uint) error {
	r := GetEventRule(id)
	deleteEventResource(r)
	tx := Db.Where("id =?", id).Delete(&EventRule{})
	return tx.Error
}

func AddEventRule(r *EventRule) (*EventRule, error) {
	result := Db.Create(r)
	if result.Error == nil {
		createEventResource(r)
	}
	return nil, result.Error
}

func createEventResource(r *EventRule) {
	receiver := GetReceiverById(r.ReceiverRefer)
	events := strings.Replace(r.Events, ",", "|", -1)
	var xesApps []*XesApp
	if r.App > 0 {
		xesApp := getDeployemtByApp(r.GroupRefer, r.App)
		xesApps = append(xesApps, xesApp)
	} else {
		xesApps = getNamespacesByGroup(r.GroupRefer)
	}
	for _, xesApp := range xesApps {
		namesapce := xesApp.Namespace
		deployment := xesApp.Deployment

		var labels map[string]string
		var matchers = []v1.Matcher{
			{Name: "namespace", Value: namesapce},
			{Name: "obj_kind", Value: "Pod"},
			{Name: "event_reason", Value: events, Regex: true},
			{Name: "obj_name", Value: fmt.Sprintf("%s.*", xesApp.Deployment), Regex: true},
		}

		_, err := Clients.client.EventmeshV1().EventRoutes(namesapce).Create(context.TODO(), &v1.EventRoute{
			ObjectMeta: metav1.ObjectMeta{
				Name:   strings.ToLower(deployment + "-events"),
				Labels: labels,
			},
			Spec: v1.EventRouteSpec{Route: &v1.Route{
				Receiver: receiver.Name,
				Matchers: matchers,
			}},
		}, metav1.CreateOptions{})

		if err != nil {
			log.Error(err)
		}

	}
}

func deleteEventResource(r *EventRule) {
	var xesApps []*XesApp
	if r.App > 0 {
		xesApp := getDeployemtByApp(r.GroupRefer, r.App)
		xesApps = append(xesApps, xesApp)
	} else {
		xesApps = getNamespacesByGroup(r.GroupRefer)
	}
	for _, xesApp := range xesApps {
		namesapce := xesApp.Namespace
		deployment := xesApp.Deployment
		err := Clients.client.EventmeshV1().EventRoutes(namesapce).Delete(context.TODO(),
			strings.ToLower(deployment+"-events"),
			metav1.DeleteOptions{},
		)
		if err != nil {
			log.Info("delete EventResource", err)
		}
	}
}

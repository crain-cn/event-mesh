package model

import (
	"context"
	v1 "github.com/crain-cn/event-mesh/pkg/k8s/apis/eventmesh/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
	"time"
)

type EventHistory struct {
	ID              uint      `gorm:"primarykey" json:"id"`
	GroupRefer      uint      `json:"group_id" gorm:"index"`
	Reason          string    `json:"reason" example:"NotFound"`
	Severity        string    `json:"severity" example:"warning"`
	Cluster         string    `json:"cluster" example:"aaa"`
	Namespace       string    `json:"namespace" example:"jichujiagou_common"`
	ObjKind         string    `json:"obj_kind" example:"deployment"`
	ObjName         string    `json:"obj_name" example:"fend-demo"`
	Datetime        time.Time `json:"datetime" example:"2021-03-03 22:00:00"`
	Message         string    `json:"message" example:"fend-demo Not found"`
	SourceComponent string    `json:"source_component"`
	SourceHost      string    `json:"source_host"`
}

func (a *EventHistory) TableName() string {
	return "notification_event_history"
}

type EventTends struct {
	DateTime time.Time        `json:"datetime" example:"datetime"`
	Counts   []*SeverityCount `json:"counts"`
}

const TIME_LAYOUT = "2006-01-02 15:04:05"

type TendCount struct {
	CriticalCount int       `json:"critical_count"`
	WarningCount  int       `json:"warning_count"`
	NormalCount   int       `json:"normal_count"`
	DateTime      time.Time `json:"date_time" example:"date_time" gorm:"column:date;"`
}

type EventHistoryWithTends struct {
	Tends  []*TendCount    `json:"tends"`
	Events []*EventHistory `json:"events"`
}

type EventHistoryRepose struct {
	Code    int             `json:"code" example:"0"`
	Stat    int             `json:"stat" example:"0"`
	Message string          `json:"msg" example:""`
	Data    []*EventHistory `json:"data"`
}

var Events = []string{
	"ErrImageNeverPull", "BackOff", "Pulled",
}

var ReasonsSample = map[string]string{
	"ScalingReplicaSet": "副本伸缩",
	"SuccessfulDelete":  "副本完全删除",
	"ErrImageNeverPull": "Pod镜像拉取失败",
	"BackOff":           "Pod启动失败",
	"Killing":           "Pod杀死",
	"Unhealthy":         "Pod Unhealthy",
	"UPDATE":            "ingress UPDATE",
}

type EventReasonsRepose struct {
	Code    int               `json:"code" example:"0"`
	Stat    int               `json:"stat" example:"0"`
	Message string            `json:"msg" example:""`
	Data    map[string]string `json:"data"`
}

func ListEventHistory(group uint, startTime string, endTime string) []*EventHistory {
	var histories []*EventHistory
	start, _ := time.Parse(TIME_LAYOUT, startTime)
	end, _ := time.Parse(TIME_LAYOUT, endTime)
	tx := Db.Table("notification_event_history").
		Where("group_refer = ?", group).
		Where("datetime  > ?", start).
		Where("datetime  < ?", end).Limit(100).Order("id Desc")
	tx.Scan(&histories)
	return histories
}

func GetEventTends(group uint, startTime string, endTime string) []*TendCount {
	var items []*SeverityCount

	t1, _ := time.Parse(TIME_LAYOUT, startTime)
	t2, _ := time.Parse(TIME_LAYOUT, endTime)

	tx := Db.Table("notification_event_history").
		Select("DATE(datetime) as datetime, count(1) as count, severity").
		Group("date(datetime)").Group("severity").
		Where("group_refer = ?", group).
		Where("datetime  > ?", t1).
		Where("datetime  < ?", t2)

	//stmt := tx.Session(&gorm.Session{DryRun: true}).Scan(items).Statement
	//log.Info(stmt.SQL.String())
	tx.Scan(&items)
	var counts []*TendCount
	for _, v := range items {
		switch v.Severity {
		case "normal":
			counts = append(counts, &TendCount{
				NormalCount: v.Count,
				DateTime:    v.Datetime,
			})
		case "warning":
			counts = append(counts, &TendCount{
				WarningCount: v.Count,
				DateTime:     v.Datetime,
			})
		case "critical":
			counts = append(counts, &TendCount{
				CriticalCount: v.Count,
				DateTime:      v.Datetime,
			})
		}
	}
	return counts
}

func AddEventHistory(r *EventHistory) (error,uint) {
	// @todo 导致panic
	//var xesApp *XesApp
	//switch r.ObjKind {
	//case "Pod":
	//	xesApp = SplitForGetXesApp(r, 2)
	//case "Deployment":
	//	xesApp = SplitForGetXesApp(r, 0)
	//case "Ingress":
	//	xesApp = SplitForGetXesApp(r, 1)
	//case "Service":
	//	xesApp = SplitForGetXesApp(r, 1)
	//case "ReplicaSet":
	//	xesApp = SplitForGetXesApp(r, 1)
	//case "Node":
	//
	//}
	//if xesApp != nil && xesApp.GroupId > 0 {
	//	r.GroupRefer = xesApp.GroupId
	//}
	result := Db.Create(r)
	return result.Error,r.ID
}

func SplitForGetXesApp(r *EventHistory, subLen int) *XesApp {
	var deployment string
	if subLen > 0 {
		podArr := strings.Split(r.ObjName, "-")
		deployment = strings.Join(podArr[0:len(podArr)-subLen], "-")
	} else {
		deployment = r.ObjName
	}
	return getGroupByApp(r.Namespace, deployment)
}

func GetEventsResource() (*v1.EventRouteList, error) {
	return Clients.client.EventmeshV1().EventRoutes("default").List(context.TODO(), metav1.ListOptions{})
}
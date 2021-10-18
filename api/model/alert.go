package model

import (
	"strings"
	"time"
)

type AlertHistory struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	GroupRefer  uint      `json:"group_id"`
	AlertName   string    `json:"alert_name" example:"alert"`
	Severity    string    `json:"severity" example:"warning"`
	Cluster     string    `json:"cluster" example:"dev"`
	Namespace   string    `json:"namespace" example:"jichujiagou_common"`
	Node        string    `json:"node" example:"lghx-74-81"`
	Pod         string    `json:"pod" example:"fend-demo-xx12"`
	Message     string    `json:"message" example:"contanier oom kill"`
	Labels      string    `json:"labels" example:""`
	Annotations string    `json:"annotations" example:""`
	StartAt     time.Time `json:"start_at" example:"2021-03-03 22:00:00"`
	EndsAt      time.Time `json:"ends_at" example:"2021-03-03 22:00:00"`
}

func (a *AlertHistory) TableName() string {
	return "notification_alert_history"
}

type AlertTendCount struct {
	DateTime time.Time `json:"datetime" example:"date_time"`
	Count    int       `json:"counts"`
}

type AlertTendCounts []*AlertTendCount
type AlertTends map[string]*AlertTendCounts

type SeverityCount struct {
	Severity string    `json:"severity"  gorm:"column:severity;"`
	Datetime time.Time `json:"datetime"  gorm:"column:datetime;"`
	Count    int       `json:"count" gorm:"column:count;"`
}

type AlertHistoryWithTends struct {
	Tends  []*TendCount    `json:"tends"`
	Alerts []*AlertHistory `json:"alerts"`
}

type AlertHistoryRepose struct {
	Code    int                      `json:"code" example:"0"`
	Stat    int                      `json:"stat" example:"0"`
	Message string                   `json:"msg" example:""`
	Data    []*AlertHistoryWithTends `json:"data"`
}

var MetricsSample = map[string]string{
	"PodMemExceedRequest": "内存占用",
	"PodCPUExceedRequest": "CPU占用",
	"pod-status-failed":   "pod状态异常",
}

type MetricsRepose struct {
	Code    int               `json:"code" example:"0"`
	Stat    int               `json:"stat" example:"0"`
	Message string            `json:"msg" example:""`
	Data    map[string]string `json:"data"`
}

func ListAlertHistory(group uint, startTime string, endTime string) []*AlertHistory {
	var histories []*AlertHistory
	start, _ := time.Parse(TIME_LAYOUT, startTime)
	end, _ := time.Parse(TIME_LAYOUT, endTime)
	tx := Db.Table("notification_alert_history").
		Where("group_refer = ?", group).
		Where("start_at  > ?", start).
		Where("start_at  < ?", end).Limit(100).Order("id Desc")
	tx.Scan(&histories)
	return histories
}

func GetAlertTends(group uint, startTime string, endTime string) []*TendCount {
	var items []*SeverityCount

	t1, _ := time.Parse(TIME_LAYOUT, startTime)
	t2, _ := time.Parse(TIME_LAYOUT, endTime)

	tx := Db.Table("notification_alert_history").
		Select("DATE(start_at) as datetime, count(1) as count, severity").
		Group("date(start_at)").Group("severity").
		Where("group_refer = ?", group).
		Where("start_at  > ?", t1).
		Where("start_at  < ?", t2)

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

func AddAlertHistory(r *AlertHistory) error {
	var deployment string
	if len(r.Pod) > 0 {
		podArr := strings.Split(r.Pod, "-")
		deployment = strings.Join(podArr[0:len(podArr)-2], "-")
		xesApp := getGroupByApp(r.Namespace, deployment)
		if xesApp != nil && xesApp.GroupId > 0 {
			r.GroupRefer = xesApp.GroupId
		}
	}
	result := Db.Create(r)
	return result.Error
}

package model

import (
	"context"
	"fmt"
	v1 "github.com/crain-cn/event-mesh/pkg/k8s/apis/eventmesh/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
	"time"
)

type AlertRule struct {
	ID            uint      `gorm:"primarykey" json:"id"`
	ReceiverRefer uint      `json:"receiver_id" gorm:"index"`
	GroupRefer    uint      `json:"group_id" gorm:"index"`
	Name          string    `json:"name" example:"normal event notify" gorm:"uniqueIndex"`
	App           uint      `json:"app" example:"fend-demo"`
	Receiver      *Receiver `json:"receiver" example:"bot1"  gorm:"foreignKey:ReceiverRefer"`
	Severity      string    `json:"severity" example:"warning"`
	Scope         string    `json:"scope" example:"app"`
	Group         *AppGroup `json:"group" example:"hudong" gorm:"foreignKey:GroupRefer"`
	//	Labels        []*Label    `json:"labels,omitempty"  gorm:"many2many:notification_alert_rule_labels;"`
	RuleType  string      `json:"rule_type" example:"prom_rules"`
	PromRules []*RuleExpr `json:"prom_rules"  gorm:"foreignKey:AlertRuleID"`
	PromQL    string      `json:"prom_ql" example:""`
	Datetime  time.Time   `json:"datetime" example:"2021-03-03 22:00:00"`
	Status    string      `json:"status" example:"On"`
}

func (a *AlertRule) TableName() string {
	return "notification_alert_rule"
}

type RuleExpr struct {
	AlertRuleID  uint
	Alert        string `json:"alert" example:"NodeDown"`
	For          string `json:"for"  example:"10m"`
	ExprFunc     string `json:"expr_func" example:"avg"`
	ExprPeroid   string `json:"expr_peroid" example:"5m"`
	ExprOperator string `json:"expr_operator" example:"<="`
	ExprValue    string `json:"expr_value" example:"0.4"`
}

func (a *RuleExpr) TableName() string {
	return "notification_alert_rule_expr"
}

type AlertRuleListRepose struct {
	Code    int          `json:"code" example:"0"`
	Stat    int          `json:"stat" example:"0"`
	Message string       `json:"msg" example:""`
	Data    []*AlertRule `json:"data"`
}

type AlertRuleRepose struct {
	Code    int        `json:"code" example:"0"`
	Stat    int        `json:"stat" example:"0"`
	Message string     `json:"msg" example:""`
	Data    *AlertRule `json:"data"`
}

var prometheusRule = []*RuleExpr{
	{
		Alert:        "NodeDown",
		For:          "10m",
		ExprPeroid:   "5m",
		ExprOperator: "<=",
		ExprValue:    "0.5",
	},
}

func ListAlertRule(name string, group uint) []*AlertRule {
	var alertRules []*AlertRule
	tx := Db.Model(&AlertRule{})
	if len(name) > 0 {
		tx = tx.Where("name like ?", "%"+name+"%")
	}
	if group > 0 {
		tx = tx.Where("group_refer = ?", group)
	}
	tx.Order("id desc").Limit(100).Find(&alertRules)
	for key, r := range alertRules {
		r.Group = getAppGroup(r.GroupRefer)
		r.Receiver = GetReceiverById(r.ReceiverRefer)
		if r.RuleType == "prom_rules" {

		}
		alertRules[key] = r
	}
	return alertRules
}

func GetAlertRule(id uint) (error, *AlertRule) {
	alertRule := &AlertRule{}
	result := Db.Where(&AlertRule{ID: id}).First(alertRule)
	if result.Error == nil && alertRule.RuleType == "prom_rules" {
		var promRules = []*RuleExpr{}
		Db.Where("alert_rule_id =?", alertRule.ID).Limit(100).Find(&promRules)
		if len(promRules) > 0 {
			alertRule.PromRules = promRules
		}
	}
	return result.Error, alertRule
}

const (
	STATUS_ON  = "On"
	STATUS_OFF = "Off"
)

func UpdateAlertRule(old *AlertRule, update *AlertRule) (error, *AlertRule) {

	switch update.RuleType {
	case "promQL":
		Db.Where("alert_rule_id =?", update.ID).Delete(&RuleExpr{})
		break
	case "prom_rules":
		if old.RuleType != update.RuleType {
			for k, v := range update.PromRules {
				v.AlertRuleID = update.ID
				update.PromRules[k] = v
			}
			Db.Create(&update.PromRules)
		} else {
			for _, v := range update.PromRules {
				Db.Where(&RuleExpr{AlertRuleID: update.ID, Alert: v.Alert}).Updates(v)
			}
		}
	}
	result := Db.Model(&AlertRule{}).Where("id =?", update.ID).Updates(update)
	_, r := GetAlertRule(update.ID)
	if result.Error == nil {
		deleteAlertResource(old)
		if update.Status == STATUS_ON {

			createAlertResource(update)
		}
	}
	return result.Error, r
}

func DeleteAlertRule(id uint) error {
	err, r := GetAlertRule(id)
	if err != nil {
		return err
	}
	deleteAlertResource(r)
	Db.Where("alert_rule_id =?", id).Delete(&RuleExpr{})
	tx := Db.Where("id =?", id).Delete(&AlertRule{})
	return tx.Error
}

func AddAlertRule(r *AlertRule) error {
	result := Db.Create(r)
	if result.Error == nil {
		createAlertResource(r)
	}
	return result.Error
}

func getDeployemtByApp(group uint, app uint) *XesApp {
	xesApp := &XesApp{}
	tx := Db.Table("k8s_platform.xes_cloud_app").
		Select("deployment, namespace").
		Where("group_id  = ?", group).
		Where("id  = ?", app)

	//stmt := tx.Session(&gorm.Session{DryRun: true}).Scan(items).Statement
	//log.Info(stmt.SQL.String())
	tx.Find(xesApp)
	return xesApp
}

func createAlertResource(r *AlertRule) {
	var matchers []v1.Matcher
	var xesApps []*XesApp
	labels := map[string]string{}

	if r.App > 0 {
		xesApp := getDeployemtByApp(r.GroupRefer, r.App)
		xesApps = append(xesApps, xesApp)
	} else {
		xesApps = getNamespacesByGroup(r.GroupRefer)
	}
	if r.RuleType == "prom_rules" && r.PromRules == nil {
		var promRules = []*RuleExpr{}
		Db.Where("alert_rule_id =?", r.ID).Limit(100).Find(&promRules)
		r.PromRules = promRules
	}

	for _, xesApp := range xesApps {
		var rules []string
		namesapce := xesApp.Namespace
		for _, rule := range r.PromRules {
			rules = append(rules, rule.Alert)
		}
		matchers = []v1.Matcher{
			{Name: "alertname", Value: strings.Join(rules, "|"), Regex: true},
			{Name: "namespace", Value: xesApp.Namespace},
			{Name: "pod", Value: fmt.Sprintf("%s.*", xesApp.Deployment), Regex: true},
		}
		receiver := GetReceiverById(r.ReceiverRefer)
		_, err := Clients.client.EventmeshV1().EventRoutes(namesapce).Create(context.TODO(), &v1.EventRoute{
			ObjectMeta: metav1.ObjectMeta{
				Name:   strings.ToLower(xesApp.Deployment + "-alerts"),
				Labels: labels,
			},
			Spec: v1.EventRouteSpec{Route: &v1.Route{
				Receiver: receiver.Name,
				Matchers: matchers,
			}},
		}, metav1.CreateOptions{})
		if err != nil {
			log.Info(err)
		}
	}
}

func deleteAlertResource(r *AlertRule) {
	var xesApps []*XesApp
	if r.App > 0 {
		xesApp := getDeployemtByApp(r.GroupRefer, r.App)
		xesApps = append(xesApps, xesApp)
	} else {
		xesApps = getNamespacesByGroup(r.GroupRefer)
	}
	for _, xesApp := range xesApps {
		namespace := xesApp.Namespace
		var rules []string
		for _, rule := range r.PromRules {
			rules = append(rules, rule.Alert)
		}

		err := Clients.client.EventmeshV1().EventRoutes(namespace).Delete(
			context.TODO(),
			strings.ToLower(xesApp.Deployment+"-alerts"),
			metav1.DeleteOptions{},
		)
		if err != nil {
			log.Error("delete AlertResource", err)
		}
	}
}

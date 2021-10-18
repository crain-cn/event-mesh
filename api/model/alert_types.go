package model

import "time"

type AlertLables struct {
	Cluster    string `json:"cluster"`
	Alertname  string `json:"alertname"`
	Namespace  string `json:"namespace"`
	Node       string `json:"node"`
	Pod        string `json:"pod"`
	Deployment string `json:"deployment"`
	ReplicaSet string `json:"replicaset"`
	Prometheus string `json:"prometheus"`
	Severity   string `json:"severity"`
}

type AlertAnnotations struct {
	Message    string `json:"message"`
	RunbookURL string `json:"runbook_url"`
}

type alert struct {
	Status       string            `json:"status"`
	Labels       *AlertLables      `json:"labels"`
	Annotations  *AlertAnnotations `json:"annotations"`
	StartsAt     time.Time         `json:"startsAt"`
	EndsAt       time.Time         `json:"endsAt"`
	GeneratorURL string            `json:"generatorURL"`
	Fingerprint  string            `json:"fingerprint"`
}

type groupLabels struct {
	Namespace string `json:"namespace"`
}

type commonLabels struct {
	Namespace  string `json:"namespace"`
	Node       string `json:"node"`
	Prometheus string `json:"prometheus"`
	Severity   string `json:"severity"`
}

type commonAnnotations struct{}

type AlertReceive struct {
	Receiver          string             `json:"receiver"`
	Status            string             `json:"status"`
	Alerts            []*alert           `json:"alerts"`
	GroupLabels       *groupLabels       `json:"groupLabels"`
	CommonLabels      *commonLabels      `json:"commonLabels"`
	CommonAnnotations *commonAnnotations `json:"commonAnnotations"`
	ExternalURL       string             `json:"externalURL"`
	Version           string             `json:"version"`
	GroupKey          string             `json:"groupKey"`
	TruncatedAlerts   int                `json:"truncatedAlerts"`
}

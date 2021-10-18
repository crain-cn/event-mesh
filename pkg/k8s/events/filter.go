package events

import (
	"github.com/linvon/cuckoo-filter"
	v1 "k8s.io/api/core/v1"
	"time"
)

var (
	filterNormalResons = FilterList{
		"SuccessfulCreate",
		"Pulled",
		"Created",
		"Started",
		"Scheduled",
		"Pulling",
		"SuccessfulDelete",
		"SawCompletedJob",
		"SandboxChanged",
		"Provisioning",
		"SuccessfulAttachVolume",
		"ProvisioningSucceeded",
		"ExternalProvisioning",
		"NodeReady",
	}
	filterWarnningResons = FilterList{
		"PipelineRunFailed",
	}
	notFilterNormalResons = FilterList{
		"TaintManagerEviction",
		"ScalingReplicaSet",
		"Killing",
		"BackOff",
		"NodeNotReady",
	}
	notFilterWarnningResons = FilterList{
		"FailedScheduling",
		"FailedMount",
		"FailedCreatePodSandBox",
		"DeadlineExceeded",
		"Failed",
		"BackoffLimitExceeded",
		"BackOff",
		"Unhealthy",
		"FailedComputeMetricsReplicas",
		"FailedGetResourceMetric",
		"FailedCreatePodContainer",
		"FailedGetScale",
		"FailedSync",
		"FailedToUpdateEndpoint",
	}
	filterComponents = FilterList{
		//"job-controller",
		//"cronjob-controller",
		"pipeline-controller",
		"taskrun-controller",
	}
)

const DefaultFilterCount = 2

type FilterList []string
type eventFilter struct {
	cf         *cuckoo.Filter
	filters    map[string]FilterList
	notFilters map[string]FilterList
	startTime  time.Time
}

func NewEventFilter(startTime time.Time, filterList FilterList, notFilterList FilterList) *eventFilter {
	defaultFilterList := map[string]FilterList{
		"Components": filterComponents,
		"Warnning":   filterWarnningResons,
		"Normal":     filterNormalResons,
	}
	defaultNotFilterList := map[string]FilterList{
		"Warnning": notFilterWarnningResons,
		"Normal":   notFilterNormalResons,
	}

	cf := cuckoo.NewFilter(1, 9, 10, cuckoo.TableTypePacked)
	eventFilter := &eventFilter{
		cf:         cf,
		filters:    defaultFilterList,
		notFilters: defaultNotFilterList,
		startTime:  startTime,
	}
	eventFilter.InitDefault()
	eventFilter.AppendFilterList(filterList)
	eventFilter.AppendNotFilterList(notFilterList)
	return eventFilter
}

func ToFilterList(filters []string) FilterList {
	var list FilterList
	for _, filter := range filters {
		list = append(list, filter)
	}
	return list
}

func (f *eventFilter) InitDefault() {
	for _, filterList := range f.filters {
		for _, filter := range filterList {
			f.cf.AddUnique([]byte(filter))
		}
	}
	for _, filterList := range f.notFilters {
		for _, notFilter := range filterList {
			f.cf.Delete([]byte(notFilter))
		}
	}
}

func (f *eventFilter) AppendFilterList(filterList FilterList) {
	for _, filter := range filterList {
		f.cf.AddUnique([]byte(filter))
	}
}

func (f *eventFilter) AppendNotFilterList(filterList FilterList) {
	for _, notFilter := range filterList {
		f.cf.Delete([]byte(notFilter))
	}
}

func (f *eventFilter) Filter(event *v1.Event) bool {
	if f.filterTime(event) {
		return true
	}
	if f.filterComponent(event.Source.Component) {
		return true
	}

	//if f.filterCount(event.Count) {
	//	return true
	//}

	return f.filterReason(event.Reason)
}

func (f *eventFilter) filterComponent(component string) bool {
	return f.cf.Contain([]byte(component))
}

func (f *eventFilter) filterReason(Reason string) bool {
	return f.cf.Contain([]byte(Reason))
}

func (f *eventFilter) filterTime(event *v1.Event) bool {
	if !event.FirstTimestamp.Time.IsZero() && f.startTime.After(event.LastTimestamp.Time) {
		return true
	}
	if !event.EventTime.Time.IsZero() && f.startTime.After(event.EventTime.Time) {
		return true
	}
	return false
}

func (f *eventFilter) filterCount(count int32) bool {
	if count >= DefaultFilterCount {
		return true
	}
	return false
}

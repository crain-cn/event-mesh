package events

import (
	"github.com/crain-cn/event-mesh/pkg/config"
	"github.com/crain-cn/event-mesh/pkg/dispatch"
	eventmesh_v1 "github.com/crain-cn/event-mesh/pkg/k8s/apis/eventmesh/v1"
	notification_v1 "github.com/crain-cn/event-mesh/pkg/k8s/apis/notification/v1"
	"github.com/crain-cn/event-mesh/pkg/labels"
	commoncfg "github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	"gopkg.in/yaml.v2"
	"net/url"
	"sync"
)

type configGenerator struct {
	Route     *config.Route      `yaml:"route,omitempty" json:"routes,omitempty"`
	Receivers []*config.Receiver `yaml:"receivers,omitempty" json:"receivers,omitempty"`
	rwmutex   *sync.RWMutex
}

func NewConfigGenerator() *configGenerator {
	return &configGenerator{
		rwmutex: new(sync.RWMutex),
		Route: &config.Route{
			Receiver: "default",
		},
		Receivers: []*config.Receiver{
			{Name: "default"},
		},
	}
}
func (cg *configGenerator) removeReceiver(in *notification_v1.Receiver) {
	cg.rwmutex.Lock()
	var receivers []*config.Receiver
	for _, v := range cg.Receivers {
		if v.Name != in.Name {
			receivers = append(receivers, v)
		} else {
			log.Info("Receiver has exsits:", in.Name)
		}
	}
	cg.Receivers = receivers
	cg.rwmutex.Unlock()
}

func (cg *configGenerator) removeEventRoute(in *eventmesh_v1.EventRoute) {
	cg.rwmutex.Lock()
	var routes []*config.Route
	for _, v := range cg.Route.Routes {
		if v.Name != in.Name {
			routes = append(routes, v)
		} else {
			log.Info("EventRoute has exsits:", in.Name)
		}
	}
	cg.Route.Routes = routes
	cg.Route.Continue = false
	cg.rwmutex.Unlock()
}

func (cg *configGenerator) appendEventRoute(in *eventmesh_v1.EventRoute) {
	cg.rwmutex.Lock()
	var groupBy []model.LabelName
	var matchers config.Matchers
	if in.Spec.Route.GroupBy != nil {
		for _, l := range in.Spec.Route.GroupBy {
			labelName := model.LabelName(l)
			if !labelName.IsValid() {
				//return fmt.Errorf("invalid label name %q in group_by list", l)
			}
			groupBy = append(groupBy, labelName)
		}
	}

	if in.Spec.Route.Matchers != nil {
		for _, v := range in.Spec.Route.Matchers {
			var matcher = &labels.Matcher{}
			if v.Regex {
				matcher, _ = labels.NewMatcher(labels.MatchRegexp, v.Name, v.Value)
			} else {
				matcher, _ = labels.NewMatcher(labels.MatchEqual, v.Name, v.Value)
			}
			matchers = append(matchers, matcher)
		}
	}
	groupWait := model.Duration(dispatch.DefaultRouteOpts.GroupWait)
	groupInterval := model.Duration(dispatch.DefaultRouteOpts.GroupInterval)
	repeatInterval := model.Duration(dispatch.DefaultRouteOpts.RepeatInterval)

	if in.Spec.Route.RepeatInterval != "" {
		rInterval ,err := model.ParseDuration(in.Spec.Route.RepeatInterval)
		if err == nil {
			repeatInterval = rInterval
		}
	}

	if in.Spec.Route.GroupInterval != "" {
		gInterval ,err := model.ParseDuration(in.Spec.Route.GroupInterval)
		if err == nil {
			groupInterval = gInterval
		}
	}

	if in.Spec.Route.GroupWait != "" {
		gWait ,err := model.ParseDuration(in.Spec.Route.GroupWait)
		if err == nil {
			groupWait = gWait
		}
	}

	route := &config.Route{
		Name:           in.Name,
		Receiver:       in.Spec.Route.Receiver,
		GroupBy:        groupBy,
		Matchers:       matchers,
		Continue:       false,
		GroupWait:      &groupWait,
		GroupInterval:  &groupInterval,
		RepeatInterval: &repeatInterval,
	}
	cg.Route.Routes = append(cg.Route.Routes, route)
	cg.rwmutex.Unlock()
}

func (cg *configGenerator) appendReceiver(in *notification_v1.Receiver) {
	cg.rwmutex.RLock()
	receiver := &config.Receiver{
		Name: in.Name,
	}

	var webhookConfigs []*config.WebhookConfig
	var dogConfigs []*config.DogConfig
	var yachConfigs []*config.YachConfig

	if in.Spec.WebhookConfig != nil {
		if len(*in.Spec.WebhookConfig.URL) > 0 {
			webhookConfig, _ := cg.convertWebhookConfig(in.Spec.WebhookConfig)
			receiver.WebhookConfigs = append(webhookConfigs, webhookConfig)
		}
	}

	if in.Spec.DogConfig != nil {
		if in.Spec.DogConfig.TaskId > 0 {
			dogConfig, _ := cg.convertDogConfig(in.Spec.DogConfig)
			receiver.DogConfigs = append(dogConfigs, dogConfig)
		}
	}

	if in.Spec.YachConfig != nil {
		if len(in.Spec.YachConfig.AccessToken) > 0 {
			yachConfig, _ := cg.convertYachConfig(in.Spec.YachConfig)
			receiver.YachConfigs = append(yachConfigs, yachConfig)
		}
	}
	cg.Receivers = append(cg.Receivers, receiver)
	cg.rwmutex.RUnlock()
}

func (cg *configGenerator) generateYaml() ([]byte, error) {
	cg.rwmutex.RLock()
	byte, err := yaml.Marshal(cg)
	cg.rwmutex.RUnlock()
	return byte, err
}

func (cg *configGenerator) convertDogConfig(in *notification_v1.DogConfig) (*config.DogConfig, error) {
	out := &config.DogConfig{
		NotifierConfig: config.NotifierConfig{
			VSendResolved: true,
		},
		TaskId: in.TaskId,
	}

	if in.MaxAlerts > 0 {
		out.MaxAlerts = uint64(in.MaxAlerts)
	}

	return out, nil
}

func (cg *configGenerator) convertYachConfig(in *notification_v1.YachConfig) (*config.YachConfig, error) {
	out := &config.YachConfig{
		NotifierConfig: config.NotifierConfig{
			VSendResolved: true,
		},
		AccessToken: in.AccessToken,
		Secret:      in.Secret,
	}

	return out, nil
}

func (cg *configGenerator) convertWebhookConfig(in *notification_v1.WebhookConfig) (*config.WebhookConfig, error) {
	u, err := url.Parse(*in.URL)
	if err != nil {
		//require.NoError(t, err)
	}

	out := &config.WebhookConfig{
		NotifierConfig: config.NotifierConfig{
			VSendResolved: true,
		},
		URL:        &config.URL{URL: u},
		HTTPConfig: &commoncfg.HTTPClientConfig{},
	}

	if in.MaxAlerts > 0 {
		out.MaxAlerts = uint64(in.MaxAlerts)
	}

	return out, nil
}

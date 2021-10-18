package module

import (
	"context"
	"github.com/crain-cn/event-mesh/pkg/config"
	"github.com/crain-cn/event-mesh/pkg/dispatch"
	"github.com/crain-cn/event-mesh/pkg/logging"
	"github.com/crain-cn/event-mesh/pkg/logging/logfields"
	"github.com/crain-cn/event-mesh/pkg/notify"
	"github.com/crain-cn/event-mesh/pkg/notify/webhook"
	"github.com/crain-cn/event-mesh/pkg/notify/wechat"
	"github.com/crain-cn/event-mesh/pkg/notify/yach"
	"github.com/crain-cn/event-mesh/pkg/provider"
	"github.com/crain-cn/event-mesh/pkg/provider/mem"
	"github.com/crain-cn/event-mesh/pkg/template"
	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	"github.com/prometheus/alertmanager/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func SetALertMemProvider() (provider.Alerts, types.Marker) {
	var alertGCInterval time.Duration
	marker := types.NewMarker(prometheus.NewRegistry())

	alertGCInterval, _ = time.ParseDuration("30m")
	alerts, err := mem.NewAlerts(context.Background(), marker, alertGCInterval)
	if err != nil {
		logging.DefaultLogger.WithError(err).Error()
	}
	defer alerts.Close()
	return alerts, marker
}

func RunAlertDispatch(o options, alerts provider.Alerts, marker types.Marker) int {

	var retention time.Duration
	retention, _ = time.ParseDuration("120h")

	configCoordinator := config.NewCoordinator(
		o.configFile,
		prometheus.DefaultRegisterer,
	)
	configLogger := configCoordinator.Log()

	var disp *dispatch.Dispatcher
	defer disp.Stop()

	pipelineBuilder := notify.NewPipelineBuilder(prometheus.DefaultRegisterer)

	timeoutFunc := func(d time.Duration) time.Duration {
		if d < notify.MinTimeout {
			d = notify.MinTimeout
		}
		return d
	}
	dispMetrics := dispatch.NewDispatcherMetrics(prometheus.DefaultRegisterer)

	configCoordinator.Subscribe(func(conf *config.Config) error {
		tmpl, err := template.FromGlobs(conf.Templates...)
		if err != nil {
			return errors.Wrap(err, "failed to parse templates")
		}
		amdUrl, _ := url.Parse("http://127.0.0.1")
		tmpl.ExternalURL = amdUrl
		// Build the routing tree and record which receivers are used.
		routes := dispatch.NewRoute(conf.Route, nil)
		activeReceivers := make(map[string]struct{})
		routes.Walk(func(r *dispatch.Route) {
			activeReceivers[r.RouteOpts.Receiver] = struct{}{}
		})

		// Build the map of receiver to integrations.
		receivers := make(map[string][]notify.Integration, len(activeReceivers))
		var integrationsNum int
		for _, rcv := range conf.Receivers {
			if _, found := activeReceivers[rcv.Name]; !found {
				// No need to build a receiver if no route is using it.
				configLogger.WithFields(logrus.Fields{
					"msg":      "skipping creation of receiver not referenced by any route",
					"receiver": rcv.Name,
				}).Info()
				continue
			}
			integrations, err := buildReceiverIntegrations(rcv, tmpl)
			if err != nil {
				return err
			}
			// rcv.Name is guaranteed to be unique across all receivers.
			receivers[rcv.Name] = integrations
			integrationsNum += len(integrations)
		}


		disp.Stop()
		pipeline := pipelineBuilder.New(
			receivers,
			//	waitFunc,
			//	inhibitor,
			//	silencer,
			//	notificationLog,
			//	peer,
		)
		config.StaticRoute = conf.Route
		config.StaticPipeline = pipeline
		disp = dispatch.NewDispatcher(alerts, routes, pipeline, marker, timeoutFunc, dispMetrics)
		routes.Walk(func(r *dispatch.Route) {
			if r.RouteOpts.RepeatInterval > retention {
				configLogger.WithFields(logrus.Fields{
					"msg":             "repeat_interval is greater than the data retention period. It can lead to notifications being repeated more often than expected.",
					"repeat_interval": r.RouteOpts.RepeatInterval,
					"retention":       retention,
					"route":           r.Key(),
				}).Warn()
			}
		})

		go disp.Run()

		return nil
	})

	if err := configCoordinator.Reload(); err != nil {
		return 1
	}

	reloadChan := make(chan chan error)
	var (
		hup      = make(chan os.Signal, 1)
		hupReady = make(chan bool)
		term     = make(chan os.Signal, 1)
	)
	signal.Notify(hup, syscall.SIGHUP)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)

	go configWatcher(o.configFile, reloadChan)
	go func() {
		<-hupReady
		for {
			select {
			case <-hup:
				// ignore error, already logged in `reload()`
				_ = configCoordinator.Reload()
			case errc := <-reloadChan:
				errc <- configCoordinator.Reload()
			}
		}
	}()

	// Wait for reload or termination signals.
	close(hupReady) // Unblock SIGHUP handler.

	for {
		select {
		case <-term:
			logging.DefaultLogger.WithField("msg", "Received SIGTERM, exiting gracefully...").Info()
			return 0
		}
	}

}

func configWatcher(configfile string, reloadCh chan<- chan error) {
	logger := logging.DefaultLogger.WithField(logfields.LogSubsys, "configWactcher")
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		//handle err
	}
	defer watcher.Close()
	err = watcher.Add(configfile)
	if err != nil {
		//handle err
	}
	for {
		select {
		case event := <-watcher.Events:
			// k8s configmaps uses symlinks, we need this workaround.
			// original configmap file is removed
			if event.Op == fsnotify.Remove {
				// remove the watcher since the file is removed
				watcher.Remove(event.Name)
				// add a new watcher pointing to the new symlink/file
				watcher.Add(configfile)
				//reloadConfig()
			}
			// also allow normal files to be modified and reloaded.
			if event.Op&fsnotify.Write == fsnotify.Write {
				logger.Info("config file wacther write event..")
				errc := make(chan error)
				defer close(errc)
				reloadCh <- errc
				if err := <-errc; err != nil {
					logger.Error("failed to reload config: %s", err)
				}
			}
		case err := <-watcher.Errors:
			logger.Error("watcher.Errors: %s", err)
		}
	}
}

// buildReceiverIntegrations builds a list of integration notifiers off of a
// receiver config.
func buildReceiverIntegrations(nc *config.Receiver, tmpl *template.Template) ([]notify.Integration, error) {
	var (
		errs         types.MultiError
		integrations []notify.Integration
		add          = func(name string, i int, rs notify.ResolvedSender, f func() (notify.Notifier, error)) {
			n, err := f()
			if err != nil {
				errs.Add(err)
				return
			}
			integrations = append(integrations, notify.NewIntegration(n, rs, name, i))
		}
	)

	for i, c := range nc.WebhookConfigs {
		add("webhook", i, c, func() (notify.Notifier, error) { return webhook.New(c, tmpl) })
	}

	for i, c := range nc.YachConfigs {
		add("yach", i, c, func() (notify.Notifier, error) { return yach.New(c, tmpl) })
	}
	for i, c := range nc.WechatConfigs {
		add("wechat", i, c, func() (notify.Notifier, error) { return wechat.New(c, tmpl) })
	}
	if errs.Len() > 0 {
		return nil, &errs
	}
	return integrations, nil
}

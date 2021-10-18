package config

import (
	"crypto/md5"
	"encoding/binary"
	"github.com/crain-cn/event-mesh/pkg/logging"
	"github.com/crain-cn/event-mesh/pkg/logging/logfields"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"sync"
)

// Coordinator coordinates Alertmanager configurations beyond the lifetime of a
// single configuration.
type Coordinator struct {
	configFilePath string
	logger         *logrus.Entry
	// Protects config and subscribers
	mutex       sync.Mutex
	config      *Config
	subscribers []func(*Config) error

	configHashMetric        prometheus.Gauge
	configSuccessMetric     prometheus.Gauge
	configSuccessTimeMetric prometheus.Gauge
}

// NewCoordinator returns a new coordinator with the given configuration file
// path. It does not yet load the configuration from file. This is done in
// `Reload()`.
func NewCoordinator(configFilePath string, r prometheus.Registerer) *Coordinator {
	c := &Coordinator{
		configFilePath: configFilePath,
	}

	c.registerMetrics(r)
	c.logger = logging.DefaultLogger.WithField(logfields.LogSubsys, "configuration")

	return c
}

func (c *Coordinator) registerMetrics(r prometheus.Registerer) {
	configHash := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "alertmanager_config_hash",
		Help: "Hash of the currently loaded alertmanager configuration.",
	})
	configSuccess := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "alertmanager_config_last_reload_successful",
		Help: "Whether the last configuration reload attempt was successful.",
	})
	configSuccessTime := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "alertmanager_config_last_reload_success_timestamp_seconds",
		Help: "Timestamp of the last successful configuration reload.",
	})

	r.MustRegister(configHash, configSuccess, configSuccessTime)

	c.configHashMetric = configHash
	c.configSuccessMetric = configSuccess
	c.configSuccessTimeMetric = configSuccessTime
}

// Subscribe subscribes the given Subscribers to configuration changes.
func (c *Coordinator) Subscribe(ss ...func(*Config) error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.subscribers = append(c.subscribers, ss...)
}

func (c *Coordinator) notifySubscribers() error {
	for _, s := range c.subscribers {
		if err := s(c.config); err != nil {
			return err
		}
	}

	return nil
}

// loadFromFile triggers a configuration load, discarding the old configuration.
func (c *Coordinator) loadFromFile() error {
	conf, err := LoadFile(c.configFilePath)
	if err != nil {
		return err
	}

	c.config = conf

	return nil
}

// Reload triggers a configuration reload from file and notifies all
// configuration change subscribers.
func (c *Coordinator) Reload() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.logger.WithFields(logrus.Fields{
		"msg":  "Loading configuration file",
		"file": c.configFilePath,
	}).Info()

	if err := c.loadFromFile(); err != nil {
		c.logger.WithFields(logrus.Fields{
			"msg":  "Loading configuration file failed",
			"file": c.configFilePath,
		}).WithError(err).Error()
		c.configSuccessMetric.Set(0)
		return err
	}

	c.logger.WithFields(logrus.Fields{
		"msg":  "Completed loading of configuration file",
		"file": c.configFilePath,
	}).Info()

	if err := c.notifySubscribers(); err != nil {
		c.logger.WithFields(logrus.Fields{
			"msg":  "one or more config change subscribers failed to apply new config",
			"file": c.configFilePath,
		}).WithError(err).Error()
		c.configSuccessMetric.Set(0)
		return err
	}

	c.configSuccessMetric.Set(1)
	c.configSuccessTimeMetric.SetToCurrentTime()
	hash := md5HashAsMetricValue([]byte(c.config.original))
	c.configHashMetric.Set(hash)

	return nil
}

func (c *Coordinator) Log() *logrus.Entry {
	return c.logger
}

func md5HashAsMetricValue(data []byte) float64 {
	sum := md5.Sum(data)
	// We only want 48 bits as a float64 only has a 53 bit mantissa.
	smallSum := sum[0:6]
	var bytes = make([]byte, 8)
	copy(bytes, smallSum)
	return float64(binary.LittleEndian.Uint64(bytes))
}

// Copyright 2016 Prometheus Team
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mem

import (
	"context"
	"github.com/crain-cn/event-mesh/pkg/logging"
	"github.com/crain-cn/event-mesh/pkg/logging/logfields"
	"github.com/sirupsen/logrus"
	"sync"
	"time"

	"github.com/prometheus/common/model"

	"github.com/crain-cn/event-mesh/pkg/provider"
	"github.com/prometheus/alertmanager/store"
	"github.com/prometheus/alertmanager/types"
)

const alertChannelLength = 10000

// Alerts gives access to a set of alerts. All methods are goroutine-safe.
type Alerts struct {
	cancel context.CancelFunc

	mtx       sync.Mutex
	alerts    *store.Alerts
	listeners map[int]listeningAlerts
	next      int
	logger    *logrus.Entry
}

type listeningAlerts struct {
	alerts chan *types.Alert
	done   chan struct{}
}

// NewAlerts returns a new alert provider.
func NewAlerts(ctx context.Context, m types.Marker, intervalGC time.Duration) (*Alerts, error) {
	ctx, cancel := context.WithCancel(ctx)
	a := &Alerts{
		alerts:    store.NewAlerts(),
		cancel:    cancel,
		listeners: map[int]listeningAlerts{},
		next:      0,
		logger:    logging.DefaultLogger.WithField(logfields.LogSubsys, "provider"),
	}
	a.alerts.SetGCCallback(func(alerts []*types.Alert) {
		for _, alert := range alerts {
			// As we don't persist alerts, we no longer consider them after
			// they are resolved. Alerts waiting for resolved notifications are
			// held in memory in aggregation groups redundantly.
			m.Delete(alert.Fingerprint())
		}

		a.mtx.Lock()
		for i, l := range a.listeners {
			select {
			case <-l.done:
				delete(a.listeners, i)
				close(l.alerts)
			default:
				// listener is not closed yet, hence proceed.
			}
		}
		a.mtx.Unlock()
	})
	go a.alerts.Run(ctx, intervalGC)

	return a, nil
}

// Close the alert provider.
func (a *Alerts) Close() {
	if a.cancel != nil {
		a.cancel()
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Subscribe returns an iterator over active alerts that have not been
// resolved and successfully notified about.
// They are not guaranteed to be in chronological order.
func (a *Alerts) Subscribe() provider.AlertIterator {
	a.mtx.Lock()
	defer a.mtx.Unlock()

	var (
		done   = make(chan struct{})
		alerts = a.alerts.List()
		ch     = make(chan *types.Alert, max(len(alerts), alertChannelLength))
	)

	for _, a := range alerts {
		ch <- a
	}

	a.listeners[a.next] = listeningAlerts{alerts: ch, done: done}
	a.next++

	return provider.NewAlertIterator(ch, done, nil)
}

// GetPending returns an iterator over all the alerts that have
// pending notifications.
func (a *Alerts) GetPending() provider.AlertIterator {
	var (
		ch   = make(chan *types.Alert, alertChannelLength)
		done = make(chan struct{})
	)

	go func() {
		defer close(ch)

		for _, a := range a.alerts.List() {
			select {
			case ch <- a:
			case <-done:
				return
			}
		}
	}()

	return provider.NewAlertIterator(ch, done, nil)
}

// Get returns the alert for a given fingerprint.
func (a *Alerts) Get(fp model.Fingerprint) (*types.Alert, error) {
	return a.alerts.Get(fp)
}

func (a *Alerts) Listeners(alert *types.Alert)  {
	a.mtx.Lock()
	for _, l := range a.listeners {
		select {
		case l.alerts <- alert:
		case <-l.done:
		}
	}
	a.mtx.Unlock()
}


// Put adds the given alert to the set.
func (a *Alerts) Put(alerts ...*types.Alert) error {

	for _, alert := range alerts {
		fp := alert.Fingerprint()

		// Check that there's an alert existing within the store before
		// trying to merge.
		if old, err := a.alerts.Get(fp); err == nil {
			// Merge alerts if there is an overlap in activity range.
			if (alert.EndsAt.After(old.StartsAt) && alert.EndsAt.Before(old.EndsAt)) ||
				(alert.StartsAt.After(old.StartsAt) && alert.StartsAt.Before(old.EndsAt)) {
				alert = old.Merge(alert)
			}
		}

		if err := a.alerts.Set(alert); err != nil {
			a.logger.WithField("msg", "error on set alert").WithError(err)
			continue
		}

		go a.Listeners(alert)
	}

	a.Subscribe()
	return nil
}

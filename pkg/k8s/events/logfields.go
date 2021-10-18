package events

import (
	"github.com/crain-cn/event-mesh/pkg/logging"
	logfields "github.com/crain-cn/event-mesh/pkg/logging/logfields"
)

const (
	subsysEvent = "k8s-events"
)

var (
	log = logging.DefaultLogger.WithField(logfields.LogSubsys, subsysEvent)
)

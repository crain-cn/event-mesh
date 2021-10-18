package clustermesh

import (
	"github.com/crain-cn/event-mesh/pkg/logging"
	"github.com/crain-cn/event-mesh/pkg/logging/logfields"
)

// logging field definitions
const (
	subsysClusterMesh = "clustermesh"
)

var (
	log = logging.DefaultLogger.WithField(logfields.LogSubsys, subsysClusterMesh)
)

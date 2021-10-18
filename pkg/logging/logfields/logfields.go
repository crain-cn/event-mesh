package logfields

const (
	LogSubsys = "component"

	K8sAPIVersion = "k8sApiVersion"

	EventRoute = "eventRoute"

	EventRouteName = "eventRouteName"

	Object = "obj"

	// K8sNamespace is the namespace something belongs to
	K8sNamespace = "k8sNamespace"

	// StartTime is the start time of an event
	StartTime = "startTime"

	// EndTime is the end time of an event
	EndTime = "endTime"

	// Interval is the duration for periodic execution of an operation.
	Interval = "interval"

	// Duration is the duration of a measured operation
	Duration = "duration"

	// changing the value
	Path = "file-path"

	// Line is a line number within a file
	Line = "line"
)

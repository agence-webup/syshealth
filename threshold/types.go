package threshold

import (
	"webup/syshealth"
)

// Level represents a level of threshold
type Level int

// Label returns a label representing the level
func (l Level) Label() string {
	switch l {
	case Critical:
		return "Critical"
	case Warning:
		return "Warning"
	default:
		return ""
	}
}

// TriggerKey represents a key to identify a trigger
type TriggerKey string

const (
	// None represents that everything is OK
	None Level = 0
	// Warning represents the warning threshold
	Warning Level = 1
	// Critical represents the critical threshold
	Critical Level = 2
)

// MetricTrigger is a definition of a trigger for a specific metric key
type MetricTrigger interface {
	GetKey() TriggerKey
	Check(metrics syshealth.Data) Level
}

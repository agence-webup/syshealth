package threshold

import (
	"webup/syshealth"
)

// TriggerKey represents a key to identify a trigger
type TriggerKey string

// MetricTrigger is a definition of a trigger for a specific metric key
type MetricTrigger interface {
	GetKey() TriggerKey
	Check(metrics syshealth.Data) syshealth.ThresholdLevel
}

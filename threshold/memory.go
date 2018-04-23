package threshold

import "webup/syshealth"

type MemoryUsageTrigger struct {
}

func (trigger *MemoryUsageTrigger) GetKey() key {
	return "memory.usage"
}

func (trigger *MemoryUsageTrigger) Check(metrics syshealth.Data) syshealth.ThresholdLevel {
	if raw, ok := metrics["memory.available"]; ok {
		if available, ok := raw.(float64); ok {
			if available <= 0.3 {
				return syshealth.Critical
			}
			if available <= 0.5 {
				return syshealth.Warning
			}
		}
	}
	return syshealth.None
}

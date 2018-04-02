package threshold

import "webup/syshealth"

type DiskUsageTrigger struct {
}

func (trigger *DiskUsageTrigger) GetKey() TriggerKey {
	return "disk.usage"
}

func (trigger *DiskUsageTrigger) Check(metrics syshealth.Data) Level {
	if raw, ok := metrics["disk.usage"]; ok {
		if partitions, ok := raw.(map[string]map[string]float64); ok {
			if usageForDefaultPartition, ok := partitions["/"]; ok {
				if free, ok := usageForDefaultPartition["free"]; ok {
					if free <= 1.0 {
						return Critical
					}
					if free <= 2.0 {
						return Warning
					}
				}
			}
		}
	}
	return None
}

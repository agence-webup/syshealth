package threshold

import "webup/syshealth"

type CPULoadTrigger struct {
}

func (trigger *CPULoadTrigger) GetKey() TriggerKey {
	return "cpu.overload"
}

func (trigger *CPULoadTrigger) Check(metrics syshealth.Data) Level {
	if rawLoad, ok := metrics["cpu.load_5"]; ok {
		if load, ok := rawLoad.(float64); ok {
			if load >= 0.8 {
				return Critical
			}
			if load >= 0.05 {
				return Warning
			}
		}
	}
	return None
}

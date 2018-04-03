package threshold

import (
	"log"
	"time"
	"webup/syshealth"
	"webup/syshealth/alert"
)

type manager struct {
	triggers       []MetricTrigger
	stateByTrigger map[TriggerKey]triggerState
}

type triggerState struct {
	LastChange time.Time
	Count      int
	Level      syshealth.ThresholdLevel
}

var man *manager

func StartWatching() (receivedData chan syshealth.Data) {

	man = new(manager)

	// enable triggers
	man.triggers = []MetricTrigger{
		new(CPULoadTrigger),
		new(MemoryUsageTrigger),
		new(DiskUsageTrigger),
	}

	// prepare state storage
	man.stateByTrigger = map[TriggerKey]triggerState{}
	for _, t := range man.triggers {
		man.stateByTrigger[t.GetKey()] = triggerState{}
	}

	receivedData = make(chan syshealth.Data)

	go func() {
		for {
			select {
			case data := <-receivedData:
				for _, t := range man.triggers {
					result := t.Check(data)

					// get current state
					state := man.stateByTrigger[t.GetKey()]

					// detect a change
					if state.Level == syshealth.None && result > syshealth.None || state.Level > syshealth.None && result == syshealth.None {
						state.LastChange = time.Now()
						log.Printf("%v: change detected\n", t.GetKey())
					}

					// update the level
					state.Level = result

					// check if the trigger must be activated
					if state.Level > syshealth.None && time.Now().Sub(state.LastChange) >= time.Duration(10)*time.Second {

						// send alert
						err := alert.SendSlackAlert(syshealth.Alert{
							IssueTitle: string(t.GetKey()),
							Server:     syshealth.Server{Name: "test", IP: "0.0.0.0"},
							Level:      state.Level,
						})
						if err != nil {
							log.Println("cannot send alert:", err)
						}
						state.LastChange = time.Now()
					}

					man.stateByTrigger[t.GetKey()] = state
				}
			}
		}
	}()

	return
}

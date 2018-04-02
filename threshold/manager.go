package threshold

import (
	"fmt"
	"log"
	"time"
	"webup/syshealth"
)

type manager struct {
	triggers       []MetricTrigger
	stateByTrigger map[TriggerKey]triggerState
}

type triggerState struct {
	LastChange time.Time
	Count      int
	Level      Level
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
					if state.Level == None && result > None || state.Level > None && result == None {
						state.LastChange = time.Now()
						log.Println("trigger.cpu: change detected", result)
					}

					// update the level
					state.Level = result

					// check if the trigger must be activated
					if state.Level > None && time.Now().Sub(state.LastChange) >= time.Duration(2)*time.Minute {
						fmt.Println("TRIGGER:", state.Level)
						state.LastChange = time.Now()
					} else {
						log.Println("trigger.cpu: no trigger needed")
						log.Println((time.Now().Sub(state.LastChange)).Seconds(), "sec remaining before triggering")
					}

					man.stateByTrigger[t.GetKey()] = state
				}
			}
		}
	}()

	return
}

package threshold

import (
	"log"
	"time"
	"webup/syshealth"
	"webup/syshealth/alert"
)

type key string

type watcher struct {
	triggers       []trigger
	stateByTrigger map[key]triggerState
}

type triggerState struct {
	LastChange    time.Time
	LastSentAlert time.Time
	AlertCount    int
	Level         syshealth.ThresholdLevel
}

func (state *triggerState) reset() {
	state.LastChange = time.Now()
	state.LastSentAlert = time.Time{}
	state.AlertCount = 0
}

// trigger is a definition of a trigger for a specific metric key
type trigger interface {
	GetKey() key
	Check(metrics syshealth.Data) syshealth.ThresholdLevel
}

const maxCountForAlerts = 3

// NewWatcher returns a watcher for metrics threshold
func NewWatcher() syshealth.Watcher {
	w := watcher{
		triggers: []trigger{
			new(CPULoadTrigger),
			new(MemoryUsageTrigger),
			new(DiskUsageTrigger),
		},
	}

	// prepare state storage
	w.stateByTrigger = map[key]triggerState{}
	for _, t := range w.triggers {
		w.stateByTrigger[t.GetKey()] = triggerState{}
	}

	return &w
}

func (w *watcher) GetKey() syshealth.WatcherKey {
	return "metrics_threshold"
}

func (w *watcher) Watch(data syshealth.WatcherData) {

	for _, t := range w.triggers {
		result := t.Check(data.Metrics)

		// get current state
		state := w.stateByTrigger[t.GetKey()]

		// detect a change
		if state.Level == syshealth.None && result > syshealth.None || state.Level > syshealth.None && result == syshealth.None {
			state.reset()
			log.Printf("%v(%v): change detected\n", t.GetKey(), data.Server.Name)
		}

		// update the level
		state.Level = result

		// check if the trigger must be activated
		// - the level must be greater than 'None'
		// - the level must not have changed for 2 minutes
		if state.Level > syshealth.None && time.Now().Sub(state.LastChange) >= time.Duration(2)*time.Minute {

			// current alert count
			count := state.AlertCount
			if count > maxCountForAlerts {
				count = maxCountForAlerts
			}

			log.Printf("%v(%v): trigger activated\n", t.GetKey(), data.Server.Name)

			timeSinceLastAlert := time.Now().Sub(state.LastSentAlert)
			if timeSinceLastAlert >= time.Duration(count*10)*time.Minute {
				// send alert
				err := alert.SendSlackAlert(syshealth.Alert{
					IssueTitle: string(t.GetKey()),
					Server:     data.Server,
					Level:      state.Level,
				})
				if err != nil {
					log.Println("cannot send alert:", err)
				}

				state.AlertCount++
				state.LastSentAlert = time.Now()

			} else {
				nextAlertIn := time.Duration(count*10)*time.Minute - timeSinceLastAlert
				log.Printf("%v(%v): no alert sent (next alert in %v)\n", t.GetKey(), data.Server.Name, nextAlertIn.String())
			}

			state.LastChange = time.Now()
		}

		w.stateByTrigger[t.GetKey()] = state
	}
}

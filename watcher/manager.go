package watcher

import (
	"webup/syshealth"
	"webup/syshealth/threshold"
)

type manager struct {
	watchers []syshealth.Watcher
}

var man *manager

// Start launches the routine responsible to start and handle watchers
func Start() (receivedData chan syshealth.WatcherData) {

	man = new(manager)

	// enable triggers
	man.watchers = []syshealth.Watcher{
		threshold.NewWatcher(),
	}

	receivedData = make(chan syshealth.WatcherData)

	go func() {
		for {
			select {
			case data := <-receivedData:
				for _, w := range man.watchers {
					// execute the `Watch` function in a dedicated thread
					go w.Watch(data)
				}
			}
		}
	}()

	return
}

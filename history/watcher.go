package history

import (
	"time"
	"webup/syshealth"
)

type serverID string

type watcher struct {
	aggregatorsByServer map[serverID]serverAggregator
	fetcher             DataFetcher
}

type DataFetcher func(serverId string) map[string][]Data

type Data struct {
	Date  time.Time   `json:"t"`
	Value interface{} `json:"y"`
}

type serverAggregator struct {
	Aggregators map[string]aggregator
	Data        map[string][]Data
}

func newServerAggregator() serverAggregator {
	s := serverAggregator{}
	s.Aggregators = map[string]aggregator{
		"cpu.usage":           new(cpuUsageAggregator),
		"memory.used_percent": new(memoryUsageAggregator),
	}
	s.Data = map[string][]Data{
		"cpu.usage":           []Data{},
		"memory.used_percent": []Data{},
	}
	return s
}

type aggregator interface {
	AddValue(value interface{})
	GetAverageValue() interface{}
}

// NewWatcher returns a watcher responsible to store history for each metric
func NewWatcher() (syshealth.Watcher, DataFetcher) {
	w := watcher{}

	// init maps
	w.aggregatorsByServer = map[serverID]serverAggregator{}
	// init fetcher
	w.fetcher = func(id string) map[string][]Data {
		return w.GetServerHistory(id)
	}

	go func() {
		// 1h of data
		maxValues := int(time.Hour.Minutes()) + 1

		ticker := time.Tick(time.Duration(1) * time.Minute)
		for {
			select {
			case t := <-ticker:
				for server, sg := range w.aggregatorsByServer {
					for k, agg := range sg.Aggregators {
						// add data for the aggregated value on the server
						w.aggregatorsByServer[server].Data[k] = append(w.aggregatorsByServer[server].Data[k], Data{
							Date:  t,
							Value: agg.GetAverageValue(),
						})

						// remove the first value if needed
						if len(w.aggregatorsByServer[server].Data[k]) == maxValues {
							w.aggregatorsByServer[server].Data[k] = append(w.aggregatorsByServer[server].Data[k][:0], w.aggregatorsByServer[server].Data[k][1:]...)
						}
					}
				}

				// fmt.Printf("%+v\n\n", w.aggregatorsByServer)
			}
		}
	}()

	return &w, w.fetcher
}

func (w *watcher) GetServerHistory(id string) map[string][]Data {
	if data, ok := w.aggregatorsByServer[serverID(id)]; ok {
		return data.Data
	}
	return nil
}

func (w *watcher) GetKey() syshealth.WatcherKey {
	return "history"
}

func (w *watcher) Watch(data syshealth.WatcherData) {
	id := serverID(data.Server.ID)

	// init server aggregator if needed
	if _, ok := w.aggregatorsByServer[id]; !ok {
		w.aggregatorsByServer[id] = newServerAggregator()
	}

	historyMetrics := []string{"cpu.usage", "memory.used_percent"}

	for _, metric := range historyMetrics {
		if val, ok := data.Metrics[metric]; ok {
			w.aggregatorsByServer[id].Aggregators[metric].AddValue(val)
		}
	}
}

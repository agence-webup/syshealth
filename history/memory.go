package history

import "math"

type memoryUsageAggregator struct {
	AggregatedValue float64
	Count           float64
}

func (agg *memoryUsageAggregator) AddValue(value interface{}) {
	if usage, ok := value.(float64); ok {
		agg.AggregatedValue += usage
		agg.Count++
	}
}

func (agg *memoryUsageAggregator) GetAverageValue() interface{} {
	avg := agg.AggregatedValue / agg.Count

	// reset
	agg.AggregatedValue = 0
	agg.Count = 0

	return math.Round(avg/0.01) * 0.01
}

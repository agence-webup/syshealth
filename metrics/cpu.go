package metrics

import (
	"time"
	"webup/syshealth"

	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/load"
)

const (
	coreCountKey = "cpu.core_count"
	usageKey     = "cpu.usage"
	load1Key     = "cpu.load_1"
	load5Key     = "cpu.load_5"
	load15Key    = "cpu.load_15"
)

type name struct {
}

func GetCPU() (syshealth.Data, error) {

	data := syshealth.Data{}

	// cpu count
	cpuCount, err := cpu.Counts(false)
	if err != nil {
		cpuCount = 1
	}

	data[coreCountKey] = cpuCount

	// percent
	percent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return data, errors.Wrap(err, "cannot get CPU percent")
	}
	data[usageKey] = percent[0]

	// load
	l, err := load.Avg()
	if err != nil {
		return data, errors.Wrap(err, "cannot get CPU load")
	}

	load1 := l.Load1 / float64(cpuCount)
	load5 := l.Load5 / float64(cpuCount)
	load15 := l.Load15 / float64(cpuCount)

	data[load1Key] = load1
	data[load5Key] = load5
	data[load15Key] = load15

	return data, nil
}

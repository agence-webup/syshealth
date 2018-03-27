package metrics

import (
	"time"
	"webup/syshealth"

	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/load"
)

func GetCPU() (syshealth.Data, error) {

	data := syshealth.Data{}

	// cpu count
	cpuCount, err := cpu.Counts(false)
	if err != nil {
		cpuCount = 1
	}

	data["cpu.core_count"] = cpuCount

	// percent
	percent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return data, errors.Wrap(err, "cannot get CPU percent")
	}
	data["cpu.usage"] = percent[0]

	// load
	l, err := load.Avg()
	if err != nil {
		return data, errors.Wrap(err, "cannot get CPU load")
	}

	load1 := l.Load1 / float64(cpuCount)
	load5 := l.Load5 / float64(cpuCount)
	load15 := l.Load15 / float64(cpuCount)

	data["cpu.load_1"] = load1
	data["cpu.load_5"] = load5
	data["cpu.load_15"] = load15

	return data, nil
}

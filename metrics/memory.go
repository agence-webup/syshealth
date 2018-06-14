package metrics

import (
	"webup/syshealth"

	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/mem"
)

func GetMemory() (syshealth.Data, error) {

	data := syshealth.Data{}

	v, err := mem.VirtualMemory()
	if err != nil {
		return data, errors.Wrap(err, "cannot get virtual memory")
	}

	data["memory.available"] = float64(v.Available) / 1024.0 / 1024.0 / 1024.0 // from bytes to gigabytes
	data["memory.used_percent"] = v.UsedPercent

	return data, nil
}

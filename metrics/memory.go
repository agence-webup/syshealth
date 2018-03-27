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

	data["memory.used_percent"] = v.UsedPercent

	return data, nil
}

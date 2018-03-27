package metrics

import (
	"webup/syshealth"

	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/disk"
)

func GetDisk() (syshealth.Data, error) {

	data := syshealth.Data{}

	// disk
	d, err := disk.Partitions(false)
	if err != nil {
		return data, errors.Wrap(err, "cannot get partitions")
	}

	partitions := map[string]interface{}{}

	for _, info := range d {
		u, err := disk.Usage(info.Mountpoint)
		if err != nil {
			return data, errors.Wrap(err, "cannot get partitions")
		}

		total := ((float64(u.Total) / 1024) / 1024) / 1024
		free := ((float64(u.Free) / 1024) / 1024) / 1024

		partitions[info.Mountpoint] = map[string]interface{}{
			"total":   total,
			"free":    free,
			"percent": u.UsedPercent,
		}
	}

	data["disk.usage"] = partitions

	return data, nil
}

package main

import (
	"fmt"

	"os"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
)

const (
	memoryCritical = 85  // %
	memoryWarning  = 75  // %
	loadCritical   = 2.0 // load average value
	loadWarning    = 1.0 // load average value
	diskCritical   = 3   // Gb remaining
	diskWarning    = 5
)

type status struct {
	Warning  bool
	Critical bool
}

func main() {
	fmt.Println("### Syshealth 1 (build 3) ###")

	currentStatus := status{Warning: false, Critical: false}

	// memory
	fmt.Println("Memory:")
	v, err := mem.VirtualMemory()
	if err != nil {
		fmt.Printf("   Unable to get memory info: %v", err)
		currentStatus.Critical = true
	} else {
		fmt.Printf("   Total: %v MB, Free: %v MB, Used: %f%%\n", (v.Total/1024)/1024, (v.Available/1024)/1024, v.UsedPercent)

		if v.UsedPercent >= memoryCritical {
			currentStatus.Critical = true
		} else if v.UsedPercent >= memoryWarning {
			currentStatus.Warning = true
		}
	}

	// cpu count
	cpuCount, err := cpu.Counts(false)
	if err != nil {
		cpuCount = 1
	}

	// load
	fmt.Printf("Load (%v cores):\n", cpuCount)
	l, err := load.Avg()
	if err != nil {
		fmt.Printf("   Unable to get load info: %v", err)
		currentStatus.Critical = true
	} else {
		load1 := l.Load1 / float64(cpuCount)
		load5 := l.Load5 / float64(cpuCount)
		load15 := l.Load15 / float64(cpuCount)
		fmt.Printf("   1: %f, 5: %f, 15: %f\n", load1, load5, load15)

		if load15 >= loadCritical {
			currentStatus.Critical = true
		} else if load15 >= loadWarning {
			currentStatus.Warning = true
		}
	}

	// disk
	fmt.Println("Disk:")
	d, err := disk.Partitions(false)
	if err != nil {
		fmt.Printf("   Unable to get partitions: %v", err)
		currentStatus.Critical = true
	} else {
		for _, info := range d {
			u, err := disk.Usage(info.Mountpoint)
			if err != nil {
				fmt.Printf("   Unable to get disk usage info: %v", err)
				currentStatus.Critical = true
			} else {
				total := ((float64(u.Total) / 1024) / 1024) / 1024
				free := ((float64(u.Free) / 1024) / 1024) / 1024
				fmt.Printf("   %v -> Total: %v GB, Free: %v GB, Used: %f%%\n", info.Mountpoint, total, free, u.UsedPercent)

				if free <= diskCritical {
					currentStatus.Critical = true
				} else if free <= diskWarning {
					currentStatus.Warning = true
				}
			}
		}
	}

	if currentStatus.Critical {
		os.Exit(2)
	} else if currentStatus.Warning {
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}

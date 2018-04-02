package alerts

import (
	"webup/syshealth"
	"webup/syshealth/threshold"
)

type Alert struct {
	IssueTitle string
	Server     syshealth.Server
	Level      threshold.Level
}

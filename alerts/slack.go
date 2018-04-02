package alerts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
	"webup/syshealth/threshold"

	"github.com/pkg/errors"
)

/*
{
    "attachments": [
        {
            "fallback": "Required plain-text summary of the attachment.",
            "color": "warning",
            "title": "CPU/Load issue",
            "fields": [
                {
                    "title": "Server",
                    "value": "q-demos2",
                    "short": true
                },
				{
                    "title": "IP",
                    "value": "23.53.154.12",
                    "short": true
                },
				{
                    "title": "Level",
                    "value": "Warning",
                    "short": true
                }
            ],
            "ts": 123456789
        }
    ]
}
*/

type slackPayload struct {
	Attachments []slackPayloadAttachment `json:"attachments"`
}

type slackPayloadAttachment struct {
	Fallback string                        `json:"fallback"`
	Color    string                        `json:"color"`
	Title    string                        `json:"title"`
	Fields   []slackPayloadAttachmentField `json:"fields"`
}

type slackPayloadAttachmentField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

var webhookURL string

func InitSlackAlerter(URL string) {
	webhookURL = URL
}

func SendSlackAlert(alert Alert) error {

	// check if alerter is initialized
	if webhookURL == "" {
		return errors.New("webhook URL is not initialized")
	}

	// prepare payload
	payload := getPayload(alert)
	data, err := json.Marshal(payload)
	if err != nil {
		return errors.Wrap(err, "cannot marshal payload into json")
	}

	// prepare the request
	b := bytes.NewBuffer(data)
	req, err := http.NewRequest("POST", webhookURL, b)
	if err != nil {
		return errors.Wrap(err, "cannot prepare request")
	}

	client := http.Client{
		Timeout: time.Duration(5) * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, "error with request to slack API")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "cannot read response body")
	}
	fmt.Println(string(body))

	return nil
}

func getPayload(alert Alert) slackPayload {
	return slackPayload{
		Attachments: []slackPayloadAttachment{
			slackPayloadAttachment{
				Title:    alert.IssueTitle,
				Color:    getSlackColorForLevel(alert.Level),
				Fallback: fmt.Sprintf("%v: %s on '%v' (%v)", alert.Level.Label(), alert.IssueTitle, alert.Server.Name, alert.Server.IP),
				Fields: []slackPayloadAttachmentField{
					slackPayloadAttachmentField{
						Title: "Server",
						Value: alert.Server.Name,
						Short: true,
					},
					slackPayloadAttachmentField{
						Title: "IP",
						Value: alert.Server.IP,
						Short: true,
					},
					slackPayloadAttachmentField{
						Title: "Level",
						Value: alert.Level.Label(),
						Short: true,
					},
				},
			},
		},
	}
}

func getSlackColorForLevel(level threshold.Level) string {
	switch level {
	case threshold.Critical:
		return "danger"
	case threshold.Warning:
		return "warning"
	default:
		return "good"
	}
}

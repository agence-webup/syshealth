package http

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"webup/syshealth"
	"webup/syshealth/metrics"

	"github.com/pkg/errors"
)

// SendData is responsible to fetch server metrics and execute
// a HTTP request to send collected data
func SendData(serverURL string, jwt string) error {

	data := syshealth.Data{}

	// fetch data
	cpu, err := metrics.GetCPU()
	if err != nil {
		return errors.Wrap(err, "cannot get cpu metrics data")
	}
	for k, v := range cpu {
		data[k] = v
	}

	memory, err := metrics.GetMemory()
	if err != nil {
		return errors.Wrap(err, "cannot get memory metrics data")
	}
	for k, v := range memory {
		data[k] = v
	}

	disk, err := metrics.GetDisk()
	if err != nil {
		return errors.Wrap(err, "cannot get disk metrics data")
	}
	for k, v := range disk {
		data[k] = v
	}

	jsonData := syshealth.MetricBag{Metrics: data}

	// data
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(jsonData)

	client := &http.Client{}
	req, err := http.NewRequest("POST", serverURL+"/api/metrics", b)
	req.Header.Add("Authorization", "Bearer "+jwt)
	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, "unable to send metrics")
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		b, _ := ioutil.ReadAll(resp.Body)
		return errors.Wrap(errors.New(string(b)), "API responded with error")
	}

	return nil
}

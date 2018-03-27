package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"webup/syshealth"
	"webup/syshealth/metrics"

	"github.com/pkg/errors"
)

// SendData is responsible to fetch server metrics and execute
// a HTTP request to send collected data
func SendData() error {

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

	// jwt
	jwt := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJqdGkiOiJjMGY4MjJlYmQxOGY0MWI3MTIyMzQ3YmU0OWZhYjdmMDk2ODg4MDk4IiwiaXNzIjoic3lzaGVhbHRoLXNlcnZlciJ9.LjiHbdIL7LYx-0BMOdf1qItaxErHZHPHz5rrsrotRXE"

	client := &http.Client{}
	req, err := http.NewRequest("POST", "http://localhost:1323/api/metrics", b)
	req.Header.Add("Authorization", "Bearer "+jwt)
	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, "unable to send metrics")
	}
	defer resp.Body.Close()

	fmt.Printf("response (%d) :\n", resp.StatusCode)
	io.Copy(os.Stdout, resp.Body)
	fmt.Println("")

	return nil
}

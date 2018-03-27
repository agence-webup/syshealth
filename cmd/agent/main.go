package main

import (
	"fmt"
	"log"
	"os/signal"
	"syscall"
	"time"
	"webup/syshealth/http"

	"os"

	"github.com/jawher/mow.cli"
)

func main() {

	app := cli.App("syshealth-agent", "Syshealth agent gathering server metrics")

	app.Version("v version", "syshealth-agent v1.0 (build 1)")
	app.Spec = "[--polling-rate]"

	var (
		pollingRate = app.IntOpt("polling-rate", 5, "Polling rate for gathering metrics (in seconds)")
	)

	app.Action = func() {

		sigs := make(chan os.Signal, 1)
		done := make(chan bool, 1)

		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

		go func() {
			ticker := time.NewTicker(time.Duration(*pollingRate) * time.Second)

			for {
				select {
				case <-ticker.C:
					fmt.Println("fetch metrics...")
					err := http.SendData()
					if err != nil {
						log.Println("error sending data:", err)
					}
				case <-sigs:
					ticker.Stop()
					done <- true
				}
			}
		}()

		fmt.Println("awaiting signal")
		<-done
		fmt.Println("")
		fmt.Println("exiting")
	}

	app.Run(os.Args)

}

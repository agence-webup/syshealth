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

	app := cli.App("syshealth-agent", "Syshealth agent gathering host metrics (requires syshealth-server v1.x")

	app.Version("v version", "syshealth-agent v2.0 (build 2)")

	app.Spec = "--jwt --server-url [--polling-rate]"

	var (
		jwt = app.String(cli.StringOpt{
			Name:   "jwt",
			Desc:   "JWT token given by the server after server registration",
			Value:  "",
			EnvVar: "SYSHEALTH_AGENT_JWT",
		})
		serverURL = app.String(cli.StringOpt{
			Name:   "server-url",
			Desc:   "Server URL",
			Value:  "https://syshealth.io",
			EnvVar: "SYSHEALTH_AGENT_SERVER_URL",
		})
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
					err := http.SendData(*serverURL, *jwt)
					if err != nil {
						log.Println("error sending data:", err)
					}
				case <-sigs:
					ticker.Stop()
					done <- true
				}
			}
		}()

		fmt.Println("ready.")
		<-done
		fmt.Println("")
		fmt.Println("exiting")
	}

	app.Run(os.Args)

}

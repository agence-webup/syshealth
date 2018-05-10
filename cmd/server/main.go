//go:generate statik -src=./ui

package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
	"webup/syshealth"
	"webup/syshealth/alert"
	"webup/syshealth/history"
	"webup/syshealth/repository/bolt"
	"webup/syshealth/repository/memory"
	"webup/syshealth/threshold"
	"webup/syshealth/watcher"

	_ "webup/syshealth/cmd/server/statik"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/jawher/mow.cli"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/pkg/errors"
	"github.com/rakyll/statik/fs"
)

func main() {

	app := cli.App("syshealth-server", "Syshealth server with API gathering server metrics sent by agents and an UI to display collected data")

	app.Version("v version", "syshealth-server v1 (build 1)")

	app.Command("daemon", "Start the API listening for metrics and serving the UI", func(cmd *cli.Cmd) {

		cmd.Spec = "[--listening-ip] [--listening-port] [--agent-jwt-secret] [--client-jwt-secret] [--slack-webhook-url] [--database-directory]"

		listeningIP := cmd.String(cli.StringOpt{
			Name:   "listening-ip",
			Value:  "0.0.0.0",
			Desc:   "Listening IP of the server",
			EnvVar: "SYSHEALTH_LISTEN_IP",
		})
		listeningPort := cmd.String(cli.StringOpt{
			Name:   "listening-port",
			Value:  "1323",
			Desc:   "Listening port of the server",
			EnvVar: "SYSHEALTH_LISTEN_PORT",
		})
		listeningPrivatePort := cmd.String(cli.StringOpt{
			Name:   "listening-private-port",
			Value:  "1324",
			Desc:   "Listening port of the private API",
			EnvVar: "SYSHEALTH_LISTEN_PRIVATE_PORT",
		})
		agentJwtSecret := cmd.String(cli.StringOpt{
			Name:   "agent-jwt-secret",
			Value:  "nnqs10sn#éQ$*svn2q",
			Desc:   "JWT secret for agents authentication",
			EnvVar: "SYSHEALTH_AGENT_JWT_SECRET",
		})
		clientJwtSecret := cmd.String(cli.StringOpt{
			Name:   "client-jwt-secret",
			Value:  "yoQaoQN3&Dq*nOdqn1§",
			Desc:   "JWT secret for clients authentication",
			EnvVar: "SYSHEALTH_CLIENT_JWT_SECRET",
		})
		slackWebhookURL := cmd.String(cli.StringOpt{
			Name:   "slack-webhook-url",
			Value:  "",
			Desc:   "Slack webhook URL for sending alerts",
			EnvVar: "SYSHEALTH_SLACK_WEBHOOK_URL",
		})
		databaseDirectory := cmd.String(cli.StringOpt{
			Name:   "database-directory",
			Value:  ".",
			Desc:   "Directory path where the database file will be stored",
			EnvVar: "SYSHEALTH_DATABASE_DIRECTORY",
		})

		cmd.Action = func() {

			adminUserRepo := bolt.GetAdminUserRepository(*databaseDirectory)
			serverRepo := bolt.GetServerRepository(*databaseDirectory)
			metricRepo := memory.GetMetricRepository()

			alert.InitSlackAlerter(*slackWebhookURL)

			// prepare watchers
			historyWatcher, historyFetcher := history.NewWatcher()
			watchers := []syshealth.Watcher{
				threshold.NewWatcher(),
				historyWatcher,
			}
			// start the watcher process
			receivedDataForWatchers := watcher.Start(watchers)

			// setup
			authEnabled, err := adminUserRepo.IsSetup()
			if err != nil {
				log.Fatalln(errors.Wrap(err, "unable to check if auth is setup"))
				return
			}
			if !authEnabled {
				err := adminUserRepo.Create("admin", "admin")
				if err != nil {
					log.Fatalln(errors.Wrap(err, "unable to create the default 'admin' user"))
					return
				}
			}

			e := echo.New()

			// UI

			statikFS, err := fs.New()
			if err != nil {
				log.Fatal(err)
			}
			e.GET("/ui/*", echo.WrapHandler(http.StripPrefix("/ui/", http.FileServer(statikFS))))

			// API

			e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
				AllowOrigins: []string{"*"},
			}))

			agentJwtMiddleware := middleware.JWTWithConfig(middleware.JWTConfig{
				SigningKey: []byte(*agentJwtSecret),
			})
			clientJwtMiddleware := middleware.JWTWithConfig(middleware.JWTConfig{
				SigningKey: []byte(*clientJwtSecret),
			})

			// endpoint used by agents to send their metrics
			e.POST("/api/metrics", func(c echo.Context) error {

				token := c.Get("user").(*jwt.Token)
				claims := token.Claims.(jwt.MapClaims)
				id := claims["jti"].(string)

				// check if server is registered
				server, err := serverRepo.GetServer(id)
				if err != nil {
					return echo.NewHTTPError(http.StatusBadRequest, errors.Wrap(err, "unable to check if server is registered"))
				}
				if server == nil {
					return c.NoContent(http.StatusUnauthorized)
				}

				// parse data
				data := syshealth.MetricBag{}
				err = c.Bind(&data)
				if err != nil {
					return echo.NewHTTPError(http.StatusBadRequest, errors.Wrap(err, "unable to parse json body"))
				}

				// store data
				err = metricRepo.Store(id, data.Metrics)
				if err != nil {
					return echo.NewHTTPError(http.StatusBadRequest, errors.Wrap(err, "unable to store metric"))
				}

				// send data for triggers
				receivedDataForWatchers <- syshealth.WatcherData{Server: *server, Metrics: data.Metrics}

				return c.NoContent(http.StatusOK)

			}, agentJwtMiddleware)

			// authentication endpoint for API clients (not agents)
			e.POST("/api/login", func(c echo.Context) error {

				data := struct {
					Username string `json:"username"`
					Password string `json:"password"`
				}{}

				err := c.Bind(&data)
				if err != nil {
					return echo.NewHTTPError(http.StatusBadRequest, errors.Wrap(err, "unable to parse json body"))
				}

				logged, err := adminUserRepo.Login(data.Username, data.Password)
				if err != nil {
					return echo.NewHTTPError(http.StatusInternalServerError, errors.Wrap(err, "unable to login user"))
				}

				if !logged {
					return echo.NewHTTPError(http.StatusUnauthorized, "unable to login. check credentials")
				}

				// Set claims
				claims := jwt.StandardClaims{
					Issuer:    "syshealth-server",
					IssuedAt:  time.Now().Unix(),
					ExpiresAt: time.Now().Add(time.Hour * 72).Unix(),
				}

				// Create token
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

				// Generate encoded token and send it as response.
				t, err := token.SignedString([]byte(*clientJwtSecret))
				if err != nil {
					return err
				}

				return c.JSON(http.StatusOK, map[string]string{
					"jwt": t,
				})
			})

			e.GET("/api/users", func(c echo.Context) error {

				users, err := adminUserRepo.GetUsers()
				if err != nil {
					return echo.NewHTTPError(http.StatusInternalServerError, errors.Wrap(err, "unable to fetch admin users"))
				}

				data := map[string]interface{}{
					"users": users,
				}

				return c.JSON(http.StatusOK, data)
			}, clientJwtMiddleware)

			e.POST("/api/users", func(c echo.Context) error {

				data := struct {
					Username string `json:"username"`
					Password string `json:"password"`
				}{}

				err = c.Bind(&data)
				if err != nil {
					return echo.NewHTTPError(http.StatusBadRequest, errors.Wrap(err, "unable to parse json body"))
				}

				err = adminUserRepo.Create(data.Username, data.Password)
				if err != nil {
					return echo.NewHTTPError(http.StatusInternalServerError, errors.Wrap(err, "unable to create admin user"))
				}

				return c.NoContent(http.StatusNoContent)
			}, clientJwtMiddleware)

			e.DELETE("/api/users/:username", func(c echo.Context) error {

				username := c.Param("username")
				if username == "" {
					return echo.NewHTTPError(http.StatusBadRequest, "a username must be provided")
				}

				err := adminUserRepo.Delete(username)
				if err != nil {
					return echo.NewHTTPError(http.StatusInternalServerError, errors.Wrap(err, "unable to delete the admin user"))
				}

				return c.NoContent(http.StatusOK)
			}, clientJwtMiddleware)

			e.GET("/api/metrics", func(c echo.Context) error {

				servers, err := serverRepo.GetServers()
				if err != nil {
					return echo.NewHTTPError(http.StatusInternalServerError, errors.Wrap(err, "unable to get servers"))
				}

				type serverData struct {
					syshealth.Server
					DefaultPartition string `json:"default_partition"`
				}

				type metric struct {
					Server serverData      `json:"server"`
					Data   *syshealth.Data `json:"data"`
				}

				metrics := []metric{}
				for _, server := range servers {
					data, err := metricRepo.Get(server.ID)
					if err != nil {
						log.Println(errors.Wrap(err, "unable to get data for registered server"))
					}

					metrics = append(metrics, metric{
						Server: serverData{Server: server, DefaultPartition: "/"},
						Data:   data,
					})
				}

				return c.JSON(http.StatusOK, metrics)
			}, clientJwtMiddleware)

			e.GET("/api/metrics/:id", func(c echo.Context) error {

				id := c.Param("id")

				data := historyFetcher(id)

				type point struct {
					Time string      `json:"t"`
					Val  interface{} `json:"y"`
				}
				jsonData := map[string][]point{}
				for key, data := range data {
					points := []point{}
					for _, v := range data {
						points = append(points, point{
							Time: v.Date.Format(time.RFC3339),
							Val:  v.Value,
						})
					}
					jsonData[key] = points
				}

				return c.JSON(http.StatusOK, jsonData)
			}, clientJwtMiddleware)

			e.GET("/api/servers", func(c echo.Context) error {

				servers, err := serverRepo.GetServers()
				if err != nil {
					return echo.NewHTTPError(http.StatusInternalServerError, errors.Wrap(err, "unable to fetch servers"))
				}

				data := map[string]interface{}{
					"servers": servers,
				}

				return c.JSON(http.StatusOK, data)
			}, clientJwtMiddleware)

			e.POST("/api/servers/register", func(c echo.Context) error {

				data := syshealth.Server{}

				err := c.Bind(&data)
				if err != nil {
					return echo.NewHTTPError(http.StatusBadRequest, errors.Wrap(err, "unable to get json data"))
				}

				if data.Name == "" || data.IP == "" {
					return echo.NewHTTPError(http.StatusBadRequest, errors.Wrap(err, "data must contain 'name' and 'ip' fields"))
				}

				jwt, err := serverRepo.RegisterServer(data, *agentJwtSecret)
				if err != nil {
					return echo.NewHTTPError(http.StatusInternalServerError, errors.Wrap(err, "unable to register server"))
				}

				json := map[string]interface{}{
					"jwt":    jwt,
					"server": data,
				}

				return c.JSON(http.StatusOK, json)
			}, clientJwtMiddleware)

			e.DELETE("/api/servers/:id", func(c echo.Context) error {

				id := c.Param("id")

				err := serverRepo.RevokeServer(id)
				if err != nil {
					return echo.NewHTTPError(http.StatusInternalServerError, errors.Wrap(err, "unable to revoke server"))
				}

				return c.NoContent(http.StatusOK)
			}, clientJwtMiddleware)

			// private API for support tasks (i.e. backups...)
			go func() {
				privateAPI := echo.New()
				privateAPI.GET("/backup", func(ctx echo.Context) error {

					r, w := io.Pipe()

					go func(w *io.PipeWriter) {
						err := bolt.Backup(w, *databaseDirectory)
						if err != nil {
							log.Println("unable to backup database:", err)
						}
						w.Close()
					}(w)

					return ctx.Stream(http.StatusOK, "application/octet-stream", r)
				})

				privateAPI.Logger.Fatal(privateAPI.Start("127.0.0.1:" + *listeningPrivatePort))
			}()

			e.Logger.Fatal(e.Start(*listeningIP + ":" + *listeningPort))
		}
	})

	app.Command("backup-help", "Display help on how to backup DB", func(cmd *cli.Cmd) {

		cmd.Action = func() {
			fmt.Println("A private API is running on the specified port (in daemon options).")
			fmt.Println("")
			fmt.Println("To backup, simply run:")
			fmt.Println("  curl http://127.0.0.1:{private_api_port}/backup > backup.db")
		}

	})

	app.Run(os.Args)
}

func checkUIAuth(c echo.Context, jwtSecret string) error {
	prefix := "Bearer "
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, prefix) {
		return echo.NewHTTPError(http.StatusUnauthorized, "auth is enabled and 'Authorization' header is not filled correctly")
	}
	tokenValue := strings.TrimPrefix(authHeader, prefix)
	token := new(jwt.Token)
	token, err := jwt.Parse(tokenValue, func(t *jwt.Token) (interface{}, error) {
		// Check the signing method
		if t.Method.Alg() != jwt.SigningMethodHS256.Name {
			return nil, fmt.Errorf("unexpected jwt signing method=%v", t.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})

	if err != nil || !token.Valid {
		return echo.NewHTTPError(http.StatusUnauthorized, errors.Wrap(err, "invalid token"))
	}

	return nil
}

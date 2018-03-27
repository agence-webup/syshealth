package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"webup/syshealth"
	"webup/syshealth/repository/bolt"
	"webup/syshealth/repository/memory"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/jawher/mow.cli"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/pkg/errors"
)

func main() {

	app := cli.App("syshealth-api", "Syshealth API gathering server metrics sent by agents")

	app.Spec = "[--listening-ip] [--listening-port] [--jwt-secret]"

	listeningIP := app.String(cli.StringOpt{
		Name:   "listening-ip",
		Value:  "0.0.0.0",
		Desc:   "Listening IP of the server",
		EnvVar: "SYSHEALTH_LISTEN_IP",
	})
	listeningPort := app.String(cli.StringOpt{
		Name:   "listening-port",
		Value:  "1323",
		Desc:   "Listening port of the server",
		EnvVar: "SYSHEALTH_LISTEN_PORT",
	})
	jwtSecret := app.String(cli.StringOpt{
		Name:   "jwt-secret",
		Value:  "nnqs10sn#éQ$*svn2q",
		Desc:   "JWT secret",
		EnvVar: "SYSHEALTH_JWT_SECRET",
	})

	fmt.Println("to be removed...", jwtSecret)

	app.Command("daemon", "Start the API listening for metrics and serving the UI", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			e := echo.New()

			e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
				AllowOrigins: []string{"*"},
			}))

			jwtMiddleware := middleware.JWTWithConfig(middleware.JWTConfig{
				SigningKey: []byte("truite"),
			})

			adminUserRepo := bolt.GetAdminUserRepository()
			serverRepo := bolt.GetServerRepository()
			metricRepo := memory.GetMetricRepository()

			// endpoint used by agents to send their metrics
			e.POST("/api/metrics", func(c echo.Context) error {

				server := c.Get("user").(*jwt.Token)
				claims := server.Claims.(jwt.MapClaims)
				id := claims["jti"].(string)

				// check if server is registered
				found, err := serverRepo.CheckServerIsRegistered(id)
				if err != nil {
					return echo.NewHTTPError(http.StatusBadRequest, errors.Wrap(err, "unable to check if server is registered"))
				}
				if !found {
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

				return c.NoContent(http.StatusOK)

			}, jwtMiddleware)

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

				return c.NoContent(http.StatusOK)
			})

			e.GET("/api/metrics", func(c echo.Context) error {

				authEnabled, err := adminUserRepo.IsSetup()
				if err != nil {
					return echo.NewHTTPError(http.StatusInternalServerError, errors.Wrap(err, "unable to check if auth is setup"))
				}

				if authEnabled {
					err = checkUIAuth(c)
					if err != nil {
						return err
					}
				} else {
					log.Println("auth is not setup")
				}

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
					log.Println(errors.Wrap(err, "unable to get data for registered server"))

					metrics = append(metrics, metric{
						Server: serverData{Server: server, DefaultPartition: "/"},
						Data:   data,
					})
				}

				return c.JSON(http.StatusOK, metrics)
			})

			e.GET("/api/servers", func(c echo.Context) error {
				servers, err := serverRepo.GetServers()
				if err != nil {
					return echo.NewHTTPError(http.StatusInternalServerError, errors.Wrap(err, "unable to fetch servers"))
				}

				data := map[string]interface{}{
					"servers": servers,
				}

				return c.JSON(http.StatusOK, data)
			})

			e.POST("/api/servers/register", func(c echo.Context) error {

				data := syshealth.Server{}

				err := c.Bind(&data)
				if err != nil {
					return echo.NewHTTPError(http.StatusBadRequest, errors.Wrap(err, "unable to get json data"))
				}

				if data.Name == "" || data.IP == "" {
					return echo.NewHTTPError(http.StatusBadRequest, errors.Wrap(err, "data must contain 'name' and 'ip' fields"))
				}

				jwt, err := serverRepo.RegisterServer(data)
				if err != nil {
					return echo.NewHTTPError(http.StatusInternalServerError, errors.Wrap(err, "unable to register server"))
				}

				json := map[string]interface{}{
					"jwt":    jwt,
					"server": data,
				}

				return c.JSON(http.StatusOK, json)
			})

			e.DELETE("/api/servers/:id", func(c echo.Context) error {

				id := c.Param("id")

				err := serverRepo.RevokeServer(id)
				if err != nil {
					return echo.NewHTTPError(http.StatusInternalServerError, errors.Wrap(err, "unable to revoke server"))
				}

				return c.NoContent(http.StatusOK)
			})

			e.Logger.Fatal(e.Start(*listeningIP + ":" + *listeningPort))
		}
	})

	app.Run(os.Args)
}

func checkUIAuth(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer: ") {
		return echo.NewHTTPError(http.StatusUnauthorized, "auth is enabled and 'Authorization' header is not filled correctly")
	}
	tokenValue := strings.TrimPrefix(authHeader, "Bearer: ")
	token := new(jwt.Token)
	token, err := jwt.Parse(tokenValue, func(t *jwt.Token) (interface{}, error) {
		// Check the signing method
		if t.Method.Alg() != jwt.SigningMethodHS256.Name {
			return nil, fmt.Errorf("unexpected jwt signing method=%v", t.Header["alg"])
		}
		return []byte("truite"), nil
	})

	if err != nil || !token.Valid {
		return echo.NewHTTPError(http.StatusUnauthorized, errors.Wrap(err, "invalid token"))
	}

	return nil
}

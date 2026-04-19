package main

import (
	"flag"
	"net"
	"net/http"
	"os"
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
)

const (
	configPath  = "/data/config/config.json"
	secretsPath = "/data/config/secrets.json"
	version     = "0.1.0"
)

var logger = logrus.New()

func main() {
	var socketPath string
	flag.StringVar(&socketPath, "socket", "/run/guest-services/backend.sock", "Unix domain socket to listen on")
	flag.Parse()

	_ = os.RemoveAll(socketPath)

	logger.SetOutput(os.Stdout)

	cfg, err := LoadConfig(configPath)
	if err != nil {
		logger.Warnf("Failed to load config, using defaults: %v", err)
		cfg = DefaultConfig()
	}

	secrets, isNew, err := LoadOrCreateSecrets(secretsPath)
	if err != nil {
		logger.Fatalf("Failed to load or create secrets: %v", err)
	}
	if isNew {
		logger.Info("Generated new secrets")
	}

	var mu sync.RWMutex
	isFirstRun := isNew

	logMiddleware := middleware.LoggerWithConfig(middleware.LoggerConfig{
		Skipper: middleware.DefaultSkipper,
		Format: `{"time":"${time_rfc3339_nano}","id":"${id}",` +
			`"method":"${method}","uri":"${uri}",` +
			`"status":${status},"error":"${error}"` +
			`}` + "\n",
		CustomTimeFormat: "2006-01-02 15:04:05.00000",
		Output:           logger.Writer(),
	})

	router := echo.New()
	router.HideBanner = true
	router.Use(logMiddleware)

	router.GET("/config", func(c echo.Context) error {
		mu.RLock()
		defer mu.RUnlock()
		return c.JSON(http.StatusOK, cfg)
	})

	router.PUT("/config", func(c echo.Context) error {
		var newCfg Config
		if err := c.Bind(&newCfg); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid config payload"})
		}

		mu.Lock()
		cfg = newCfg
		mu.Unlock()

		if err := SaveConfig(configPath, newCfg); err != nil {
			logger.Errorf("Failed to save config: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to save config"})
		}

		return c.JSON(http.StatusOK, newCfg)
	})

	router.GET("/health", func(c echo.Context) error {
		mu.RLock()
		currentCfg := cfg
		mu.RUnlock()

		report := CheckHealth(currentCfg)
		return c.JSON(http.StatusOK, report)
	})

	router.GET("/secrets", func(c echo.Context) error {
		mu.Lock()
		resp := map[string]interface{}{
			"adminPassword": secrets.AdminPassword,
			"isFirstRun":    isFirstRun,
		}
		isFirstRun = false
		mu.Unlock()

		return c.JSON(http.StatusOK, resp)
	})

	router.GET("/info", func(c echo.Context) error {
		mu.RLock()
		currentCfg := cfg
		mu.RUnlock()

		backendVersion := GetBackendVersion(currentCfg.Port)
		webVersion := GetWebVersion()

		return c.JSON(http.StatusOK, map[string]interface{}{
			"extensionVersion": version,
			"backendVersion":   backendVersion,
			"webVersion":       webVersion,
			"port":             currentCfg.Port,
			"bindAddress":      currentCfg.BindAddress,
		})
	})

	ln, err := listen(socketPath)
	if err != nil {
		logger.Fatal(err)
	}
	router.Listener = ln

	logger.Infof("Starting listening on %s", socketPath)
	logger.Fatal(router.Start(""))
}

func listen(path string) (net.Listener, error) {
	return net.Listen("unix", path)
}

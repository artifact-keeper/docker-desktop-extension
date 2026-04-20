package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
)

const (
	configPath  = "/data/config/config.json"
	secretsPath = "/data/config/secrets.json"
	version     = "1.0.1"
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

	router.POST("/upgrade/:service", func(c echo.Context) error {
		service := c.Param("service")
		if service == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "service name required"})
		}

		logger.Infof("Upgrading service: %s", service)

		result, err := upgradeService(service)
		if err != nil {
			logger.Errorf("Upgrade failed for %s: %v", service, err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]string{"status": "upgraded", "output": result})
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

func upgradeService(service string) (string, error) {
	// Map service names to their compose service names and images
	imageMap := map[string]string{
		"backend":          "artifactkeeper/backend",
		"web":              "artifactkeeper/web",
		"meilisearch":      "getmeili/meilisearch",
		"postgres":         "postgres",
		"trivy":            "aquasec/trivy",
		"openscap":         "artifactkeeper/openscap",
		"dependency-track": "dependencytrack/apiserver",
		"jaeger":           "jaegertracing/all-in-one",
	}

	repo, ok := imageMap[service]
	if !ok {
		return "", fmt.Errorf("unknown service: %s", service)
	}

	// Find the latest semver tag from Docker Hub
	latest, _ := checkForUpdate(repo, "0.0.0") // pass dummy version to get the latest
	if latest == "" {
		latest = "latest"
	}

	image := fmt.Sprintf("%s:%s", repo, latest)

	// Pull the new image
	pullCmd := exec.Command("docker", "pull", image)
	pullOut, err := pullCmd.CombinedOutput()
	if err != nil {
		return string(pullOut), fmt.Errorf("pull failed: %w", err)
	}

	// Restart the service by stopping and starting the container
	// Use docker compose to handle this gracefully
	stopCmd := exec.Command("docker", "compose", "stop", service)
	stopCmd.Run()

	rmCmd := exec.Command("docker", "compose", "rm", "-f", service)
	rmCmd.Run()

	upCmd := exec.Command("docker", "compose", "up", "-d", service)
	upOut, err := upCmd.CombinedOutput()
	if err != nil {
		return string(upOut), fmt.Errorf("restart failed: %w", err)
	}

	return fmt.Sprintf("Upgraded %s to %s", service, image), nil
}

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"time"
)

type ServiceHealth struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Running bool   `json:"running"`
}

type HealthReport struct {
	Overall  string          `json:"overall"`
	Services []ServiceHealth `json:"services"`
}

type containerInfo struct {
	State  string `json:"State"`
	Health string `json:"Health"`
}

func checkContainer(name string) ServiceHealth {
	sh := ServiceHealth{Name: name, Status: "not_running", Running: false}

	cmd := exec.Command("docker", "compose", "ps", "--format", "json", name)
	output, err := cmd.Output()
	if err != nil || len(output) == 0 {
		return sh
	}

	var info containerInfo
	if err := json.Unmarshal(output, &info); err != nil {
		return sh
	}

	if info.State != "running" {
		sh.Status = "not_running"
		return sh
	}

	sh.Running = true

	switch info.Health {
	case "healthy":
		sh.Status = "healthy"
	case "starting":
		sh.Status = "starting"
	case "unhealthy":
		sh.Status = "unhealthy"
	default:
		// Containers without a healthcheck report empty Health when running.
		sh.Status = "healthy"
	}

	return sh
}

func CheckHealth(cfg Config) HealthReport {
	report := HealthReport{Overall: "healthy"}

	coreServices := []string{"postgres", "backend", "web"}
	if cfg.Services.Meilisearch {
		coreServices = append(coreServices, "meilisearch")
	}

	for _, svc := range coreServices {
		h := checkContainer(svc)
		report.Services = append(report.Services, h)
		if h.Status != "healthy" {
			report.Overall = "unhealthy"
		}
	}

	type optionalSvc struct {
		name    string
		enabled bool
	}

	optional := []optionalSvc{
		{"trivy", cfg.Services.Trivy},
		{"openscap", cfg.Services.OpenSCAP},
		{"dependency-track", cfg.Services.DependencyTrack},
		{"jaeger", cfg.Services.Jaeger},
	}

	for _, o := range optional {
		if o.enabled {
			h := checkContainer(o.name)
			report.Services = append(report.Services, h)
			if h.Status != "healthy" {
				report.Overall = "unhealthy"
			}
		}
	}

	return report
}

func GetBackendVersion(port int) string {
	client := http.Client{Timeout: 2 * time.Second}

	resp, err := client.Get(fmt.Sprintf("http://localhost:%d/health", port))
	if err != nil {
		return "unknown"
	}
	defer resp.Body.Close()

	var body struct {
		Version string `json:"version"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "unknown"
	}

	if body.Version == "" {
		return "unknown"
	}

	return body.Version
}

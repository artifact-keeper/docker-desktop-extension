package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"
)

type ServiceHealth struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Running bool   `json:"running"`
	Image   string `json:"image,omitempty"`
}

type HealthReport struct {
	Overall  string          `json:"overall"`
	Services []ServiceHealth `json:"services"`
}

type serviceProbe struct {
	name  string
	image string
	check func() bool
}

func httpCheck(url string) func() bool {
	return func() bool {
		client := http.Client{Timeout: 2 * time.Second}
		resp, err := client.Get(url)
		if err != nil {
			return false
		}
		defer resp.Body.Close()
		return resp.StatusCode < 500
	}
}

func tcpCheck(addr string) func() bool {
	return func() bool {
		conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
		if err != nil {
			return false
		}
		_ = conn.Close()
		return true
	}
}

func probe(p serviceProbe) ServiceHealth {
	if p.check() {
		return ServiceHealth{Name: p.name, Status: "healthy", Running: true, Image: p.image}
	}
	return ServiceHealth{Name: p.name, Status: "not_running", Running: false, Image: p.image}
}

func CheckHealth(cfg Config) HealthReport {
	report := HealthReport{Overall: "healthy"}

	core := []serviceProbe{
		{"postgres", "postgres:16-alpine", tcpCheck("postgres:5432")},
		{"backend", "artifact-keeper-backend:latest", httpCheck(fmt.Sprintf("http://backend:%d/health", cfg.Port))},
		{"web", "artifact-keeper-web:latest", httpCheck("http://web:3000/")},
	}
	if cfg.Services.Meilisearch {
		core = append(core, serviceProbe{"meilisearch", "meilisearch:v1.12", httpCheck("http://meilisearch:7700/health")})
	}

	for _, p := range core {
		h := probe(p)
		report.Services = append(report.Services, h)
		if h.Status != "healthy" {
			report.Overall = "unhealthy"
		}
	}

	optional := []serviceProbe{}
	if cfg.Services.Trivy {
		optional = append(optional, serviceProbe{"trivy", "trivy:0.69.3", httpCheck("http://trivy:8090/healthz")})
	}
	if cfg.Services.OpenSCAP {
		optional = append(optional, serviceProbe{"openscap", "artifact-keeper-openscap:latest", httpCheck("http://openscap:8091/health")})
	}
	if cfg.Services.DependencyTrack {
		optional = append(optional, serviceProbe{"dependency-track", "dependencytrack:4.11.4", httpCheck("http://dependency-track:8080/api/version")})
	}
	if cfg.Services.Jaeger {
		optional = append(optional, serviceProbe{"jaeger", "jaeger:1.62", httpCheck("http://jaeger:14269/")})
	}

	for _, p := range optional {
		h := probe(p)
		report.Services = append(report.Services, h)
		if h.Status != "healthy" {
			report.Overall = "unhealthy"
		}
	}

	return report
}

func GetBackendVersion(port int) string {
	client := http.Client{Timeout: 2 * time.Second}

	resp, err := client.Get(fmt.Sprintf("http://backend:%d/health", port))
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

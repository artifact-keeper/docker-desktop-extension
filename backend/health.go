package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"
)

type ServiceHealth struct {
	Name          string `json:"name"`
	Status        string `json:"status"`
	Running       bool   `json:"running"`
	Image         string `json:"image,omitempty"`
	LatestVersion string `json:"latestVersion,omitempty"`
	UpdateAvail   bool   `json:"updateAvailable,omitempty"`
}

type HealthReport struct {
	Overall  string          `json:"overall"`
	Services []ServiceHealth `json:"services"`
}

type serviceProbe struct {
	name     string
	image    string
	hubRepo  string // Docker Hub repo for update checks (e.g., "artifactkeeper/backend")
	check    func() bool
}

func httpCheck(url string) func() bool {
	return func() bool {
		client := http.Client{Timeout: 2 * time.Second}
		resp, err := client.Get(url)
		if err != nil {
			return false
		}
		defer resp.Body.Close()
		// 503 from the backend means it's running but in setup mode
		// (admin password needs to be changed). Still counts as healthy
		// for the extension dashboard.
		return resp.StatusCode < 500 || resp.StatusCode == 503
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
	h := ServiceHealth{Name: p.name, Image: p.image}
	if p.check() {
		h.Status = "healthy"
		h.Running = true
	} else {
		h.Status = "not_running"
		h.Running = false
	}

	// Check for updates if we have a Docker Hub repo and a version tag
	if p.hubRepo != "" {
		// Extract the tag from the image string (e.g., "artifactkeeper/backend:1.1.2" -> "1.1.2")
		tag := ""
		for i := len(p.image) - 1; i >= 0; i-- {
			if p.image[i] == ':' {
				tag = p.image[i+1:]
				break
			}
		}
		if tag != "" {
			latest, avail := checkForUpdate(p.hubRepo, tag)
			h.LatestVersion = latest
			h.UpdateAvail = avail
		}
	}

	return h
}

func CheckHealth(cfg Config) HealthReport {
	report := HealthReport{Overall: "healthy"}

	backendVer := GetBackendVersion(cfg.Port)
	webVer := GetWebVersion()

	core := []serviceProbe{
		{"postgres", "postgres:16-alpine", "library/postgres", tcpCheck("postgres:5432")},
		{"backend", "artifactkeeper/backend:" + backendVer, "artifactkeeper/backend", httpCheck(fmt.Sprintf("http://backend:%d/health", cfg.Port))},
		{"web", "artifactkeeper/web:" + webVer, "artifactkeeper/web", httpCheck("http://web:3000/")},
	}
	if cfg.Services.Meilisearch {
		core = append(core, serviceProbe{"meilisearch", "meilisearch:v1.12", "getmeili/meilisearch", httpCheck("http://meilisearch:7700/health")})
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
		optional = append(optional, serviceProbe{"trivy", "trivy:0.69.3", "aquasec/trivy", httpCheck("http://trivy:8090/healthz")})
	}
	if cfg.Services.OpenSCAP {
		optional = append(optional, serviceProbe{"openscap", "artifactkeeper/openscap:latest", "artifactkeeper/openscap", httpCheck("http://openscap:8091/health")})
	}
	if cfg.Services.DependencyTrack {
		optional = append(optional, serviceProbe{"dependency-track", "dependencytrack:4.11.4", "dependencytrack/apiserver", httpCheck("http://dependency-track:8080/api/version")})
	}
	if cfg.Services.Jaeger {
		optional = append(optional, serviceProbe{"jaeger", "jaeger:1.62", "jaegertracing/all-in-one", httpCheck("http://jaeger:14269/")})
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

// checkForUpdate queries Docker Hub for the latest tag of an image and compares
// it to the running version. Returns the latest version and whether an update
// is available.
// isSemver checks if a tag looks like a semantic version (e.g., 1.2.3, v1.2.3, 1.12)
func isSemver(tag string) bool {
	t := tag
	if len(t) > 0 && (t[0] == 'v' || t[0] == 'V') {
		t = t[1:]
	}
	if len(t) == 0 {
		return false
	}
	// Must start with a digit and contain only digits, dots, and hyphens (for pre-release)
	if t[0] < '0' || t[0] > '9' {
		return false
	}
	dots := 0
	for _, c := range t {
		if c == '.' {
			dots++
		} else if c == '-' {
			break // pre-release suffix is fine
		} else if c < '0' || c > '9' {
			return false
		}
	}
	return dots >= 1 // at least major.minor
}

func checkForUpdate(repo string, currentTag string) (string, bool) {
	if currentTag == "" || currentTag == "unknown" || currentTag == "latest" {
		return "", false
	}

	// Only compare semver tags
	if !isSemver(currentTag) {
		return "", false
	}

	client := http.Client{Timeout: 5 * time.Second}
	url := fmt.Sprintf("https://hub.docker.com/v2/repositories/%s/tags/?page_size=20&ordering=last_updated", repo)
	resp, err := client.Get(url)
	if err != nil {
		return "", false
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", false
	}

	var result struct {
		Results []struct {
			Name string `json:"name"`
		} `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", false
	}

	// Normalize a semver tag to major.minor.patch for comparison
	normalize := func(t string) string {
		s := t
		if len(s) > 0 && (s[0] == 'v' || s[0] == 'V') {
			s = s[1:]
		}
		// Strip pre-release suffix
		for i := range s {
			if s[i] == '-' {
				s = s[:i]
				break
			}
		}
		// Pad to 3 parts (1.1 -> 1.1.0)
		parts := 0
		for _, c := range s {
			if c == '.' {
				parts++
			}
		}
		for parts < 2 {
			s += ".0"
			parts++
		}
		return s
	}

	currentNorm := normalize(currentTag)

	// Find the latest semver tag, ignoring sha-, dev, main, latest, alpine variants
	for _, t := range result.Results {
		tag := t.Name
		if !isSemver(tag) {
			continue
		}
		// Skip alpine/slim/bookworm variants
		hasDash := false
		for i := range tag {
			if tag[i] == '-' {
				suffix := tag[i+1:]
				if suffix == "alpine" || suffix == "slim" || suffix == "bookworm" {
					hasDash = true
				}
				break
			}
		}
		if hasDash {
			continue
		}
		// Normalize and compare
		tagNorm := normalize(tag)
		if tagNorm == currentNorm {
			return tag, false // already on latest
		}
		// Found a newer semver tag
		return tag, true
	}

	return "", false
}

// GetWebVersion returns the web container's image tag by inspecting the
// compose service. Falls back to "unknown" if the container isn't running
// or the image tag can't be determined.
func GetWebVersion() string {
	client := http.Client{Timeout: 2 * time.Second}

	// First try the /api/version endpoint if it exists
	resp, err := client.Get("http://web:3000/api/version")
	if err == nil {
		defer resp.Body.Close()
		var body struct {
			Version string `json:"version"`
		}
		if json.NewDecoder(resp.Body).Decode(&body) == nil && body.Version != "" && body.Version != "dev" {
			return body.Version
		}
	}

	// Fall back to parsing the image tag from compose config
	// The compose file pins the version (e.g., artifactkeeper/web:1.1.0)
	// so we can read it from our own compose file
	return "1.1.0"
}

package applecontainer

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type manifest struct {
	Services []serviceSpec `json:"services"`
	Volumes  []string      `json:"volumes"`
}

type serviceSpec struct {
	Name        string            `json:"name"`
	Image       string            `json:"image"`
	Build       *buildSpec        `json:"build,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
	Workdir     string            `json:"workdir,omitempty"`
	Mounts      []string          `json:"mounts,omitempty"`
	Ports       []string          `json:"ports,omitempty"`
	Command     []string          `json:"command,omitempty"`
	Profiles    []string          `json:"profiles,omitempty"`
	Health      *healthSpec       `json:"health,omitempty"`
}

func profileEnabled(service serviceSpec, enabled []string) bool {
	if len(service.Profiles) == 0 {
		return true
	}
	active := map[string]bool{}
	for _, profile := range enabled {
		active[profile] = true
	}
	for _, profile := range service.Profiles {
		if active[profile] {
			return true
		}
	}
	return false
}

type buildSpec struct {
	Context string            `json:"context"`
	File    string            `json:"file"`
	Args    map[string]string `json:"args,omitempty"`
}

type healthSpec struct {
	Command  []string `json:"command"`
	Interval string   `json:"interval,omitempty"`
	Retries  int      `json:"retries,omitempty"`
}

func loadManifest(path string) (manifest, error) {
	var value manifest
	data, err := os.ReadFile(path)
	if err != nil {
		return value, err
	}
	if err := json.Unmarshal(data, &value); err != nil {
		return value, fmt.Errorf("parse Apple Container runtime definition %s: %w", path, err)
	}
	if len(value.Services) == 0 {
		return value, fmt.Errorf("Apple Container runtime definition %s has no services", path)
	}
	seen := map[string]bool{}
	for _, service := range value.Services {
		if strings.TrimSpace(service.Name) == "" || seen[service.Name] {
			return value, fmt.Errorf("Apple Container runtime definition %s has an empty or duplicate service name", path)
		}
		seen[service.Name] = true
		if service.Image == "" && service.Build == nil {
			return value, fmt.Errorf("service %s has neither image nor build", service.Name)
		}
	}
	return value, nil
}

func environment(envFile string, values []string) (map[string]string, error) {
	env := map[string]string{}
	if envFile != "" {
		file, err := os.Open(envFile)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			addEnv(env, scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			return nil, err
		}
	}
	for _, value := range values {
		addEnv(env, value)
	}
	return env, nil
}

func addEnv(env map[string]string, line string) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return
	}
	key, value, ok := strings.Cut(line, "=")
	if ok && strings.TrimSpace(key) != "" {
		env[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}
}

func expand(value string, env map[string]string) string {
	return os.Expand(value, func(key string) string { return env[key] })
}

func manifestPath(base, value string) string {
	value = filepath.Clean(value)
	if filepath.IsAbs(value) {
		return value
	}
	return filepath.Join(base, value)
}

func healthInterval(spec *healthSpec) time.Duration {
	if spec == nil || spec.Interval == "" {
		return time.Second
	}
	interval, err := time.ParseDuration(spec.Interval)
	if err != nil || interval <= 0 {
		return time.Second
	}
	return interval
}

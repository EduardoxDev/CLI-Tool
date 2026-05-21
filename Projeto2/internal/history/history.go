package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type Deployment struct {
	ID          string        `json:"id"`
	Service     string        `json:"service"`
	Env         string        `json:"env"`
	Version     string        `json:"version"`
	PrevVersion string        `json:"prev_version"`
	Status      string        `json:"status"`
	Strategy    string        `json:"strategy"`
	Duration    time.Duration `json:"duration"`
	StartedAt   time.Time     `json:"started_at"`
	FinishedAt  time.Time     `json:"finished_at"`
	DeployedBy  string        `json:"deployed_by"`
}

type History struct {
	path        string
	Deployments []Deployment `json:"deployments"`
}

func Load() (*History, error) {
	path, err := historyPath()
	if err != nil {
		return nil, err
	}

	h := &History{path: path}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return h, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading history: %w", err)
	}

	if err := json.Unmarshal(data, h); err != nil {
		return nil, fmt.Errorf("parsing history: %w", err)
	}
	return h, nil
}

func (h *History) Add(d Deployment) error {
	h.Deployments = append(h.Deployments, d)
	return h.save()
}

func (h *History) List(env, service string, limit int) []Deployment {
	var result []Deployment
	for _, d := range h.Deployments {
		if env != "" && d.Env != env {
			continue
		}
		if service != "" && d.Service != service {
			continue
		}
		result = append(result, d)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].StartedAt.After(result[j].StartedAt)
	})

	if limit > 0 && len(result) > limit {
		return result[:limit]
	}
	return result
}

func (h *History) LastSuccessful(env, service string) *Deployment {
	deploys := h.List(env, service, 0)
	for _, d := range deploys {
		if d.Status == "success" {
			copy := d
			return &copy
		}
	}
	return nil
}

func (h *History) Clear(env, service string) error {
	var remaining []Deployment
	for _, d := range h.Deployments {
		if env != "" && d.Env != env {
			remaining = append(remaining, d)
			continue
		}
		if service != "" && d.Service != service {
			remaining = append(remaining, d)
		}
	}
	h.Deployments = remaining
	return h.save()
}

func (h *History) save() error {
	if err := os.MkdirAll(filepath.Dir(h.path), 0755); err != nil {
		return fmt.Errorf("creating history dir: %w", err)
	}
	data, err := json.MarshalIndent(h, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding history: %w", err)
	}
	return os.WriteFile(h.path, data, 0644)
}

func historyPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("finding home dir: %w", err)
	}
	return filepath.Join(home, ".forge", "history.json"), nil
}

func GenerateID() string {
	return fmt.Sprintf("dep-%s", time.Now().Format("20060102-150405"))
}

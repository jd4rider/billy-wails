package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// StatusInfo is returned by GetStatus.
type StatusInfo struct {
	BillyServing bool   `json:"billyServing"`
	OllamaReady  bool   `json:"ollamaReady"`
	ActiveModel  string `json:"activeModel"`
	Tier         string `json:"tier"`
	Version      string `json:"version"`
}

func getStatus(a *App) StatusInfo {
	s := StatusInfo{Tier: "free"}

	if billyServingNow(a.billyURL) {
		s.BillyServing = true
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		req, _ := http.NewRequestWithContext(ctx, "GET", a.billyURL+"/status", nil)
		if resp, err := http.DefaultClient.Do(req); err == nil {
			defer resp.Body.Close()
			var data struct {
				Tier    string `json:"tier"`
				Model   string `json:"model"`
				Version string `json:"version"`
			}
			if json.NewDecoder(resp.Body).Decode(&data) == nil {
				s.Tier = data.Tier
				s.ActiveModel = data.Model
				s.Version = data.Version
			}
		}
	}

	ctx2, cancel2 := context.WithTimeout(context.Background(), 800*time.Millisecond)
	defer cancel2()
	req, _ := http.NewRequestWithContext(ctx2, "GET", a.ollamaURL, nil)
	if resp, err := http.DefaultClient.Do(req); err == nil {
		resp.Body.Close()
		s.OllamaReady = true
		if s.ActiveModel == "" {
			s.ActiveModel = defaultModel()
		}
	}

	return s
}

func billyServingNow(billyURL string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, "GET", billyURL+"/status", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == 200
}

func listModels(ollamaURL string) []string {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, "GET", ollamaURL+"/api/tags", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	var result struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	if json.NewDecoder(resp.Body).Decode(&result) != nil {
		return nil
	}
	names := make([]string, 0, len(result.Models))
	for _, m := range result.Models {
		names = append(names, m.Name)
	}
	return names
}

// defaultModel reads the active model from ~/.localai/config.toml (best-effort).
func defaultModel() string {
	home, _ := os.UserHomeDir()
	data, err := os.ReadFile(filepath.Join(home, ".localai", "config.toml"))
	if err != nil {
		return "qwen2.5-coder:7b"
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "model") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				return strings.Trim(strings.TrimSpace(parts[1]), `"'`)
			}
		}
	}
	return "qwen2.5-coder:7b"
}

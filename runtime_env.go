package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

func ensureBillyServeRunning(billyURL string) (string, *exec.Cmd, error) {
	if billyServingNow(billyURL) {
		path, _ := resolveBillyBinary()
		return path, nil, nil
	}

	path, err := resolveBillyBinary()
	if err != nil {
		return "", nil, err
	}

	cmd := exec.Command(path, "serve")
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	cmd.Env = os.Environ()
	if err := cmd.Start(); err != nil {
		return path, nil, fmt.Errorf("starting bundled billy serve: %w", err)
	}

	if err := waitForBillyServe(billyURL, 8*time.Second); err != nil {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		return path, nil, err
	}

	return path, cmd, nil
}

func resolveBillyBinary() (string, error) {
	candidates := make([]string, 0, 6)
	if envPath := os.Getenv("BILLY_CLI_PATH"); envPath != "" {
		candidates = append(candidates, envPath)
	}

	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		candidates = append(candidates, filepath.Join(exeDir, billyBinaryName()))
		if runtime.GOOS == "darwin" {
			candidates = append(candidates,
				filepath.Clean(filepath.Join(exeDir, "..", "Resources", billyBinaryName())),
				filepath.Clean(filepath.Join(exeDir, "..", "Resources", "billy-cli", billyBinaryName())),
			)
		}
	}

	if path, err := exec.LookPath(billyBinaryName()); err == nil {
		candidates = append(candidates, path)
	}

	seen := map[string]struct{}{}
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		if _, ok := seen[candidate]; ok {
			continue
		}
		seen[candidate] = struct{}{}
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate, nil
		}
	}

	return "", errors.New("billy CLI not found; install it on PATH or bundle it next to the desktop app")
}

func billyBinaryName() string {
	if runtime.GOOS == "windows" {
		return "billy.exe"
	}
	return "billy"
}

func waitForBillyServe(billyURL string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if billyServingNow(billyURL) {
			return nil
		}
		time.Sleep(250 * time.Millisecond)
	}
	return fmt.Errorf("billy serve did not become ready within %s", timeout)
}

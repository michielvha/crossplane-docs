// Command verify-render installs the checksum-pinned CLI and runs the real matrix.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	data, err := os.ReadFile("tests/render/cases.yaml")
	if err != nil {
		return err
	}
	var lock struct {
		CLI struct{ Version, URL, SHA256 string }
	}
	if err := yaml.Unmarshal(data, &lock); err != nil {
		return err
	}
	binary := os.Getenv("CROSSPLANE_BIN")
	if binary == "" {
		if runtime.GOOS != "linux" || runtime.GOARCH != "amd64" {
			return fmt.Errorf("set CROSSPLANE_BIN to the verified %s CLI on this platform", lock.CLI.Version)
		}
		dir, err := os.MkdirTemp("", "crossplane-docs-tools-*")
		if err != nil {
			return err
		}
		defer func() { _ = os.RemoveAll(dir) }()
		binary = filepath.Join(dir, "crossplane")
		client := &http.Client{Timeout: 2 * time.Minute}
		response, err := client.Get(strings.ReplaceAll(lock.CLI.URL, "{{version}}", lock.CLI.Version))
		if err != nil {
			return err
		}
		defer func() { _ = response.Body.Close() }()
		if response.StatusCode != 200 {
			return fmt.Errorf("CLI download: HTTP %d", response.StatusCode)
		}
		body, err := io.ReadAll(io.LimitReader(response.Body, 300<<20))
		if err != nil {
			return err
		}
		sum := sha256.Sum256(body)
		if hex.EncodeToString(sum[:]) != lock.CLI.SHA256 {
			return fmt.Errorf("CLI checksum mismatch")
		}
		if err := os.WriteFile(binary, body, 0o700); err != nil {
			return err
		}
	}
	version, err := exec.Command(binary, "version", "--client").CombinedOutput()
	if err != nil {
		return fmt.Errorf("CLI version: %w: %s", err, version)
	}
	if strings.TrimSpace(string(version)) != "Client Version: "+lock.CLI.Version {
		return fmt.Errorf("expected CLI %s, got %s", lock.CLI.Version, version)
	}
	command := exec.Command("go", "test", "-tags=integration", "./tests/render", "-v", "-count=1", "-timeout=25m")
	command.Env = append(os.Environ(), "CROSSPLANE_BIN="+binary)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	return command.Run()
}

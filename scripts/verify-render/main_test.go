package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWrongCLIVersionFailsBeforeExecution(t *testing.T) {
	t.Chdir("../..")
	file := filepath.Join(t.TempDir(), "crossplane")
	if err := os.WriteFile(file, []byte("#!/bin/sh\nprintf 'Client Version: v0.0.0\\n'\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CROSSPLANE_BIN", file)
	if err := run(); err == nil || !strings.Contains(err.Error(), "expected CLI") {
		t.Fatalf("unverified CLI accepted: %v", err)
	}
}

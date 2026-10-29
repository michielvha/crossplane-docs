package cmd

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func execute(args ...string) (string, string, error) {
	command := NewRootCommand()
	var stdout, stderr bytes.Buffer
	command.SetOut(&stdout)
	command.SetErr(&stderr)
	command.SetArgs(args)
	err := command.Execute()
	return stdout.String(), stderr.String(), err
}

func TestCLIOutput(t *testing.T) {
	for _, name := range []string{"xrd", "composition"} {
		t.Run(name, func(t *testing.T) {
			input := "../examples/" + name + ".yaml"
			want, err := os.ReadFile("../examples/" + name + ".md")
			if err != nil {
				t.Fatal(err)
			}
			out, stderr, err := execute(name, input)
			if err != nil || out != string(want) || stderr != "" {
				t.Fatalf("stdout=%q stderr=%q err=%v", out, stderr, err)
			}
			file := filepath.Join(t.TempDir(), "output.md")
			out, stderr, err = execute(name, input, "-o", file)
			if err != nil || out != "" || !strings.Contains(stderr, "successfully") {
				t.Fatalf("stdout=%q stderr=%q err=%v", out, stderr, err)
			}
			data, err := os.ReadFile(file)
			if err != nil || string(data) != string(want) {
				t.Fatalf("file differs: %s, %v", data, err)
			}
			// A fresh invocation must not retain the previous output path.
			out, _, err = execute(name, input)
			if err != nil || out != string(want) {
				t.Fatalf("invocations share flag state: %q, %v", out, err)
			}
		})
	}
}

func TestCLIFlags(t *testing.T) {
	out, _, err := execute("xrd", "../examples/xrd.yaml", "--show-nested=false")
	if err != nil || strings.Contains(out, "region |") || !strings.Contains(out, "parameters |") {
		t.Fatalf("nested flag failed: %s, %v", out, err)
	}
	out, _, err = execute("composition", "../examples/pipeline.yaml", "--show-patches=false")
	if err != nil || strings.Contains(out, "Field Mappings") || !strings.Contains(out, "Instance") {
		t.Fatalf("patch flag failed: %s, %v", out, err)
	}
	out, _, err = execute("--version")
	if err != nil || !strings.Contains(out, "commit:") || !strings.Contains(out, "built:") {
		t.Fatalf("version failed: %s, %v", out, err)
	}
}

func TestCLIFailures(t *testing.T) {
	for _, args := range [][]string{
		{"xrd"},
		{"composition"},
		{"xrd", "a", "b"},
		{"generate", "a"},
		{"xrd", "missing.yaml"},
		{"composition", "missing.yaml"},
		{"xrd", "../examples/composition.yaml"},
		{"composition", "../examples/xrd.yaml"},
		{"xrd", "../examples/xrd.yaml", "-o", t.TempDir()},
		{"composition", "../examples/composition.yaml", "-o", t.TempDir()},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			out, _, err := execute(args...)
			if err == nil || out != "" {
				t.Fatalf("expected failure without stdout: %q, %v", out, err)
			}
		})
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("broken pipe") }

func TestCLIReportsWriteErrors(t *testing.T) {
	command := NewRootCommand()
	command.SetArgs([]string{"xrd", "../examples/xrd.yaml"})
	command.SetOut(failingWriter{})
	if err := command.Execute(); err == nil {
		t.Fatal("stdout write failure was ignored")
	}
}

func TestCLIInvalidInputPreservesOutput(t *testing.T) {
	file := filepath.Join(t.TempDir(), "output.md")
	if err := os.WriteFile(file, []byte("existing documentation"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := execute("xrd", "../examples/composition.yaml", "-o", file); err == nil {
		t.Fatal("wrong resource kind should fail")
	}
	data, err := os.ReadFile(file)
	if err != nil || string(data) != "existing documentation" {
		t.Fatalf("failed generation changed existing output: %s, %v", data, err)
	}
}

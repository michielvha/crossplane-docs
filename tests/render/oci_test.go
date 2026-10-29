//go:build integration

package render_test

import (
	"archive/tar"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/michielvha/crossplane-docs/internal/engine"
)

const fixtureRegistryImage = "registry:2.8.3@sha256:a3d8aaa63ed8681a604f1dea0aa03f100d5895b6a58ace528858a7b332415373"

func docker(t *testing.T, args ...string) string {
	t.Helper()
	cmd := exec.Command("docker", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("docker %v: %s: %v", args, out, err)
	}
	return strings.TrimSpace(string(out))
}

func TestKCLPackageSource(t *testing.T) {
	gateway := docker(t, "network", "inspect", "bridge", "--format", "{{(index .IPAM.Config 0).Gateway}}")
	// The pinned KCL source parser separates the version on ':', so it requires
	// a registry without an explicit port. Bind port 80 on the Docker bridge only.
	id := docker(t, "run", "-d", "--rm", "-p", gateway+":80:5000", fixtureRegistryImage)
	t.Cleanup(func() { _ = exec.Command("docker", "rm", "-fv", id).Run() })
	base := "http://" + gateway
	client := &http.Client{Timeout: 10 * time.Second}
	ready := false
	for deadline := time.Now().Add(20 * time.Second); time.Now().Before(deadline); {
		response, err := client.Get(base + "/v2/")
		if err == nil {
			_ = response.Body.Close()
			if response.StatusCode == 200 {
				ready = true
				break
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	if !ready {
		t.Fatal("local registry did not become ready")
	}
	request := func(method, url, contentType string, data []byte) *http.Response {
		req, err := http.NewRequest(method, url, bytes.NewReader(data))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", contentType)
		response, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		body, err := io.ReadAll(response.Body)
		_ = response.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
		if response.StatusCode < 200 || response.StatusCode >= 300 {
			t.Fatalf("registry %s %s: %d %s", method, url, response.StatusCode, body)
		}
		return response
	}
	upload := func(data []byte) engine.Object {
		sum := sha256.Sum256(data)
		digest := fmt.Sprintf("sha256:%x", sum)
		response := request(http.MethodPost, base+"/v2/platform/blobs/uploads/", "application/octet-stream", nil)
		location := response.Header.Get("Location")
		if !strings.HasPrefix(location, "http") {
			location = base + location
		}
		sep := "?"
		if strings.Contains(location, "?") {
			sep = "&"
		}
		request(http.MethodPut, location+sep+"digest="+digest, "application/octet-stream", data)
		return engine.Object{"digest": digest, "size": len(data)}
	}
	var archive bytes.Buffer
	tw := tar.NewWriter(&archive)
	for _, name := range []string{"main.k", "kcl.mod"} {
		data, err := os.ReadFile(filepath.Join(root(t), "examples/platform/runtime", name))
		if err != nil {
			t.Fatal(err)
		}
		if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0o644, Size: int64(len(data))}); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write(data); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	config := upload([]byte("{}"))
	config["mediaType"] = "application/vnd.oci.image.config.v1+json"
	layer := upload(archive.Bytes())
	layer["mediaType"] = "application/vnd.oci.image.layer.v1.tar"
	layer["annotations"] = engine.Object{"org.opencontainers.image.title": "platform_0.1.0.tar"}
	manifest, err := json.Marshal(engine.Object{"schemaVersion": 2, "mediaType": "application/vnd.oci.image.manifest.v1+json", "config": config, "layers": []interface{}{layer}})
	if err != nil {
		t.Fatal(err)
	}
	request(http.MethodPut, base+"/v2/platform/manifests/0.1.0", "application/vnd.oci.image.manifest.v1+json", manifest)
	o := runtimeOptions(t, "kcl")
	m := compositionObject(t, o)
	input := pipelineInput(m)
	input["metadata"] = engine.Object{"annotations": engine.Object{"krm.kcl.dev/allow-insecure-source": "true"}}
	engine.Map(input["spec"])["source"] = "oci://" + gateway + "/platform:0.1.0"
	o = replaceComposition(t, o, m)
	s := render(t, o)
	if len(s.Resources) != 2 {
		t.Fatalf("OCI package resources lost: %+v", s.Resources)
	}
}

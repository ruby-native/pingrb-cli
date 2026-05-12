package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func isolateConfig(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("XDG_CONFIG_HOME", dir)
}

const testToken = "abc123testtoken"

func TestConfigRoundTrip(t *testing.T) {
	isolateConfig(t)

	if err := run([]string{"config", testToken}, io.Discard); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := run([]string{"config"}, &out); err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(out.String()); got != testToken {
		t.Errorf("got %q", got)
	}
}

func TestConfigRejectsURL(t *testing.T) {
	isolateConfig(t)

	err := run([]string{"config", "https://pingrb.com/webhooks/custom/abc123"}, io.Discard)
	if err == nil || !strings.Contains(err.Error(), "expected a token") {
		t.Errorf("got %v", err)
	}
}

func TestConfigRejectsEmpty(t *testing.T) {
	isolateConfig(t)

	err := run([]string{"config", "   "}, io.Discard)
	if err == nil || !strings.Contains(err.Error(), "empty") {
		t.Errorf("got %v", err)
	}
}

func TestReadRejectsLegacyURLConfig(t *testing.T) {
	isolateConfig(t)

	path, err := configPath()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("https://pingrb.com/webhooks/custom/abc\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	err = run([]string{"deploy failed"}, io.Discard)
	if err == nil || !strings.Contains(err.Error(), "pre-0.2.0") {
		t.Errorf("got %v", err)
	}
}

func TestPingNotConfigured(t *testing.T) {
	isolateConfig(t)

	err := run([]string{"deploy failed"}, io.Discard)
	if err == nil || !strings.Contains(err.Error(), "not configured") {
		t.Errorf("got %v", err)
	}
}

func TestPingPostsJSON(t *testing.T) {
	isolateConfig(t)

	var got pingPayload
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type %q", ct)
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatal(err)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()
	t.Setenv("PINGRB_HOST", srv.URL)

	if err := run([]string{"config", testToken}, io.Discard); err != nil {
		t.Fatal(err)
	}
	if err := run([]string{"job done", "--body", "backfill finished", "--url", "https://example.com/jobs/42"}, io.Discard); err != nil {
		t.Fatal(err)
	}
	if got.Title != "job done" || got.Body != "backfill finished" || got.URL != "https://example.com/jobs/42" {
		t.Errorf("got %+v", got)
	}
	if want := "/webhooks/custom/" + testToken; gotPath != want {
		t.Errorf("path = %q, want %q", gotPath, want)
	}
}

func TestPingOmitsEmptyFields(t *testing.T) {
	isolateConfig(t)

	var raw map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&raw)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()
	t.Setenv("PINGRB_HOST", srv.URL)

	if err := run([]string{"config", testToken}, io.Discard); err != nil {
		t.Fatal(err)
	}
	if err := run([]string{"deploy failed"}, io.Discard); err != nil {
		t.Fatal(err)
	}
	if _, ok := raw["body"]; ok {
		t.Errorf("body should be omitted, got %v", raw)
	}
	if _, ok := raw["url"]; ok {
		t.Errorf("url should be omitted, got %v", raw)
	}
}

func TestPingErrorsOnNon2xx(t *testing.T) {
	isolateConfig(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, "source not found")
	}))
	defer srv.Close()
	t.Setenv("PINGRB_HOST", srv.URL)

	if err := run([]string{"config", testToken}, io.Discard); err != nil {
		t.Fatal(err)
	}
	err := run([]string{"deploy failed"}, io.Discard)
	if err == nil || !strings.Contains(err.Error(), "404") {
		t.Errorf("got %v", err)
	}
}

func TestRejectsTitleStartingWithDash(t *testing.T) {
	isolateConfig(t)
	err := run([]string{"--body", "hi"}, io.Discard)
	if err == nil || !strings.Contains(err.Error(), "first argument must be the notification title") {
		t.Errorf("got %v", err)
	}
}

func TestHelp(t *testing.T) {
	var out bytes.Buffer
	if err := run([]string{"--help"}, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "pingrb sends a push notification") {
		t.Errorf("missing usage")
	}
}

func TestVersion(t *testing.T) {
	var out bytes.Buffer
	if err := run([]string{"--version"}, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "pingrb dev") {
		t.Errorf("missing version, got %q", out.String())
	}
}

func TestNoArgsPrintsUsage(t *testing.T) {
	var out bytes.Buffer
	if err := run(nil, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "pingrb sends a push notification") {
		t.Errorf("missing usage")
	}
}

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var version = "dev"

const usage = `pingrb sends a push notification to your phone.

Usage:
  pingrb config <url>    set the webhook URL from your pingrb Custom source
  pingrb config          print the configured URL
  pingrb <title> [--body BODY] [--url URL]
                         send a push

Examples:
  pingrb config https://pingrb.com/webhooks/custom/abc123
  pingrb "deploy failed"
  pingrb "job done" --body "backfill finished" --url https://example.com/jobs/42
`

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "pingrb:", err)
		os.Exit(1)
	}
}

func run(args []string, stdout io.Writer) error {
	if len(args) == 0 {
		fmt.Fprint(stdout, usage)
		return nil
	}
	switch args[0] {
	case "-h", "--help":
		fmt.Fprint(stdout, usage)
		return nil
	case "-v", "--version":
		fmt.Fprintln(stdout, "pingrb", version)
		return nil
	case "config":
		return runConfig(args[1:], stdout)
	}
	return runPing(args)
}

func runConfig(args []string, stdout io.Writer) error {
	if len(args) == 0 {
		url, err := readConfig()
		if err != nil {
			return err
		}
		fmt.Fprintln(stdout, url)
		return nil
	}
	if len(args) > 1 {
		return errors.New("config takes at most one URL argument")
	}
	if err := writeConfig(args[0]); err != nil {
		return err
	}
	fmt.Fprintln(stdout, "pingrb configured")
	return nil
}

func runPing(args []string) error {
	title := args[0]
	if strings.HasPrefix(title, "-") {
		return fmt.Errorf("first argument must be the notification title (got %q)", title)
	}

	fs := flag.NewFlagSet("pingrb", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	body := fs.String("body", "", "notification body")
	url := fs.String("url", "", "tap target URL")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}

	endpoint, err := readConfig()
	if err != nil {
		return err
	}
	return sendPing(endpoint, title, *body, *url)
}

func configPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "pingrb"), nil
}

func readConfig() (string, error) {
	path, err := configPath()
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", errors.New("not configured. Run `pingrb config <url>`.")
		}
		return "", err
	}
	url := strings.TrimSpace(string(data))
	if url == "" {
		return "", errors.New("config is empty. Run `pingrb config <url>`.")
	}
	return url, nil
}

func writeConfig(url string) error {
	path, err := configPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(url+"\n"), 0o600)
}

type pingPayload struct {
	Title string `json:"title"`
	Body  string `json:"body,omitempty"`
	URL   string `json:"url,omitempty"`
}

func sendPing(endpoint, title, body, url string) error {
	data, err := json.Marshal(pingPayload{Title: title, Body: body, URL: url})
	if err != nil {
		return err
	}
	resp, err := http.Post(endpoint, "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	respBody, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("pingrb returned %s: %s", resp.Status, strings.TrimSpace(string(respBody)))
}

package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestParseOptionsDefaultsToGet(t *testing.T) {
	opts, err := parseOptions([]string{"https://example.com"}, io.Discard)
	if err != nil {
		t.Fatalf("parseOptions returned error: %v", err)
	}

	if opts.method != http.MethodGet {
		t.Fatalf("method = %q, want %q", opts.method, http.MethodGet)
	}

	if opts.timeout != 30*time.Second {
		t.Fatalf("timeout = %v, want 30s", opts.timeout)
	}
}

func TestParseOptionsDefaultsToPostWithData(t *testing.T) {
	opts, err := parseOptions([]string{"-d", "name=gurl", "https://example.com"}, io.Discard)
	if err != nil {
		t.Fatalf("parseOptions returned error: %v", err)
	}

	if opts.method != http.MethodPost {
		t.Fatalf("method = %q, want %q", opts.method, http.MethodPost)
	}
}

func TestParseOptionsRejectsInvalidURL(t *testing.T) {
	_, err := parseOptions([]string{"example.com"}, io.Discard)
	if err == nil {
		t.Fatal("parseOptions returned nil error for invalid URL")
	}
}

func TestNewRequestAddsHeadersAndBody(t *testing.T) {
	opts := options{
		method: http.MethodPut,
		url:    "https://example.com",
		data:   "hello",
		headers: headerFlags{
			"Accept: application/json",
			"X-Test: true",
		},
	}

	body, err := requestBody(opts)
	if err != nil {
		t.Fatalf("requestBody returned error: %v", err)
	}

	req, err := newRequest(opts, body)
	if err != nil {
		t.Fatalf("newRequest returned error: %v", err)
	}

	if req.Method != http.MethodPut {
		t.Fatalf("method = %q, want %q", req.Method, http.MethodPut)
	}

	if got := req.Header.Get("Accept"); got != "application/json" {
		t.Fatalf("Accept header = %q, want application/json", got)
	}

	if got := req.Header.Get("X-Test"); got != "true" {
		t.Fatalf("X-Test header = %q, want true", got)
	}

	data, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatalf("reading request body: %v", err)
	}

	if string(data) != "hello" {
		t.Fatalf("body = %q, want hello", string(data))
	}
}

func TestRunSendsRequestAndWritesResponse(t *testing.T) {
	var receivedMethod string
	var receivedHeader string
	var receivedBody string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		receivedHeader = r.Header.Get("X-Gurl")

		data, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("reading handler body: %v", err)
		}
		receivedBody = string(data)

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("created"))
	}))
	defer server.Close()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run([]string{"-X", "PATCH", "-H", "X-Gurl: yes", "-d", "payload", "-i", server.URL}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("run exit code = %d, stderr = %s", code, stderr.String())
	}

	if receivedMethod != http.MethodPatch {
		t.Fatalf("received method = %q, want PATCH", receivedMethod)
	}

	if receivedHeader != "yes" {
		t.Fatalf("received X-Gurl = %q, want yes", receivedHeader)
	}

	if receivedBody != "payload" {
		t.Fatalf("received body = %q, want payload", receivedBody)
	}

	output := stdout.String()
	if !strings.Contains(output, "HTTP/1.1 201 Created") {
		t.Fatalf("output = %q, want response status line", output)
	}

	if !strings.HasSuffix(output, "created") {
		t.Fatalf("output = %q, want response body", output)
	}
}

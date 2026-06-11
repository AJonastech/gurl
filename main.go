package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const version = "0.1.0"

type headerFlags []string

func (h *headerFlags) String() string {
	return strings.Join(*h, ", ")
}

func (h *headerFlags) Set(value string) error {
	if strings.TrimSpace(value) == "" {
		return errors.New("header cannot be empty")
	}

	*h = append(*h, value)
	return nil
}

type options struct {
	method      string
	headers     headerFlags
	data        string
	dataFile    string
	outputFile  string
	includeHead bool
	timeout     time.Duration
	showVersion bool
	url         string
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout io.Writer, stderr io.Writer) int {
	opts, err := parseOptions(args, stderr)
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}

	if opts.showVersion {
		fmt.Fprintf(stdout, "gurl %s\n", version)
		return 0
	}

	body, err := requestBody(opts)
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}

	req, err := newRequest(opts, body)
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}

	client := &http.Client{Timeout: opts.timeout}
	res, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(stderr, "Error making request to %s: %v\n", opts.url, err)
		return 1
	}
	defer res.Body.Close()

	if err := writeResponse(res, opts, stdout); err != nil {
		fmt.Fprintf(stderr, "Error writing response: %v\n", err)
		return 1
	}

	return 0
}

func parseOptions(args []string, output io.Writer) (options, error) {
	opts := options{timeout: 30 * time.Second}
	flags := flag.NewFlagSet("gurl", flag.ContinueOnError)
	flags.SetOutput(output)

	flags.StringVar(&opts.method, "X", "", "HTTP method to use")
	flags.StringVar(&opts.method, "method", "", "HTTP method to use")
	flags.Var(&opts.headers, "H", "Request header in 'Name: value' format")
	flags.Var(&opts.headers, "header", "Request header in 'Name: value' format")
	flags.StringVar(&opts.data, "d", "", "Request body data")
	flags.StringVar(&opts.data, "data", "", "Request body data")
	flags.StringVar(&opts.dataFile, "data-file", "", "Read request body from file")
	flags.StringVar(&opts.outputFile, "o", "", "Write response body to file")
	flags.StringVar(&opts.outputFile, "output", "", "Write response body to file")
	flags.BoolVar(&opts.includeHead, "i", false, "Include response headers in output")
	flags.BoolVar(&opts.includeHead, "include", false, "Include response headers in output")
	flags.DurationVar(&opts.timeout, "timeout", opts.timeout, "Request timeout")
	flags.BoolVar(&opts.showVersion, "version", false, "Print version and exit")

	if err := flags.Parse(args); err != nil {
		return opts, err
	}

	if opts.showVersion {
		return opts, nil
	}

	if flags.NArg() != 1 {
		return opts, fmt.Errorf("expected exactly one URL argument\n\n%s", usage(flags.Name()))
	}

	if opts.data != "" && opts.dataFile != "" {
		return opts, errors.New("use either -d/--data or --data-file, not both")
	}

	if opts.timeout <= 0 {
		return opts, errors.New("timeout must be greater than zero")
	}

	opts.url = flags.Arg(0)
	if !isValidURL(opts.url) {
		return opts, fmt.Errorf("invalid URL %q", opts.url)
	}

	if opts.method == "" {
		opts.method = http.MethodGet
		if opts.data != "" || opts.dataFile != "" {
			opts.method = http.MethodPost
		}
	}
	opts.method = strings.ToUpper(opts.method)

	return opts, nil
}

func usage(name string) string {
	return fmt.Sprintf("Usage: %s [flags] <url>", name)
}

func isValidURL(rawURL string) bool {
	u, err := url.ParseRequestURI(rawURL)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func requestBody(opts options) (io.Reader, error) {
	if opts.data != "" {
		return strings.NewReader(opts.data), nil
	}

	if opts.dataFile == "" {
		return nil, nil
	}

	data, err := os.ReadFile(opts.dataFile)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(data), nil
}

func newRequest(opts options, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(opts.method, opts.url, body)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "text/plain; charset=utf-8")
	}

	for _, raw := range opts.headers {
		name, value, ok := strings.Cut(raw, ":")
		if !ok || strings.TrimSpace(name) == "" {
			return nil, fmt.Errorf("invalid header %q; expected 'Name: value'", raw)
		}

		req.Header.Add(strings.TrimSpace(name), strings.TrimSpace(value))
	}

	return req, nil
}

func writeResponse(res *http.Response, opts options, stdout io.Writer) error {
	if opts.includeHead {
		if _, err := fmt.Fprintf(stdout, "HTTP/%d.%d %s\n", res.ProtoMajor, res.ProtoMinor, res.Status); err != nil {
			return err
		}

		if err := res.Header.Write(stdout); err != nil {
			return err
		}

		if _, err := fmt.Fprintln(stdout); err != nil {
			return err
		}
	}

	if opts.outputFile != "" {
		file, err := os.Create(opts.outputFile)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(file, res.Body)
		return err
	}

	_, err := io.Copy(stdout, res.Body)
	return err
}

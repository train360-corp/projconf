/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	converterURL = "https://converter.swagger.io/api/convert"
)

func main() {
	in := flag.String("in", "", "Input URL to Swagger/OpenAPI 2.0 JSON/YAML (e.g., http://127.0.0.1:54323/.../swagger.json)")
	out := flag.String("out", "openapi.yaml", "Output file path")
	format := flag.String("format", "yaml", "Output format: yaml|json (controls Accept header)")
	timeout := flag.Duration("timeout", 30*time.Second, "Total timeout")
	flag.Parse()

	if *in == "" {
		fmt.Fprintln(os.Stderr, "missing -in")
		os.Exit(2)
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	// 1) GET the source spec
	srcBody, err := httpGet(ctx, *in)
	if err != nil {
		die("fetch spec: %v", err)
	}

	// 2) POST to converter
	accept := "application/yaml"
	if *format == "json" {
		accept = "application/json"
	}
	converted, err := httpPost(ctx, converterURL, srcBody, accept)
	if err != nil {
		die("convert spec: %v", err)
	}

	// 3) Write result to disk
	if err := os.WriteFile(*out, converted, 0o644); err != nil {
		die("write output: %v", err)
	}

	fmt.Printf("Wrote %s (%d bytes)\n", *out, len(converted))
}

func httpGet(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GET %s: status %s: %s", url, resp.Status, string(b))
	}
	return io.ReadAll(resp.Body)
}

func httpPost(ctx context.Context, url string, body []byte, accept string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, io.NopCloser(io.NewSectionReader(
		// SectionReader wraps []byte without copy; but simplest is bytes.NewReader(body). Using SectionReader to avoid alloc.
		// if you prefer, replace with bytes.NewReader(body)
		newBytesReader(body), 0, int64(len(body)),
	)))
	if err != nil {
		return nil, err
	}
	// The converter accepts JSON or YAML payloads. If your source is JSON, this is fine.
	// If it's YAML, setting Content-Type to application/yaml also works.
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", accept)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("POST %s: status %s: %s", url, resp.Status, string(b))
	}
	return io.ReadAll(resp.Body)
}

// newBytesReader returns an io.ReaderAt over b without copying.
// You can replace all of httpPost's reader with bytes.NewReader(body) if you prefer simpler code.
type sliceAt []byte

func (s sliceAt) ReadAt(p []byte, off int64) (int, error) {
	if off >= int64(len(s)) {
		return 0, io.EOF
	}
	n := copy(p, s[off:])
	if n < len(p) {
		return n, io.EOF
	}
	return n, nil
}
func newBytesReader(b []byte) io.ReaderAt { return sliceAt(b) }

func die(f string, a ...any) {
	fmt.Fprintf(os.Stderr, f+"\n", a...)
	os.Exit(1)
}

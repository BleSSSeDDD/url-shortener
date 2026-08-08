//go:build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

// baseURL points at the running stack. CI sets it through BASE_URL (Traefik on :80);
// locally it defaults to hitting the server directly on :8080.
func baseURL() string {
	if v := strings.TrimRight(os.Getenv("BASE_URL"), "/"); v != "" {
		return v
	}
	return "http://localhost:8080"
}

// noRedirectClient does not follow 3xx so the redirect itself can be inspected
// (status and Location header) instead of leaving for an external site.
func noRedirectClient() *http.Client {
	return &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

type shortenResponse struct {
	ShortURL string `json:"short_url"`
	Code     string `json:"code"`
}

// shortenJSON shortens a url through the JSON API and returns the parsed response.
func shortenJSON(t *testing.T, target string) shortenResponse {
	t.Helper()

	body, err := json.Marshal(map[string]string{"url": target})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	resp, err := http.Post(baseURL()+"/api/v1/shorten", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /api/v1/shorten: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("POST /api/v1/shorten: expected 201, got %d: %s", resp.StatusCode, raw)
	}

	var out shortenResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.Code == "" {
		t.Fatal("response contains an empty code")
	}
	return out
}

func TestHealthEndpoints(t *testing.T) {
	t.Run("plain /health", func(t *testing.T) {
		resp, err := http.Get(baseURL() + "/health")
		if err != nil {
			t.Fatalf("GET /health: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
		raw, _ := io.ReadAll(resp.Body)
		if strings.TrimSpace(string(raw)) != "OK" {
			t.Errorf("expected body OK, got %q", raw)
		}
	})

	t.Run("json /api/v1/health", func(t *testing.T) {
		resp, err := http.Get(baseURL() + "/api/v1/health")
		if err != nil {
			t.Fatalf("GET /api/v1/health: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
		var out struct {
			Status string `json:"status"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if out.Status != "ok" {
			t.Errorf(`expected status "ok", got %q`, out.Status)
		}
	})
}

// TestShortenAndRedirect is the main scenario: shorten a link and verify that
// /r/{code} returns a 302 to the original url.
func TestShortenAndRedirect(t *testing.T) {
	const target = "https://example.com/e2e/redirect-check"

	res := shortenJSON(t, target)

	resp, err := noRedirectClient().Get(baseURL() + "/r/" + res.Code)
	if err != nil {
		t.Fatalf("GET /r/%s: %v", res.Code, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		t.Fatalf("expected 302, got %d", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != target {
		t.Errorf("Location: expected %q, got %q", target, loc)
	}
}

// TestShortenIsIdempotent: the same url must always yield the same code
// (ON CONFLICT (url) in the database).
func TestShortenIsIdempotent(t *testing.T) {
	const target = "https://example.com/e2e/idempotent-check"

	first := shortenJSON(t, target)
	second := shortenJSON(t, target)

	if first.Code != second.Code {
		t.Errorf("the same url produced different codes: %q and %q", first.Code, second.Code)
	}
}

// TestShortenRejectsBadInput: the server must answer 400 on malformed input.
func TestShortenRejectsBadInput(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{"empty url", `{"url":""}`},
		{"broken json", `{"url":`},
		{"javascript scheme", `{"url":"javascript:alert(1)"}`},
		{"no scheme", `{"url":"example.com"}`},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			resp, err := http.Post(baseURL()+"/api/v1/shorten", "application/json", strings.NewReader(c.body))
			if err != nil {
				t.Fatalf("POST /api/v1/shorten: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusBadRequest {
				t.Errorf("expected 400, got %d", resp.StatusCode)
			}
		})
	}
}

// TestRedirectUnknownCodeReturns404: an unknown code must give 404, not a redirect.
func TestRedirectUnknownCodeReturns404(t *testing.T) {
	resp, err := noRedirectClient().Get(baseURL() + "/r/e2e-definitely-missing")
	if err != nil {
		t.Fatalf("GET /r/...: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

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

// baseURL — адрес поднятого стека. В CI задаётся через BASE_URL (Traefik на :80),
// локально по умолчанию бьём напрямую в сервер на :8080.
func baseURL() string {
	if v := strings.TrimRight(os.Getenv("BASE_URL"), "/"); v != "" {
		return v
	}
	return "http://localhost:8080"
}

// noRedirectClient не следует за 3xx, чтобы можно было проверить сам редирект
// (статус и заголовок Location), а не улетать на внешний сайт.
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

// shortenJSON сокращает url через JSON API и возвращает разобранный ответ.
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
		t.Fatalf("POST /api/v1/shorten: ожидали 201, получили %d: %s", resp.StatusCode, raw)
	}

	var out shortenResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.Code == "" {
		t.Fatal("в ответе пустой code")
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
			t.Fatalf("ожидали 200, получили %d", resp.StatusCode)
		}
		raw, _ := io.ReadAll(resp.Body)
		if strings.TrimSpace(string(raw)) != "OK" {
			t.Errorf("ожидали тело OK, получили %q", raw)
		}
	})

	t.Run("json /api/v1/health", func(t *testing.T) {
		resp, err := http.Get(baseURL() + "/api/v1/health")
		if err != nil {
			t.Fatalf("GET /api/v1/health: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("ожидали 200, получили %d", resp.StatusCode)
		}
		var out struct {
			Status string `json:"status"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if out.Status != "ok" {
			t.Errorf(`ожидали status "ok", получили %q`, out.Status)
		}
	})
}

// TestShortenAndRedirect — основной сценарий: сократили ссылку и убедились, что
// /r/{code} отдаёт 302 на исходный url.
func TestShortenAndRedirect(t *testing.T) {
	const target = "https://example.com/e2e/redirect-check"

	res := shortenJSON(t, target)

	resp, err := noRedirectClient().Get(baseURL() + "/r/" + res.Code)
	if err != nil {
		t.Fatalf("GET /r/%s: %v", res.Code, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		t.Fatalf("ожидали 302, получили %d", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != target {
		t.Errorf("Location: ожидали %q, получили %q", target, loc)
	}
}

// TestShortenIsIdempotent — один и тот же url должен возвращать тот же код
// (ON CONFLICT (url) в БД).
func TestShortenIsIdempotent(t *testing.T) {
	const target = "https://example.com/e2e/idempotent-check"

	first := shortenJSON(t, target)
	second := shortenJSON(t, target)

	if first.Code != second.Code {
		t.Errorf("один и тот же url дал разные коды: %q и %q", first.Code, second.Code)
	}
}

// TestShortenRejectsBadInput — сервер должен отвечать 400 на некорректный ввод.
func TestShortenRejectsBadInput(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{"пустой url", `{"url":""}`},
		{"битый json", `{"url":`},
		{"схема javascript", `{"url":"javascript:alert(1)"}`},
		{"без схемы", `{"url":"example.com"}`},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			resp, err := http.Post(baseURL()+"/api/v1/shorten", "application/json", strings.NewReader(c.body))
			if err != nil {
				t.Fatalf("POST /api/v1/shorten: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusBadRequest {
				t.Errorf("ожидали 400, получили %d", resp.StatusCode)
			}
		})
	}
}

// TestRedirectUnknownCodeReturns404 — неизвестный код должен давать 404, а не редирект.
func TestRedirectUnknownCodeReturns404(t *testing.T) {
	resp, err := noRedirectClient().Get(baseURL() + "/r/e2e-definitely-missing")
	if err != nil {
		t.Fatalf("GET /r/...: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("ожидали 404, получили %d", resp.StatusCode)
	}
}

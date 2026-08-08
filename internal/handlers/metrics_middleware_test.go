package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/BleSSSeDDD/url-shortener/internal/metrics"
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

// TestMetricsMiddleware checks that the middleware counts requests by route
// TEMPLATE: two requests with different ids must land in the same counter
// (route="/test/{id}") instead of spawning one series per URL.
func TestMetricsMiddleware(t *testing.T) {
	r := chi.NewRouter()
	r.Use(Metrics)
	r.Get("/test/{id}", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	counter := metrics.HTTPRequestsTotal.WithLabelValues("GET", "/test/{id}", "200")
	before := testutil.ToFloat64(counter)

	for _, id := range []string{"123", "456"} {
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/test/"+id, nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	}

	if got := testutil.ToFloat64(counter) - before; got != 2 {
		t.Errorf("http_requests_total{route=/test/{id}}: expected +2, got +%v", got)
	}
}

// TestMetricsEndpoint checks that /metrics returns 200 and exposes our metrics.
func TestMetricsEndpoint(t *testing.T) {
	// make at least one observation, otherwise the metric is absent from the output
	metrics.CacheRequests.WithLabelValues("hit").Inc()
	metrics.HTTPRequestsTotal.WithLabelValues("GET", "/x", "200").Inc()

	rec := httptest.NewRecorder()
	promhttp.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 from /metrics, got %d", rec.Code)
	}
	body := rec.Body.String()
	for _, name := range []string{"http_requests_total", "urlshortener_cache_requests_total"} {
		if !strings.Contains(body, name) {
			t.Errorf("metric %q is missing from /metrics output", name)
		}
	}
}

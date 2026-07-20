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

// TestMetricsMiddleware проверяет, что middleware считает запросы по ШАБЛОНУ
// маршрута: два запроса с разными id должны попасть в один и тот же счётчик
// (route="/test/{id}"), а не расплодить серию на каждый URL.
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
			t.Fatalf("ожидали 200, получили %d", rec.Code)
		}
	}

	if got := testutil.ToFloat64(counter) - before; got != 2 {
		t.Errorf("http_requests_total{route=/test/{id}}: ожидали +2, получили +%v", got)
	}
}

// TestMetricsEndpoint проверяет, что /metrics отдаёт 200 и в теле есть наши метрики.
func TestMetricsEndpoint(t *testing.T) {
	// гарантируем хотя бы по одному наблюдению, иначе метрика не попадёт в вывод
	metrics.CacheRequests.WithLabelValues("hit").Inc()
	metrics.HTTPRequestsTotal.WithLabelValues("GET", "/x", "200").Inc()

	rec := httptest.NewRecorder()
	promhttp.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("ожидали 200 от /metrics, получили %d", rec.Code)
	}
	body := rec.Body.String()
	for _, name := range []string{"http_requests_total", "urlshortener_cache_requests_total"} {
		if !strings.Contains(body, name) {
			t.Errorf("в выводе /metrics нет метрики %q", name)
		}
	}
}

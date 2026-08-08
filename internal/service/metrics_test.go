package service

import (
	"context"
	"errors"
	"testing"

	"github.com/BleSSSeDDD/url-shortener/internal/metrics"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/mock"
)

// TestGetRecordsCacheHitMiss checks that Get increments the right counter:
// hit on a cache hit and miss when the value comes from the database.
func TestGetRecordsCacheHitMiss(t *testing.T) {
	hit := metrics.CacheRequests.WithLabelValues("hit")
	miss := metrics.CacheRequests.WithLabelValues("miss")
	hitBefore := testutil.ToFloat64(hit)
	missBefore := testutil.ToFloat64(miss)

	// --- cache hit: cacheGetter returned a value with no error ---
	mg := new(MockCacheGetter)
	mg.On("GetFromCache", mock.Anything, "abc").Return("https://example.com", nil)
	getter, _ := NewURLShortener(mg, new(MockCacheSetter), new(MockStorageGetter), new(MockStorageSetter))
	if _, err := getter.Get(context.Background(), "abc"); err != nil {
		t.Fatalf("hit: unexpected error %v", err)
	}
	if got := testutil.ToFloat64(hit) - hitBefore; got != 1 {
		t.Errorf("cache hit: expected +1, got +%v", got)
	}

	// --- cache miss: the cache errored, the value comes from the database ---
	mg2 := new(MockCacheGetter)
	mg2.On("GetFromCache", mock.Anything, "xyz").Return("", errors.New("cache miss"))
	ms2 := new(MockCacheSetter)
	ms2.On("AddToCache", mock.Anything, "xyz", "https://go.dev").Return(nil)
	sg2 := new(MockStorageGetter)
	sg2.On("GetURLFromCode", mock.Anything, "xyz").Return("https://go.dev", nil)
	getter2, _ := NewURLShortener(mg2, ms2, sg2, new(MockStorageSetter))
	if _, err := getter2.Get(context.Background(), "xyz"); err != nil {
		t.Fatalf("miss: unexpected error %v", err)
	}
	if got := testutil.ToFloat64(miss) - missBefore; got != 1 {
		t.Errorf("cache miss: expected +1, got +%v", got)
	}
}

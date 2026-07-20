package service

import (
	"context"
	"errors"
	"testing"

	"github.com/BleSSSeDDD/url-shortener/internal/metrics"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/mock"
)

// TestGetRecordsCacheHitMiss проверяет, что Get инкрементит нужный счётчик:
// hit при попадании в кэш и miss при промахе (когда идём в БД).
func TestGetRecordsCacheHitMiss(t *testing.T) {
	hit := metrics.CacheRequests.WithLabelValues("hit")
	miss := metrics.CacheRequests.WithLabelValues("miss")
	hitBefore := testutil.ToFloat64(hit)
	missBefore := testutil.ToFloat64(miss)

	// --- попадание в кэш: cacheGetter вернул значение без ошибки ---
	mg := new(MockCacheGetter)
	mg.On("GetFromCache", mock.Anything, "abc").Return("https://example.com", nil)
	getter, _ := NewURLShortener(mg, new(MockCacheSetter), new(MockStorageGetter), new(MockStorageSetter))
	if _, err := getter.Get(context.Background(), "abc"); err != nil {
		t.Fatalf("hit: неожиданная ошибка %v", err)
	}
	if got := testutil.ToFloat64(hit) - hitBefore; got != 1 {
		t.Errorf("cache hit: ожидали +1, получили +%v", got)
	}

	// --- промах: кэш вернул ошибку, значение достаём из БД ---
	mg2 := new(MockCacheGetter)
	mg2.On("GetFromCache", mock.Anything, "xyz").Return("", errors.New("cache miss"))
	ms2 := new(MockCacheSetter)
	ms2.On("AddToCache", mock.Anything, "xyz", "https://go.dev").Return(nil)
	sg2 := new(MockStorageGetter)
	sg2.On("GetURLFromCode", mock.Anything, "xyz").Return("https://go.dev", nil)
	getter2, _ := NewURLShortener(mg2, ms2, sg2, new(MockStorageSetter))
	if _, err := getter2.Get(context.Background(), "xyz"); err != nil {
		t.Fatalf("miss: неожиданная ошибка %v", err)
	}
	if got := testutil.ToFloat64(miss) - missBefore; got != 1 {
		t.Errorf("cache miss: ожидали +1, получили +%v", got)
	}
}

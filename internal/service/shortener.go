// Package service holds the URL shortening business logic.
package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math/big"

	"github.com/BleSSSeDDD/url-shortener/internal/handlers"
	"github.com/BleSSSeDDD/url-shortener/internal/metrics"
	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"
	"golang.org/x/sync/singleflight"
)

type CacheGetter interface {
	GetFromCache(ctx context.Context, code string) (string, error)
}

type CacheSetter interface {
	AddToCache(ctx context.Context, code string, url string) error
}

type StorageGetter interface {
	GetURLFromCode(ctx context.Context, code string) (originalURL string, err error)
}

type StorageSetter interface {
	SetNewPair(ctx context.Context, url string, code string) (string, error)
}

// Parameters used to build unique short codes.
const (
	CodeLength  = 6
	MaxAttempts = 10
	URLCharset  = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

type urlShortener struct {
	cacheGetter   CacheGetter
	cacheSetter   CacheSetter
	storageGetter StorageGetter
	storageSetter StorageSetter
	group         singleflight.Group
}

// NewURLShortener builds a URLShortener and returns a pointer to it.
func NewURLShortener(cacheGetter CacheGetter, cacheSetter CacheSetter, storageGetter StorageGetter, storageSetter StorageSetter) (handlers.URLShortenerGetter, handlers.URLShortenerSetter) {
	short := &urlShortener{cacheGetter: cacheGetter, cacheSetter: cacheSetter, storageGetter: storageGetter, storageSetter: storageSetter}
	return short, short
}

// Generates a random six-character string.
func generateShortenedURL() string {
	res := make([]byte, CodeLength)

	for i := 0; i < CodeLength; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(URLCharset))))
		if err != nil {
			log.Printf("failed to generate a cryptographically secure sequence: %v", err)
			res[i] = URLCharset[0]
			continue
		}
		res[i] = URLCharset[n.Int64()]
	}

	return string(res)
}

// Takes: the original URL.
//
// Returns: the short code and an error.
//
// Generates codes until one is unique, stores the pair and returns the code.
//
// If the URL is already stored, the existing code is returned instead.
func (u *urlShortener) Set(ctx context.Context, url string) (shortenedURL string, err error) {
	// Generate a code and try to store the pair. On a primary key collision
	// on code, generate another one until the attempts are exhausted.
	for i := 0; i < MaxAttempts; i++ {
		code := generateShortenedURL()

		resultCode, seterr := u.storageSetter.SetNewPair(ctx, url, code)
		if seterr == nil {
			// Success: either a new pair was inserted, or the url already existed and
			// ON CONFLICT (url) returned its existing code.
			metrics.URLsShortenedTotal.Inc()
			return resultCode, nil
		}

		var pgErr *pq.Error
		if errors.As(seterr, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			// The generated code collided with an existing one, so try again.
			continue
		}

		// Retrying will not fix any other error.
		return "", seterr
	}

	return "", fmt.Errorf("failed to generate a unique code in %d attempts", MaxAttempts)
}

// Returns the stored URL, or an empty string and an error when it is absent.
func (u *urlShortener) Get(ctx context.Context, shortCode string) (originalURL string, err error) {
	originalURL, err = u.cacheGetter.GetFromCache(ctx, shortCode)
	if err == nil {
		metrics.CacheRequests.WithLabelValues("hit").Inc()
		return originalURL, nil
	}
	metrics.CacheRequests.WithLabelValues("miss").Inc()

	code, singleflightError, _ := u.group.Do(shortCode, func() (any, error) {
		return u.storageGetter.GetURLFromCode(ctx, shortCode)
	})

	if singleflightError != nil {
		return "", singleflightError
	}

	originalURL, ok := code.(string)
	if !ok {
		return "", fmt.Errorf("could not assert the singleflight result to a string")
	}

	if addToCacheErr := u.cacheSetter.AddToCache(ctx, shortCode, originalURL); addToCacheErr != nil {
		log.Printf("failed to store the value in cache: %v", addToCacheErr)
	}

	return originalURL, nil
}

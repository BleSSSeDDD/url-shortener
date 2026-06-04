package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGenerateShortenedUrl(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "тест длины ссылки",
		},
		{
			name: "тест на чарсет",
		},
		{
			name: "тест на различность",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := generateShortenedURL()

			switch tc.name {
			case "тест длины ссылки":
				assert.Equal(t, CodeLength, len(got), fmt.Sprintf("ошибка в тесте %s: длина %d, ожидалось %d", tc.name, len(got), CodeLength))
			case "тест на чарсет":
				flag := true
				for _, r := range got {
					if !strings.Contains(URLCharset, string(r)) {
						flag = false
					}
				}
				assert.Equal(t, true, flag, fmt.Sprintf("ошибка в тесте %s: присутствует сивол, которого нет в CODE_CAHRSET", tc.name))
			case "тест на различность":
				got2 := generateShortenedURL()
				got3 := generateShortenedURL()
				assert.Equal(t, false, got == got2 && got2 == got3, fmt.Sprintf("ошибка в тесте %s: сненерировались одинаковые ссылки, ожидались разные", tc.name))
			}
		})
	}
}

type MockCache struct {
	mock.Mock
}

func (m *MockCache) GetFromCache(ctx context.Context, code string) (string, error) {
	args := m.Called(ctx, code)
	return args.String(0), args.Error(1)
}

func (m *MockCache) AddToCache(ctx context.Context, code string, url string) error {
	args := m.Called(ctx, code, url)
	return args.Error(0)
}

// Мок для storage.Postgres
type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) GetURLFromCode(ctx context.Context, code string) (string, error) {
	args := m.Called(ctx, code)
	return args.String(0), args.Error(1)
}

func (m *MockStorage) SetNewPair(ctx context.Context, url string, code string) (string, error) {
	args := m.Called(ctx, url, code)
	return args.String(0), args.Error(1)
}

func TestSetNewURL(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		mockCode  string
		mockError error
		wantError bool
	}{
		{
			name:      "успешное создание",
			url:       "https://example.com",
			mockCode:  "abc123",
			mockError: nil,
			wantError: false,
		},
		{
			name:      "ошибка БД",
			url:       "https://example.com",
			mockCode:  "",
			mockError: errors.New("db error"),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCache := new(MockCache)
			mockStorage := new(MockStorage)

			mockStorage.On("SetNewPair", mock.Anything, tt.url, mock.AnythingOfType("string")).
				Return(tt.mockCode, tt.mockError)

			shortener := NewURLShortener(mockCache, mockStorage)

			code, err := shortener.Set(context.Background(), tt.url)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.mockCode, code)
			}

			mockStorage.AssertExpectations(t)
		})
	}
}

func TestGetURL(t *testing.T) {
	tests := []struct {
		name           string
		code           string
		cacheResult    string
		cacheError     error
		storageResult  string
		storageError   error
		expectedURL    string
		expectCacheAdd bool
		wantError      bool
	}{
		{
			name:           "из кеша",
			code:           "abc123",
			cacheResult:    "https://example.com",
			cacheError:     nil,
			expectedURL:    "https://example.com",
			expectCacheAdd: false,
			wantError:      false,
		},
		{
			name:           "из БД (кеш пуст)",
			code:           "abc123",
			cacheResult:    "",
			cacheError:     errors.New("cache miss"),
			storageResult:  "https://example.com",
			storageError:   nil,
			expectedURL:    "https://example.com",
			expectCacheAdd: true,
			wantError:      false,
		},
		{
			name:           "не найдено",
			code:           "notfound",
			cacheResult:    "",
			cacheError:     errors.New("cache miss"),
			storageResult:  "",
			storageError:   errors.New("not found"),
			expectedURL:    "",
			expectCacheAdd: false,
			wantError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCache := new(MockCache)
			mockStorage := new(MockStorage)

			mockCache.On("GetFromCache", mock.Anything, tt.code).
				Return(tt.cacheResult, tt.cacheError)

			if tt.storageResult != "" || tt.storageError != nil {
				mockStorage.On("GetURLFromCode", mock.Anything, tt.code).
					Return(tt.storageResult, tt.storageError)
			}

			if tt.expectCacheAdd {
				mockCache.On("AddToCache", mock.Anything, tt.code, tt.storageResult).
					Return(nil)
			}

			shortener := NewURLShortener(mockCache, mockStorage)

			url, err := shortener.Get(context.Background(), tt.code)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedURL, url)
			}

			mockCache.AssertExpectations(t)
			mockStorage.AssertExpectations(t)
		})
	}
}

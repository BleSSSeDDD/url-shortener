package service

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ============================================
// 1. Tests for generateShortenedURL
// ============================================

func TestGenerateShortenedUrl(t *testing.T) {
	tests := []struct {
		name  string
		check func(t *testing.T, got string)
	}{
		{
			name: "code length",
			check: func(t *testing.T, got string) {
				assert.Equal(t, CodeLength, len(got))
			},
		},
		{
			name: "characters from the charset",
			check: func(t *testing.T, got string) {
				for _, r := range got {
					assert.Contains(t, URLCharset, string(r))
				}
			},
		},
		{
			name: "codes differ",
			check: func(t *testing.T, got string) {
				got2 := generateShortenedURL()
				got3 := generateShortenedURL()
				assert.False(t, got == got2 && got2 == got3)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.check(t, generateShortenedURL())
		})
	}
}

// ============================================
// 2. Mocks
// ============================================

type MockCacheGetter struct {
	mock.Mock
}

func (m *MockCacheGetter) GetFromCache(ctx context.Context, code string) (string, error) {
	args := m.Called(ctx, code)
	return args.String(0), args.Error(1)
}

type MockCacheSetter struct {
	mock.Mock
}

func (m *MockCacheSetter) AddToCache(ctx context.Context, code string, url string) error {
	args := m.Called(ctx, code, url)
	return args.Error(0)
}

type MockStorageGetter struct {
	mock.Mock
}

func (m *MockStorageGetter) GetURLFromCode(ctx context.Context, code string) (string, error) {
	args := m.Called(ctx, code)
	return args.String(0), args.Error(1)
}

type MockStorageSetter struct {
	mock.Mock
}

func (m *MockStorageSetter) SetNewPair(ctx context.Context, url string, code string) (string, error) {
	args := m.Called(ctx, url, code)
	return args.String(0), args.Error(1)
}

// ============================================
// 3. Tests for Set
// ============================================

func TestSetNewURL(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		mockCode  string
		mockError error
		wantError bool
	}{
		{
			name:      "created successfully",
			url:       "https://example.com",
			mockCode:  "abc123",
			mockError: nil,
			wantError: false,
		},
		{
			name:      "database error",
			url:       "https://example.com",
			mockCode:  "",
			mockError: errors.New("db error"),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorageSetter := new(MockStorageSetter)

			mockStorageSetter.On("SetNewPair", mock.Anything, tt.url, mock.AnythingOfType("string")).
				Return(tt.mockCode, tt.mockError)

			// Stub out the remaining dependencies
			mockGetter := new(MockCacheGetter)
			mockSetter := new(MockCacheSetter)
			mockStorageGetter := new(MockStorageGetter)

			_, setter := NewURLShortener(
				mockGetter,
				mockSetter,
				mockStorageGetter,
				mockStorageSetter,
			)

			code, err := setter.Set(context.Background(), tt.url)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.mockCode, code)
			}

			mockStorageSetter.AssertExpectations(t)
		})
	}
}

// TestSetCodeCollisionRetries: on a generated code collision Set must produce
// a new code and retry instead of returning an empty string.
func TestSetCodeCollisionRetries(t *testing.T) {
	mockStorageSetter := new(MockStorageSetter)
	pgErr := &pq.Error{Code: pgerrcode.UniqueViolation}

	// The first attempt collides on the code primary key, the second succeeds.
	mockStorageSetter.On("SetNewPair", mock.Anything, "https://example.com", mock.AnythingOfType("string")).
		Return("", pgErr).Once()
	mockStorageSetter.On("SetNewPair", mock.Anything, "https://example.com", mock.AnythingOfType("string")).
		Return("abc123", nil).Once()

	_, setter := NewURLShortener(
		new(MockCacheGetter),
		new(MockCacheSetter),
		new(MockStorageGetter),
		mockStorageSetter,
	)

	code, err := setter.Set(context.Background(), "https://example.com")

	assert.NoError(t, err)
	assert.Equal(t, "abc123", code)
	mockStorageSetter.AssertNumberOfCalls(t, "SetNewPair", 2)
	mockStorageSetter.AssertExpectations(t)
}

// TestSetCodeCollisionExhausted: when every attempt collides Set must return
// an error rather than an empty code with a nil error.
func TestSetCodeCollisionExhausted(t *testing.T) {
	mockStorageSetter := new(MockStorageSetter)
	pgErr := &pq.Error{Code: pgerrcode.UniqueViolation}

	mockStorageSetter.On("SetNewPair", mock.Anything, "https://example.com", mock.AnythingOfType("string")).
		Return("", pgErr)

	_, setter := NewURLShortener(
		new(MockCacheGetter),
		new(MockCacheSetter),
		new(MockStorageGetter),
		mockStorageSetter,
	)

	code, err := setter.Set(context.Background(), "https://example.com")

	assert.Error(t, err)
	assert.Empty(t, code)
	mockStorageSetter.AssertNumberOfCalls(t, "SetNewPair", MaxAttempts)
}

// ============================================
// 4. Tests for Get
// ============================================

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
			name:           "from cache",
			code:           "abc123",
			cacheResult:    "https://example.com",
			cacheError:     nil,
			expectedURL:    "https://example.com",
			expectCacheAdd: false,
			wantError:      false,
		},
		{
			name:           "from database (cache miss)",
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
			name:           "not found",
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
			mockGetter := new(MockCacheGetter)
			mockSetter := new(MockCacheSetter)
			mockStorageGetter := new(MockStorageGetter)
			mockStorageSetter := new(MockStorageSetter)

			mockGetter.On("GetFromCache", mock.Anything, tt.code).
				Return(tt.cacheResult, tt.cacheError)

			if tt.storageResult != "" || tt.storageError != nil {
				mockStorageGetter.On("GetURLFromCode", mock.Anything, tt.code).
					Return(tt.storageResult, tt.storageError)
			}

			if tt.expectCacheAdd {
				mockSetter.On("AddToCache", mock.Anything, tt.code, tt.storageResult).
					Return(nil)
			}

			getter, _ := NewURLShortener(
				mockGetter,
				mockSetter,
				mockStorageGetter,
				mockStorageSetter,
			)

			url, err := getter.Get(context.Background(), tt.code)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedURL, url)
			}

			mockGetter.AssertExpectations(t)
			mockStorageGetter.AssertExpectations(t)

			if tt.expectCacheAdd {
				mockSetter.AssertExpectations(t)
			}
		})
	}
}

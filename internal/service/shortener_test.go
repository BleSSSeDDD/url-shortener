package service

import (
	"strings"
	"testing"
)

func TestGenerateShortenedUrl(t *testing.T) {
	result := generateShortenedUrl()

	if len(result) != 4 {
		t.Errorf("Expected length 4, got %d", len(result))
	}

	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	for _, char := range result {
		if !strings.Contains(charset, string(char)) {
			t.Errorf("Invalid character %c in result", char)
		}
	}
}

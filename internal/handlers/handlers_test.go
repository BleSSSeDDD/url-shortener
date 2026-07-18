package handlers

import (
	"strings"
	"testing"
)

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		wantErr bool
	}{
		{"обычный https", "https://example.com/path?x=1", false},
		{"обычный http", "http://example.com", false},
		{"без схемы", "example.com", true},
		{"недопустимая схема javascript", "javascript:alert(1)", true},
		{"недопустимая схема ftp", "ftp://example.com", true},
		{"схема без хоста", "http://", true},
		{"мусор", "not a url", true},
		{"слишком длинный", "https://example.com/" + strings.Repeat("a", maxURLLength), true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateURL(tc.raw)
			if tc.wantErr && err == nil {
				t.Errorf("validateURL(%q): ожидалась ошибка, получили nil", tc.raw)
			}
			if !tc.wantErr && err != nil {
				t.Errorf("validateURL(%q): не ожидали ошибку, получили %v", tc.raw, err)
			}
		})
	}
}

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
		{"plain https", "https://example.com/path?x=1", false},
		{"plain http", "http://example.com", false},
		{"no scheme", "example.com", true},
		{"disallowed scheme javascript", "javascript:alert(1)", true},
		{"disallowed scheme ftp", "ftp://example.com", true},
		{"scheme without host", "http://", true},
		{"garbage", "not a url", true},
		{"too long", "https://example.com/" + strings.Repeat("a", maxURLLength), true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateURL(tc.raw)
			if tc.wantErr && err == nil {
				t.Errorf("validateURL(%q): expected an error, got nil", tc.raw)
			}
			if !tc.wantErr && err != nil {
				t.Errorf("validateURL(%q): unexpected error: %v", tc.raw, err)
			}
		})
	}
}

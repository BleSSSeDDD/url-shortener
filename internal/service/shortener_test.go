package service

import (
	"strings"
	"testing"
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

			got := generateShortenedUrl()

			switch tc.name {
			case "тест длины ссылки":
				if len(got) != CODE_LENGTH {
					t.Errorf("ошибка в тесте %s: длина %d, ожидалось %d", tc.name, len(got), CODE_LENGTH)
				}
			case "тест на чарсет":
				for _, r := range got {
					if !strings.Contains(URL_CHARSET, string(r)) {
						t.Errorf("ошибка в тесте %s: присутствует сивол %q, которого нет в CODE_CAHRSET", tc.name, r)
					}
				}
			case "тест на различность":
				got2 := generateShortenedUrl()
				got3 := generateShortenedUrl()
				if got == got2 && got2 == got3 {
					t.Errorf("ошибка в тесте %s: сненерировались одинаковые ссылки, ожидались разные", tc.name)
				}
			}
		})
	}
}

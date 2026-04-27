package service

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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
				assert.Equal(t, CODE_LENGTH, len(got), fmt.Sprintf("ошибка в тесте %s: длина %d, ожидалось %d", tc.name, len(got), CODE_LENGTH))
			case "тест на чарсет":
				flag := true
				for _, r := range got {
					if !strings.Contains(URL_CHARSET, string(r)) {
						flag = false
					}
				}
				assert.Equal(t, true, flag, fmt.Sprintf("ошибка в тесте %s: присутствует сивол, которого нет в CODE_CAHRSET", tc.name))
			case "тест на различность":
				got2 := generateShortenedUrl()
				got3 := generateShortenedUrl()
				assert.Equal(t, false, got == got2 && got2 == got3, fmt.Sprintf("ошибка в тесте %s: сненерировались одинаковые ссылки, ожидались разные", tc.name))
			}
		})
	}
}

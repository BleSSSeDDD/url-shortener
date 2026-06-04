// Package handlers пакет для структур, используемых в апи
package handlers

// EndpointInfo структура для описания эндппоинтов
type EndpointInfo struct {
	Path        string `json:"path"`
	Method      string `json:"method"`
	Description string `json:"description"`
}

// APIV1Response структура для ответа APIV1
type APIV1Response struct {
	Version   string         `json:"version"`
	Status    string         `json:"status"`
	Endpoints []EndpointInfo `json:"endpoints"`
}

// APIResponse структура для ответов от апи
type APIResponse struct {
	Service       string   `json:"service"`
	Versions      []string `json:"versions"`
	Latest        string   `json:"latest"`
	Documentation string   `json:"documentation"`
}

// HealthResponse структура для ответа ручки health в апи
type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	Version string `json:"version"`
}

// ShortenResponse структура для ответа сокращенной ссылкой
type ShortenResponse struct {
	ShortURL string `json:"short_url"`
	Code     string `json:"code"`
}

// ShortenRequest структура для запроса на сокращение ссылки
type ShortenRequest struct {
	URL string `json:"url"`
}

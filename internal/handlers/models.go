// Package handlers also defines the structures used by the API.
package handlers

// EndpointInfo describes a single API endpoint.
type EndpointInfo struct {
	Path        string `json:"path"`
	Method      string `json:"method"`
	Description string `json:"description"`
}

// APIV1Response is the payload returned by the v1 API index.
type APIV1Response struct {
	Version   string         `json:"version"`
	Status    string         `json:"status"`
	Endpoints []EndpointInfo `json:"endpoints"`
}

// APIResponse is the generic API response payload.
type APIResponse struct {
	Service       string   `json:"service"`
	Versions      []string `json:"versions"`
	Latest        string   `json:"latest"`
	Documentation string   `json:"documentation"`
}

// HealthResponse is the payload returned by the health endpoint.
type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	Version string `json:"version"`
}

// ShortenResponse carries a freshly created short code.
type ShortenResponse struct {
	ShortURL string `json:"short_url"`
	Code     string `json:"code"`
}

// ShortenRequest is the body of a shorten request.
type ShortenRequest struct {
	URL string `json:"url"`
}

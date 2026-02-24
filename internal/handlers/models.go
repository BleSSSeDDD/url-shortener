package handlers

type EndpointInfo struct {
	Path        string `json:"path"`
	Method      string `json:"method"`
	Description string `json:"description"`
}

type APIV1Response struct {
	Version   string         `json:"version"`
	Status    string         `json:"status"`
	Endpoints []EndpointInfo `json:"endpoints"`
}

type APIResponse struct {
	Service       string   `json:"service"`
	Versions      []string `json:"versions"`
	Latest        string   `json:"latest"`
	Documentation string   `json:"documentation"`
}

type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	Version string `json:"version"`
}

type ShortenResponse struct {
	ShortURL string `json:"short_url"`
	Code     string `json:"code"`
}

type ShortenRequest struct {
	URL string `json:"url"`
}

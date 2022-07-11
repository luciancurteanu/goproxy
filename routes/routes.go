package routes

import "net/http"

// Clients is a simple container for two http.Clients
type Clients struct {
	Normal *http.Client
	Proxy  *http.Client
}

// CustomRequest is used by the Custom route to configure
// a HTTP request
type CustomRequest struct {
	URL     string
	Method  string
	Headers map[string]string
	Cookies map[string]string
	Body    string
}

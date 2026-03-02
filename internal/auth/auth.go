package auth

import "net/http"

// Authenticator applies authentication to an HTTP request.
type Authenticator interface {
	Apply(req *http.Request) error
}

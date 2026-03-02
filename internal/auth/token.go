package auth

import (
	"fmt"
	"net/http"
)

// TokenAuth implements Authenticator using an API key.
type TokenAuth struct {
	Token string
}

func NewTokenAuth(token string) *TokenAuth {
	return &TokenAuth{Token: token}
}

func (t *TokenAuth) Apply(req *http.Request) error {
	if t.Token == "" {
		return fmt.Errorf("no API key configured. Run 'nusii auth login' or set NUSII_API_KEY")
	}
	req.Header.Set("Authorization", "Token token="+t.Token)
	return nil
}

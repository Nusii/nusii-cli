package auth

import (
	"encoding/base64"
	"testing"
)

func TestGenerateCodeVerifier(t *testing.T) {
	v, err := GenerateCodeVerifier()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 32 bytes → 43 base64url chars (no padding)
	if len(v) != 43 {
		t.Errorf("expected verifier length 43, got %d", len(v))
	}
	// Must be valid base64url
	if _, err := base64.RawURLEncoding.DecodeString(v); err != nil {
		t.Errorf("verifier is not valid base64url: %v", err)
	}
}

func TestGenerateCodeVerifierUniqueness(t *testing.T) {
	v1, _ := GenerateCodeVerifier()
	v2, _ := GenerateCodeVerifier()
	if v1 == v2 {
		t.Error("two verifiers should not be equal")
	}
}

func TestGenerateCodeChallenge(t *testing.T) {
	// Known test vector from RFC 7636 Appendix B
	verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	challenge := GenerateCodeChallenge(verifier)
	expected := "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"
	if challenge != expected {
		t.Errorf("expected challenge %s, got %s", expected, challenge)
	}
}

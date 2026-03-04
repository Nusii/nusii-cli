package auth

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"
)

const callbackAddr = "127.0.0.1:18192"

// CallbackResult holds the authorization code or an error from the OAuth callback.
type CallbackResult struct {
	Code  string
	Error string
}

// RedirectURI returns the full redirect URI for OAuth configuration.
func RedirectURI() string {
	return "http://" + callbackAddr + "/callback"
}

// StartCallbackServer starts a local HTTP server that listens for the OAuth callback.
// It validates the state parameter, extracts the authorization code, and returns the
// result on the channel. The server shuts down after receiving a callback or after the timeout.
func StartCallbackServer(expectedState string, timeout time.Duration) (<-chan CallbackResult, error) {
	resultCh := make(chan CallbackResult, 1)

	mux := http.NewServeMux()

	server := &http.Server{
		Handler: mux,
	}

	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		// Check for error from authorization server
		if errParam := r.URL.Query().Get("error"); errParam != "" {
			desc := r.URL.Query().Get("error_description")
			if desc == "" {
				desc = errParam
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, errorHTML, desc)
			resultCh <- CallbackResult{Error: desc}
			go server.Shutdown(context.Background())
			return
		}

		// Validate state
		state := r.URL.Query().Get("state")
		if state != expectedState {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, errorHTML, "Invalid state parameter. Please try again.")
			resultCh <- CallbackResult{Error: "state mismatch"}
			go server.Shutdown(context.Background())
			return
		}

		code := r.URL.Query().Get("code")
		if code == "" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, errorHTML, "No authorization code received.")
			resultCh <- CallbackResult{Error: "no code"}
			go server.Shutdown(context.Background())
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, successHTML)
		resultCh <- CallbackResult{Code: code}
		go server.Shutdown(context.Background())
	})

	listener, err := net.Listen("tcp", callbackAddr)
	if err != nil {
		return nil, fmt.Errorf("starting callback server: %w", err)
	}

	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			resultCh <- CallbackResult{Error: fmt.Sprintf("callback server error: %v", err)}
		}
	}()

	// Timeout
	go func() {
		time.Sleep(timeout)
		server.Shutdown(context.Background())
		select {
		case resultCh <- CallbackResult{Error: "timed out waiting for authorization"}:
		default:
		}
	}()

	return resultCh, nil
}

const successHTML = `<!DOCTYPE html>
<html><head><title>Nusii CLI</title>
<style>body{font-family:system-ui,sans-serif;display:flex;justify-content:center;align-items:center;min-height:100vh;margin:0;background:#f8f9fa}
.card{text-align:center;padding:2rem;background:white;border-radius:12px;box-shadow:0 2px 8px rgba(0,0,0,.1)}
h1{color:#22c55e;margin:0 0 .5rem}p{color:#666}</style></head>
<body><div class="card"><h1>Authenticated!</h1><p>You can close this window and return to the terminal.</p></div></body></html>`

const errorHTML = `<!DOCTYPE html>
<html><head><title>Nusii CLI - Error</title>
<style>body{font-family:system-ui,sans-serif;display:flex;justify-content:center;align-items:center;min-height:100vh;margin:0;background:#f8f9fa}
.card{text-align:center;padding:2rem;background:white;border-radius:12px;box-shadow:0 2px 8px rgba(0,0,0,.1)}
h1{color:#ef4444;margin:0 0 .5rem}p{color:#666}</style></head>
<body><div class="card"><h1>Authentication Failed</h1><p>%s</p></div></body></html>`

package handlers

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/shanehull/obsidian-remote/internal/config"
)

func HandleDiscovery(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		scheme := "http"
		if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}
		host := fmt.Sprintf("%s://%s", scheme, r.Host)

		resourcePath := strings.TrimPrefix(r.URL.Path, "/.well-known/oauth-protected-resource")
		if resourcePath == "" {
			resourcePath = "/mcp"
		}

		slog.Info("discovery request", "host", host, "path", r.URL.Path, "resource_path", resourcePath)

		res := map[string]interface{}{
			"resource":                 fmt.Sprintf("%s%s", strings.TrimSuffix(host, "/"), resourcePath),
			"resource_name":            "obsidian-remote",
			"authorization_servers":    []string{host}, // Tell the CLI to look at us for the login URLs
			"client_id":                cfg.OAuthAudience,
			"clientId":                 cfg.OAuthAudience,
			"bearer_methods_supported": []string{"header"},
		}

		if err := json.NewEncoder(w).Encode(res); err != nil {
			slog.Error("failed to encode discovery response", "error", err)
		}
	}
}

func HandleAuthServerDiscovery(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		scheme := "http"
		if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}
		host := fmt.Sprintf("%s://%s", scheme, r.Host)

		// This is the RFC 8414 / OpenID Configuration the CLI is looking for
		res := map[string]interface{}{
			"issuer":                           host,
			"authorization_endpoint":           fmt.Sprintf("%s/authorize", host),
			"token_endpoint":                   fmt.Sprintf("%s/token", host),
			"registration_endpoint":            fmt.Sprintf("%s/register", host), // Add this!
			"jwks_uri":                         cfg.OAuthJwksURL,
			"response_types_supported":         []string{"code"},
			"grant_types_supported":            []string{"authorization_code", "refresh_token"},
			"scopes_supported":                 []string{"openid", "email", "profile"},
			"code_challenge_methods_supported": []string{"S256"},
		}

		if err := json.NewEncoder(w).Encode(res); err != nil {
			slog.Error("failed to encode auth server discovery response", "error", err)
		}
	}
}

func HandleRegistration(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var req map[string]interface{}
		if r.Body != nil {
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				slog.Warn("registration: failed to decode body", "error", err)
			}
		}

		redirectURIs, _ := req["redirect_uris"].([]interface{})
		if redirectURIs == nil {
			redirectURIs = []interface{}{"http://localhost"}
		}

		scope, _ := req["scope"].(string)
		if scope == "" {
			scope = "openid email profile"
		}

		res := map[string]interface{}{
			"client_id":                  cfg.OAuthAudience,
			"client_id_issued_at":        0,
			"client_secret_expires_at":   0,
			"redirect_uris":              redirectURIs,
			"scope":                      scope,
			"token_endpoint_auth_method": "none",
		}

		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(res); err != nil {
			slog.Error("failed to encode registration response", "error", err)
		}
	}
}

func HandleAuthorizeProxy(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()
		if params.Get("scope") == "" {
			params.Set("scope", "openid email profile")
		}
		// Google issues a refresh token only for offline access, and only repeats
		// it when consent is explicit. Without both, a login cannot outlive its
		// access token.
		if params.Get("access_type") == "" {
			params.Set("access_type", "offline")
		}
		if params.Get("prompt") == "" {
			params.Set("prompt", "consent")
		}

		// Route the response through our own callback and stash the client's
		// redirect_uri so the upstream IdP never redirects to the client
		// directly. The IdP stamps its own iss there, which the MCP client
		// rejects as an issuer mismatch (RFC 9207).
		state, err := encodeState(cfg, params.Get("redirect_uri"), params.Get("state"))
		if err != nil {
			http.Error(w, "missing or invalid redirect_uri", http.StatusBadRequest)
			return
		}

		params.Set("redirect_uri", callbackURL(cfg, r))
		params.Set("state", state)

		target := cfg.OAuthAuthorizeURL + "?" + params.Encode()
		http.Redirect(w, r, target, http.StatusFound)
	}
}

func HandleAuthorizeCallback(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()

		redirectURI, clientState, err := decodeState(cfg, params.Get("state"))
		if err != nil {
			slog.Warn("authorize callback: invalid state", "error", err)
			http.Error(w, "Invalid or expired OAuth state", http.StatusBadRequest)
			return
		}

		target, err := url.Parse(redirectURI)
		if err != nil {
			http.Error(w, "Invalid redirect_uri", http.StatusBadRequest)
			return
		}

		query := target.Query()
		if code := params.Get("code"); code != "" {
			query.Set("code", code)
		} else {
			if e := params.Get("error"); e != "" {
				query.Set("error", e)
			} else {
				query.Set("error", "access_denied")
			}
			if desc := params.Get("error_description"); desc != "" {
				query.Set("error_description", desc)
			}
		}
		if clientState != "" {
			query.Set("state", clientState)
		}
		target.RawQuery = query.Encode()

		http.Redirect(w, r, target.String(), http.StatusFound)
	}
}

func callbackURL(cfg *config.Config, r *http.Request) string {
	if cfg.OAuthCallbackURL != "" {
		return cfg.OAuthCallbackURL
	}
	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s/oauth/callback", scheme, r.Host)
}

type oauthState struct {
	RedirectURI string `json:"r"`
	ClientState string `json:"s"`
}

// State is signed when a client secret is available so the callback cannot be
// abused as an open redirect. Without one, a per-process key keeps the proxy
// working in unauthenticated configurations.
var stateSigningKey = randomSigningKey()

func randomSigningKey() []byte {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		slog.Error("failed to generate state signing key", "error", err)
	}
	return key
}

func encodeState(cfg *config.Config, redirectURI, clientState string) (string, error) {
	if redirectURI == "" {
		return "", fmt.Errorf("missing redirect_uri")
	}

	payload, err := json.Marshal(oauthState{RedirectURI: redirectURI, ClientState: clientState})
	if err != nil {
		return "", err
	}

	encoded := base64.RawURLEncoding.EncodeToString(payload)
	return encoded + "." + signState(cfg, encoded), nil
}

func decodeState(cfg *config.Config, state string) (string, string, error) {
	encoded, signature, ok := strings.Cut(state, ".")
	if !ok {
		return "", "", fmt.Errorf("malformed state")
	}

	expected := signState(cfg, encoded)
	if !hmac.Equal([]byte(expected), []byte(signature)) {
		return "", "", fmt.Errorf("invalid state signature")
	}

	payload, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return "", "", err
	}

	var decoded oauthState
	if err := json.Unmarshal(payload, &decoded); err != nil {
		return "", "", err
	}
	if decoded.RedirectURI == "" {
		return "", "", fmt.Errorf("missing redirect_uri")
	}
	return decoded.RedirectURI, decoded.ClientState, nil
}

func signState(cfg *config.Config, encoded string) string {
	key := stateSigningKey
	if cfg.OAuthClientSecret != "" {
		key = []byte(cfg.OAuthClientSecret)
	}

	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(encoded))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func HandleTokenProxy(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		form := url.Values{}
		for k, v := range r.Form {
			form[k] = v
		}
		form.Set("client_id", cfg.OAuthAudience)
		if cfg.OAuthClientSecret != "" {
			form.Set("client_secret", cfg.OAuthClientSecret)
		}
		// The code was issued against our callback, so the exchange must
		// present the same redirect_uri rather than the client's.
		if form.Get("grant_type") == "authorization_code" {
			form.Set("redirect_uri", callbackURL(cfg, r))
		}

		resp, err := http.PostForm(cfg.OAuthTokenURL, form)
		if err != nil {
			slog.Error("token proxy: upstream request failed", "error", err)
			http.Error(w, "Token exchange failed", http.StatusBadGateway)
			return
		}
		defer func() { _ = resp.Body.Close() }()

		w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
		w.WriteHeader(resp.StatusCode)
		if _, err := io.Copy(w, resp.Body); err != nil {
			slog.Error("token proxy: failed to write response", "error", err)
		}
	}
}

func HandleConfig(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		res := map[string]string{
			"type":     "oauth",
			"issuer":   cfg.OAuthIssuer,
			"clientId": cfg.OAuthAudience,
		}

		if err := json.NewEncoder(w).Encode(res); err != nil {
			slog.Error("failed to encode config response", "error", err)
		}
	}
}

func HandleHealthz(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		host := cfg.ObsidianURL
		host = strings.TrimPrefix(host, "http://")
		host = strings.TrimPrefix(host, "https://")

		conn, err := net.DialTimeout("tcp", host, 500*time.Millisecond)
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"status":"unavailable"}`))
			return
		}
		_ = conn.Close()

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}
}

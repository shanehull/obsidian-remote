package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

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
			"resource":              fmt.Sprintf("%s%s", strings.TrimSuffix(host, "/"), resourcePath),
			"resource_name":         "obsidian-remote",
			"authorization_servers": []string{host}, // Tell the CLI to look at us for the login URLs
			"client_id":             cfg.OAuthAudience,
			"clientId":              cfg.OAuthAudience,
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
			"issuer":                 host,
			"authorization_endpoint": fmt.Sprintf("%s/authorize", host),
			"token_endpoint":         fmt.Sprintf("%s/token", host),
			"registration_endpoint":  fmt.Sprintf("%s/register", host), // Add this!
			"jwks_uri":               cfg.OAuthJwksURL,
			"response_types_supported":               []string{"code"},
			"grant_types_supported":                  []string{"authorization_code", "refresh_token"},
			"scopes_supported":                       []string{"openid", "email", "profile"},
			"code_challenge_methods_supported":        []string{"S256"},
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
			_ = json.NewDecoder(r.Body).Decode(&req)
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
			"redirect_uris":             redirectURIs,
			"scope":                     scope,
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
		target := cfg.OAuthAuthorizeURL + "?" + params.Encode()
		http.Redirect(w, r, target, http.StatusFound)
	}
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

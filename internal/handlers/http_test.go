package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/shanehull/obsidian-remote/internal/config"
	"github.com/shanehull/obsidian-remote/internal/handlers"
)

func testConfig() *config.Config {
	return &config.Config{
		OAuthIssuer:       "https://accounts.google.com",
		OAuthJwksURL:      "https://www.googleapis.com/oauth2/v3/certs",
		OAuthAuthorizeURL: "https://accounts.google.com/o/oauth2/v2/auth",
		OAuthTokenURL:     "https://oauth2.googleapis.com/token",
		OAuthAudience:     "test-client-id",
		OAuthClientSecret: "test-secret",
	}
}

func TestHandleDiscovery(t *testing.T) {
	cfg := testConfig()
	h := handlers.HandleDiscovery(cfg)

	req := httptest.NewRequest("GET", "/.well-known/oauth-protected-resource/mcp", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var body map[string]interface{}
	_ = json.NewDecoder(rec.Body).Decode(&body)

	if body["resource_name"] != "obsidian-remote" {
		t.Errorf("expected resource_name=obsidian-remote, got %v", body["resource_name"])
	}
	if body["client_id"] != "test-client-id" {
		t.Errorf("expected client_id=test-client-id, got %v", body["client_id"])
	}
	if _, ok := body["authorization_servers"]; !ok {
		t.Error("missing authorization_servers")
	}
}

func TestHandleDiscovery_RootPath(t *testing.T) {
	cfg := testConfig()
	h := handlers.HandleDiscovery(cfg)

	req := httptest.NewRequest("GET", "/.well-known/oauth-protected-resource", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	var body map[string]interface{}
	_ = json.NewDecoder(rec.Body).Decode(&body)

	resource, _ := body["resource"].(string)
	if !strings.HasSuffix(resource, "/mcp") {
		t.Errorf("expected resource ending in /mcp, got %s", resource)
	}
}

func TestHandleAuthServerDiscovery(t *testing.T) {
	cfg := testConfig()
	h := handlers.HandleAuthServerDiscovery(cfg)

	req := httptest.NewRequest("GET", "/.well-known/oauth-authorization-server", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var body map[string]interface{}
	_ = json.NewDecoder(rec.Body).Decode(&body)

	if _, ok := body["registration_endpoint"]; !ok {
		t.Error("missing registration_endpoint")
	}
	if _, ok := body["authorization_endpoint"]; !ok {
		t.Error("missing authorization_endpoint")
	}
	if _, ok := body["token_endpoint"]; !ok {
		t.Error("missing token_endpoint")
	}
}

func TestHandleRegistration(t *testing.T) {
	cfg := testConfig()
	h := handlers.HandleRegistration(cfg)

	body := strings.NewReader(`{"redirect_uris":["http://localhost:3000"],"scope":"openid"}`)
	req := httptest.NewRequest("POST", "/register", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}

	var res map[string]interface{}
	_ = json.NewDecoder(rec.Body).Decode(&res)

	if res["client_id"] != "test-client-id" {
		t.Errorf("expected client_id=test-client-id, got %v", res["client_id"])
	}
	if res["token_endpoint_auth_method"] != "none" {
		t.Errorf("expected token_endpoint_auth_method=none, got %v", res["token_endpoint_auth_method"])
	}
}

func TestHandleRegistration_Defaults(t *testing.T) {
	cfg := testConfig()
	h := handlers.HandleRegistration(cfg)

	req := httptest.NewRequest("POST", "/register", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}

	var res map[string]interface{}
	_ = json.NewDecoder(rec.Body).Decode(&res)

	uris, _ := res["redirect_uris"].([]interface{})
	if len(uris) != 1 || uris[0] != "http://localhost" {
		t.Errorf("expected default redirect_uris, got %v", uris)
	}
	if res["scope"] != "openid email profile" {
		t.Errorf("expected default scope, got %v", res["scope"])
	}
}

func TestHandleAuthorizeProxy_InjectScope(t *testing.T) {
	cfg := testConfig()
	h := handlers.HandleAuthorizeProxy(cfg)

	req := httptest.NewRequest("GET", "/authorize?client_id=x&redirect_uri=http://localhost", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf("expected 302, got %d", rec.Code)
	}
	loc := rec.Header().Get("Location")
	if !strings.HasPrefix(loc, cfg.OAuthAuthorizeURL) {
		t.Errorf("expected redirect to %s, got %s", cfg.OAuthAuthorizeURL, loc)
	}
	if !strings.Contains(loc, "scope=openid+email+profile") {
		t.Errorf("expected scope injected, got %s", loc)
	}
}

func TestHandleAuthorizeProxy_PreservesScope(t *testing.T) {
	cfg := testConfig()
	h := handlers.HandleAuthorizeProxy(cfg)

	req := httptest.NewRequest("GET", "/authorize?client_id=x&redirect_uri=http://localhost&scope=custom", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	loc := rec.Header().Get("Location")
	if !strings.Contains(loc, "scope=custom") {
		t.Errorf("expected custom scope preserved, got %s", loc)
	}
}

func TestHandleTokenProxy_MethodNotAllowed(t *testing.T) {
	cfg := testConfig()
	h := handlers.HandleTokenProxy(cfg)

	req := httptest.NewRequest("GET", "/token", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestHandleConfig(t *testing.T) {
	cfg := testConfig()
	h := handlers.HandleConfig(cfg)

	req := httptest.NewRequest("GET", "/config", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var body map[string]string
	_ = json.NewDecoder(rec.Body).Decode(&body)

	if body["type"] != "oauth" {
		t.Errorf("expected type=oauth, got %s", body["type"])
	}
	if body["clientId"] != "test-client-id" {
		t.Errorf("expected clientId=test-client-id, got %s", body["clientId"])
	}
}

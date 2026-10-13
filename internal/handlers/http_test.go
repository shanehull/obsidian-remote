package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
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
		OAuthCallbackURL:  "https://bridge.example.com/oauth/callback",
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

func TestHandleAuthorizeProxy_RewritesRedirectURI(t *testing.T) {
	cfg := testConfig()
	h := handlers.HandleAuthorizeProxy(cfg)

	req := httptest.NewRequest("GET", "/authorize?client_id=x&redirect_uri=http%3A%2F%2Flocalhost%3A8080%2Fcallback&state=client-state&code_challenge=abc&code_challenge_method=S256", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	loc, err := url.Parse(rec.Header().Get("Location"))
	if err != nil {
		t.Fatalf("invalid location: %v", err)
	}

	if got := loc.Query().Get("redirect_uri"); got != cfg.OAuthCallbackURL {
		t.Errorf("expected redirect_uri=%s, got %s", cfg.OAuthCallbackURL, got)
	}
	if loc.Query().Get("state") == "client-state" {
		t.Error("expected client state to be wrapped, not passed through")
	}
	if loc.Query().Get("code_challenge") != "abc" {
		t.Error("expected PKCE challenge to be preserved")
	}
}

func TestHandleAuthorizeProxy_RequiresRedirectURI(t *testing.T) {
	cfg := testConfig()
	h := handlers.HandleAuthorizeProxy(cfg)

	req := httptest.NewRequest("GET", "/authorize?client_id=x", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestHandleAuthorizeCallback_RoundTrip(t *testing.T) {
	cfg := testConfig()
	clientRedirect := "http://localhost:8080/callback"

	authReq := httptest.NewRequest("GET", "/authorize?client_id=x&redirect_uri="+url.QueryEscape(clientRedirect)+"&state=client-state", nil)
	authRec := httptest.NewRecorder()
	handlers.HandleAuthorizeProxy(cfg).ServeHTTP(authRec, authReq)

	wrapped, err := url.Parse(authRec.Header().Get("Location"))
	if err != nil {
		t.Fatalf("invalid location: %v", err)
	}

	cbReq := httptest.NewRequest("GET", "/oauth/callback?code=auth-code&state="+url.QueryEscape(wrapped.Query().Get("state")), nil)
	cbRec := httptest.NewRecorder()
	handlers.HandleAuthorizeCallback(cfg).ServeHTTP(cbRec, cbReq)

	if cbRec.Code != http.StatusFound {
		t.Fatalf("expected 302, got %d", cbRec.Code)
	}

	loc, err := url.Parse(cbRec.Header().Get("Location"))
	if err != nil {
		t.Fatalf("invalid location: %v", err)
	}
	if loc.Scheme != "http" || loc.Host != "localhost:8080" || loc.Path != "/callback" {
		t.Errorf("expected redirect to %s, got %s", clientRedirect, loc)
	}
	if loc.Query().Get("code") != "auth-code" {
		t.Errorf("expected code forwarded, got %s", loc.Query().Get("code"))
	}
	if loc.Query().Get("state") != "client-state" {
		t.Errorf("expected client state restored, got %s", loc.Query().Get("state"))
	}
	if loc.Query().Has("iss") {
		t.Error("iss must not be forwarded to the client")
	}
}

func TestHandleAuthorizeCallback_ErrorPassthrough(t *testing.T) {
	cfg := testConfig()
	h := handlers.HandleAuthorizeCallback(cfg)

	authReq := httptest.NewRequest("GET", "/authorize?client_id=x&redirect_uri=http%3A%2F%2Flocalhost%2Fcb&state=s1", nil)
	authRec := httptest.NewRecorder()
	handlers.HandleAuthorizeProxy(cfg).ServeHTTP(authRec, authReq)
	wrapped, _ := url.Parse(authRec.Header().Get("Location"))

	cbReq := httptest.NewRequest("GET", "/oauth/callback?error=access_denied&error_description=nope&state="+url.QueryEscape(wrapped.Query().Get("state")), nil)
	cbRec := httptest.NewRecorder()
	h.ServeHTTP(cbRec, cbReq)

	loc, _ := url.Parse(cbRec.Header().Get("Location"))
	if loc.Query().Get("error") != "access_denied" {
		t.Errorf("expected error forwarded, got %s", loc.Query().Get("error"))
	}
	if loc.Query().Get("code") != "" {
		t.Error("expected no code on error response")
	}
	if loc.Query().Get("state") != "s1" {
		t.Errorf("expected state restored, got %s", loc.Query().Get("state"))
	}
}

func TestHandleAuthorizeCallback_RejectsTamperedState(t *testing.T) {
	cfg := testConfig()
	h := handlers.HandleAuthorizeCallback(cfg)

	req := httptest.NewRequest("GET", "/oauth/callback?code=x&state=not-a-valid-state", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestHandleTokenProxy_RewritesRedirectURI(t *testing.T) {
	var received url.Values
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		received = r.Form
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"token"}`))
	}))
	defer upstream.Close()

	cfg := testConfig()
	cfg.OAuthTokenURL = upstream.URL

	form := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {"abc"},
		"redirect_uri":  {"http://localhost:8080/callback"},
		"code_verifier": {"verifier"},
	}
	req := httptest.NewRequest("POST", "/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	handlers.HandleTokenProxy(cfg).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if got := received.Get("redirect_uri"); got != cfg.OAuthCallbackURL {
		t.Errorf("expected redirect_uri=%s, got %s", cfg.OAuthCallbackURL, got)
	}
	if got := received.Get("client_secret"); got != cfg.OAuthClientSecret {
		t.Errorf("expected client_secret=%s, got %s", cfg.OAuthClientSecret, got)
	}
	if got := received.Get("code_verifier"); got != "verifier" {
		t.Errorf("expected code_verifier preserved, got %s", got)
	}
}

func TestHandleTokenProxy_RefreshGrantKeepsRedirectURI(t *testing.T) {
	var received url.Values
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		received = r.Form
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"token"}`))
	}))
	defer upstream.Close()

	cfg := testConfig()
	cfg.OAuthTokenURL = upstream.URL

	form := url.Values{"grant_type": {"refresh_token"}, "refresh_token": {"rt"}}
	req := httptest.NewRequest("POST", "/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	handlers.HandleTokenProxy(cfg).ServeHTTP(rec, req)

	if got := received.Get("redirect_uri"); got != "" {
		t.Errorf("expected no redirect_uri on refresh, got %s", got)
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

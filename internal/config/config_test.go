package config_test

import (
	"os"
	"testing"

	"github.com/shanehull/obsidian-remote/internal/config"
)

func TestLoadDefaults(t *testing.T) {
	_ = os.Unsetenv("VAULT_PATH")
	_ = os.Unsetenv("OBSIDIAN_URL")
	_ = os.Unsetenv("OBSIDIAN_KEY")
	_ = os.Unsetenv("OAUTH_ISSUER")
	_ = os.Unsetenv("OAUTH_JWKS_URL")
	_ = os.Unsetenv("OAUTH_AUTHORIZE_URL")
	_ = os.Unsetenv("OAUTH_TOKEN_URL")
	_ = os.Unsetenv("PUBLIC_HOST")
	_ = os.Unsetenv("OAUTH_AUDIENCE")
	_ = os.Unsetenv("OAUTH_CLIENT_SECRET")
	_ = os.Unsetenv("OAUTH_ALLOWED_EMAIL")

	cfg := config.Load()

	if cfg.VaultPath != "/vaults" {
		t.Errorf("expected VaultPath=/vaults, got %s", cfg.VaultPath)
	}
	if cfg.ObsidianURL != "http://127.0.0.1:27124" {
		t.Errorf("expected ObsidianURL=http://127.0.0.1:27124, got %s", cfg.ObsidianURL)
	}
	if cfg.ObsidianKey != "bridge-key" {
		t.Errorf("expected ObsidianKey=bridge-key, got %s", cfg.ObsidianKey)
	}
	if cfg.OAuthIssuer != "https://accounts.google.com" {
		t.Errorf("expected Google issuer, got %s", cfg.OAuthIssuer)
	}
	if cfg.OAuthJwksURL != "https://www.googleapis.com/oauth2/v3/certs" {
		t.Errorf("expected Google JWKS, got %s", cfg.OAuthJwksURL)
	}
	if cfg.OAuthAudience != "" {
		t.Errorf("expected empty OAuthAudience, got %s", cfg.OAuthAudience)
	}
	if cfg.PublicHost != "" {
		t.Errorf("expected empty PublicHost, got %s", cfg.PublicHost)
	}
}

func TestLoadOverrides(t *testing.T) {
	_ = os.Setenv("VAULT_PATH", "/custom-vault")
	_ = os.Setenv("OBSIDIAN_URL", "http://localhost:9999")
	_ = os.Setenv("OBSIDIAN_KEY", "custom-key")
	_ = os.Setenv("OAUTH_AUDIENCE", "test-client-id")
	_ = os.Setenv("OAUTH_CLIENT_SECRET", "test-secret")
	_ = os.Setenv("PUBLIC_HOST", "https://example.com")
	_ = os.Setenv("OAUTH_ALLOWED_EMAIL", "admin@example.com")
	defer func() { _ = os.Unsetenv("VAULT_PATH") }()
	defer func() { _ = os.Unsetenv("OBSIDIAN_URL") }()
	defer func() { _ = os.Unsetenv("OBSIDIAN_KEY") }()
	defer func() { _ = os.Unsetenv("OAUTH_AUDIENCE") }()
	defer func() { _ = os.Unsetenv("OAUTH_CLIENT_SECRET") }()
	defer func() { _ = os.Unsetenv("PUBLIC_HOST") }()
	defer func() { _ = os.Unsetenv("OAUTH_ALLOWED_EMAIL") }()

	cfg := config.Load()

	if cfg.VaultPath != "/custom-vault" {
		t.Errorf("expected VaultPath=/custom-vault, got %s", cfg.VaultPath)
	}
	if cfg.ObsidianURL != "http://localhost:9999" {
		t.Errorf("expected ObsidianURL=http://localhost:9999, got %s", cfg.ObsidianURL)
	}
	if cfg.ObsidianKey != "custom-key" {
		t.Errorf("expected ObsidianKey=custom-key, got %s", cfg.ObsidianKey)
	}
	if cfg.OAuthAudience != "test-client-id" {
		t.Errorf("expected OAuthAudience=test-client-id, got %s", cfg.OAuthAudience)
	}
	if cfg.OAuthClientSecret != "test-secret" {
		t.Errorf("expected OAuthClientSecret=test-secret, got %s", cfg.OAuthClientSecret)
	}
	if cfg.PublicHost != "https://example.com" {
		t.Errorf("expected PublicHost=https://example.com, got %s", cfg.PublicHost)
	}
	if cfg.AllowedEmail != "admin@example.com" {
		t.Errorf("expected AllowedEmail=admin@example.com, got %s", cfg.AllowedEmail)
	}
}

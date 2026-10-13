package config

import (
	"os"
	"strings"
)

type Config struct {
	VaultPath         string
	ObsidianURL       string
	ObsidianKey       string
	OAuthIssuer       string
	OAuthJwksURL      string
	OAuthAuthorizeURL string
	OAuthTokenURL     string
	PublicHost        string
	OAuthAudience     string
	OAuthClientSecret string
	OAuthCallbackURL  string
	AllowedEmail      string
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func Load() *Config {
	publicHost := os.Getenv("PUBLIC_HOST")

	// The upstream IdP redirects here, not to the MCP client, so the proxy can
	// return the authorization response under its own issuer (RFC 9207).
	callbackURL := os.Getenv("OAUTH_CALLBACK_URL")
	if callbackURL == "" && publicHost != "" {
		callbackURL = strings.TrimSuffix(publicHost, "/") + "/oauth/callback"
	}

	return &Config{
		VaultPath:         getEnv("VAULT_PATH", "/vaults"),
		ObsidianURL:       getEnv("OBSIDIAN_URL", "http://127.0.0.1:27124"),
		ObsidianKey:       getEnv("OBSIDIAN_KEY", "bridge-key"),
		OAuthIssuer:       getEnv("OAUTH_ISSUER", "https://accounts.google.com"),
		OAuthJwksURL:      getEnv("OAUTH_JWKS_URL", "https://www.googleapis.com/oauth2/v3/certs"),
		OAuthAuthorizeURL: getEnv("OAUTH_AUTHORIZE_URL", "https://accounts.google.com/o/oauth2/v2/auth"),
		OAuthTokenURL:     getEnv("OAUTH_TOKEN_URL", "https://oauth2.googleapis.com/token"),
		PublicHost:        publicHost,
		OAuthAudience:     os.Getenv("OAUTH_AUDIENCE"),
		OAuthClientSecret: os.Getenv("OAUTH_CLIENT_SECRET"),
		OAuthCallbackURL:  callbackURL,
		AllowedEmail:      os.Getenv("OAUTH_ALLOWED_EMAIL"),
	}
}

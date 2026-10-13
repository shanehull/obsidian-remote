package config

import (
	"os"
)

type Config struct {
	VaultPath         string
	ObsidianURL       string
	ObsidianKey       string
	OAuthIssuer       string
	OAuthJwksURL      string
	OAuthAuthorizeURL string
	OAuthTokenURL     string
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
	return &Config{
		VaultPath:         getEnv("VAULT_PATH", "/vaults"),
		ObsidianURL:       getEnv("OBSIDIAN_URL", "http://127.0.0.1:27124"),
		ObsidianKey:       getEnv("OBSIDIAN_KEY", "bridge-key"),
		OAuthIssuer:       getEnv("OAUTH_ISSUER", "https://accounts.google.com"),
		OAuthJwksURL:      getEnv("OAUTH_JWKS_URL", "https://www.googleapis.com/oauth2/v3/certs"),
		OAuthAuthorizeURL: getEnv("OAUTH_AUTHORIZE_URL", "https://accounts.google.com/o/oauth2/v2/auth"),
		OAuthTokenURL:     getEnv("OAUTH_TOKEN_URL", "https://oauth2.googleapis.com/token"),
		OAuthAudience:     os.Getenv("OAUTH_AUDIENCE"),
		OAuthClientSecret: os.Getenv("OAUTH_CLIENT_SECRET"),
		// Absolute URL registered with the provider. When unset, the handler
		// derives it from the request origin.
		OAuthCallbackURL: os.Getenv("OAUTH_CALLBACK_URL"),
		AllowedEmail:     os.Getenv("OAUTH_ALLOWED_EMAIL"),
	}
}

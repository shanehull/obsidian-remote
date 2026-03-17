package middleware

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/MicahParks/keyfunc/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/shanehull/obsidian-remote/internal/config"
)

func Auth(cfg *config.Config, jwks *keyfunc.JWKS) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if jwks == nil {
				next.ServeHTTP(w, r)
				return
			}

			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				scheme := "http"
				if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
					scheme = "https"
				}
				host := fmt.Sprintf("%s://%s", scheme, r.Host)
				metadataURL := fmt.Sprintf("%s/.well-known/oauth-protected-resource%s", strings.TrimSuffix(host, "/"), r.URL.Path)

				w.Header().Set("WWW-Authenticate", fmt.Sprintf("Bearer resource_metadata=\"%s\"", metadataURL))
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")

			// Try JWT validation first (ID tokens)
			token, err := jwt.Parse(tokenString, jwks.Keyfunc)
			if err == nil && token.Valid {
				claims, _ := token.Claims.(jwt.MapClaims)
				if !validateClaims(cfg, claims) {
					http.Error(w, "Forbidden", http.StatusForbidden)
					return
				}
				next.ServeHTTP(w, r)
				return
			}

			// Fallback: validate opaque access tokens via Google's tokeninfo
			email, err := validateAccessToken(tokenString, cfg.OAuthAudience)
			if err != nil {
				slog.Warn("invalid token", "error", err, "remote", r.RemoteAddr)
				http.Error(w, "Unauthorized: Invalid token", http.StatusUnauthorized)
				return
			}

			if cfg.AllowedEmail != "" && email != cfg.AllowedEmail {
				slog.Warn("email not allowed", "got", email, "want", cfg.AllowedEmail)
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func validateClaims(cfg *config.Config, claims jwt.MapClaims) bool {
	if cfg.OAuthAudience != "" {
		aud, _ := claims.GetAudience()
		if !contains(aud, cfg.OAuthAudience) {
			slog.Warn("audience mismatch", "got", aud, "want", cfg.OAuthAudience)
			return false
		}
	}
	if cfg.AllowedEmail != "" {
		email, _ := claims["email"].(string)
		if email != cfg.AllowedEmail {
			slog.Warn("email not allowed", "got", email, "want", cfg.AllowedEmail)
			return false
		}
	}
	return true
}

func validateAccessToken(token, expectedAudience string) (string, error) {
	resp, err := http.Get("https://oauth2.googleapis.com/tokeninfo?access_token=" + token)
	if err != nil {
		return "", fmt.Errorf("tokeninfo request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("tokeninfo returned %d", resp.StatusCode)
	}

	var info struct {
		Aud   string `json:"aud"`
		Email string `json:"email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return "", fmt.Errorf("tokeninfo decode failed: %w", err)
	}

	if expectedAudience != "" && info.Aud != expectedAudience {
		return "", fmt.Errorf("audience mismatch: got %s, want %s", info.Aud, expectedAudience)
	}

	return info.Email, nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/MicahParks/keyfunc/v2"
	"github.com/mark3labs/mcp-go/server"
	"github.com/shanehull/obsidian-remote/internal/config"
	"github.com/shanehull/obsidian-remote/internal/handlers"
	"github.com/shanehull/obsidian-remote/internal/middleware"
	"github.com/shanehull/obsidian-remote/internal/obsidian"
)

func main() {
	// Initialize structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg := config.Load()
	obsClient := obsidian.NewClient(cfg)

	s := server.NewMCPServer("Obsidian Remote (Go)", "1.0.0")
	handlers.RegisterTools(s, obsClient)

	var jwks *keyfunc.JWKS
	var err error
	if cfg.OAuthAudience != "" {
		jwks, err = keyfunc.Get(cfg.OAuthJwksURL, keyfunc.Options{
			RefreshInterval: time.Hour,
		})
		if err != nil {
			slog.Warn("failed to fetch JWKS", "url", cfg.OAuthJwksURL, "error", err)
		}
	}

	sse := server.NewSSEServer(s, server.WithBaseURL(cfg.PublicHost))
	streamable := server.NewStreamableHTTPServer(s)

	// Auth Middleware
	auth := middleware.Auth(cfg, jwks)

	// Setup HTTP Handlers
	mux := http.NewServeMux()
	registerHTTPHandlers(mux, cfg, sse, streamable, auth)

	// Wrap with CORS and Request Logger
	handler := loggingMiddleware(enableCORS(mux))

	slog.Info("Headless MCP Bridge starting", "port", 4000)
	if err := http.ListenAndServe(":4000", handler); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func registerHTTPHandlers(mux *http.ServeMux, cfg *config.Config, sse *server.SSEServer, streamable *server.StreamableHTTPServer, auth func(http.Handler) http.Handler) {
	mux.Handle("/sse", auth(sse.SSEHandler()))
	mux.Handle("/message", auth(sse.MessageHandler()))
	mux.Handle("/mcp", auth(streamable))

	// RFC 9728 Discovery (prefix match)
	mux.Handle("/.well-known/oauth-protected-resource/", handlers.HandleDiscovery(cfg))
	mux.HandleFunc("/.well-known/oauth-protected-resource", handlers.HandleDiscovery(cfg))
	mux.HandleFunc("/.well-known/oauth-authorization-server", handlers.HandleAuthServerDiscovery(cfg))
	mux.HandleFunc("/.well-known/mcp", handlers.HandleDiscovery(cfg))

	// Static Registration Endpoint
	mux.HandleFunc("/register", handlers.HandleRegistration(cfg))

	// OAuth proxies (inject scope and client_secret so clients don't need them)
	mux.HandleFunc("/authorize", handlers.HandleAuthorizeProxy(cfg))
	mux.HandleFunc("/token", handlers.HandleTokenProxy(cfg))

	// Dynamic Client Config
	mux.HandleFunc("/config", handlers.HandleConfig(cfg))
}

func enableCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PATCH, PUT")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func loggingMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		h.ServeHTTP(w, r)
		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"remote", r.RemoteAddr,
			"duration", time.Since(start),
		)
	})
}

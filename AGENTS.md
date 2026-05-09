# AGENTS.md — obsidian-remote

MCP bridge server for Obsidian vaults. Exposes Obsidian Local REST API as MCP tools with OAuth2 authentication. Stdio + HTTP/Serve modes.

## Commands

```bash
go build ./... && go vet ./... && golangci-lint run && go test ./...
```

All unit tests, no vault needed.

## Architecture

```
cmd/server/main.go              Single binary entry point
internal/config/config.go       Environment-based configuration
internal/handlers/http.go       HTTP handlers: discovery, OAuth proxy, registration, config, healthz
internal/handlers/tools.go      MCP tool registrations (8 tools)
internal/middleware/auth.go     JWT + opaque token validation middleware
internal/obsidian/client.go     Obsidian Local REST API HTTP client
```

## Conventions

- Tests in `package_test` (external test package) where possible
- Internal helpers use `package` (internal test package)
- Stdlib `testing` only, no testify or gomega
- HTTP handler tests use `httptest.NewRecorder` + `httptest.NewRequest`
- Tools return errors in `CallToolResult`, not as Go errors
- Use `mcp.WithReadOnlyHintAnnotation(true)` / `mcp.WithDestructiveHintAnnotation(false)` for all tools
- Semantic commits for release-please: `feat:` bumps minor (pre-1.0), `fix:` bumps patch. `docs:`, `ci:`, `chore:` don't bump version but appear in changelog.

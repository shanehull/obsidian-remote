# AGENTS.md — obsidian-remote

MCP bridge server for Obsidian vaults. Exposes Obsidian Local REST API as MCP tools with OAuth2 authentication. Stdio + HTTP/Serve modes.

## Commands

```bash
go build ./... && go vet ./... && golangci-lint run && go test ./...
```

All unit tests, no vault needed.

## Conventions

- Semantic commits for release-please: `feat:` bumps minor (pre-1.0), `fix:` bumps patch. `docs:`, `ci:`, `chore:` don't bump version but appear in changelog.

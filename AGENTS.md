# AGENTS.md — obsidian-remote

MCP bridge server for Obsidian vaults. Exposes Obsidian Local REST API as MCP tools with OAuth2 authentication. Stdio + HTTP/Serve modes.

## Commands

```bash
go build ./... && go vet ./... && golangci-lint run && go test ./...
```

All unit tests, no vault needed.

E2E tests (requires Docker, runs against a real Obsidian container):
```bash
test/e2e_test.sh
```

## Conventions

- Semantic commits for release-please: `feat:` bumps minor (pre-1.0), `fix:` bumps patch. `docs:`, `ci:`, `chore:` don't bump version and don't appear in changelog.

## Updating dependencies

Go modules:
```bash
go get github.com/modelcontextprotocol/go-sdk@latest && go mod tidy
```

Only bump direct deps (`go-sdk`). Transitive deps update automatically with `go mod tidy`.

mise (match Go version to `go.mod` directive):
```bash
# edit mise.toml, then:
mise install
```

Docker base image:
```bash
# check: https://github.com/linuxserver/docker-obsidian/pkgs/container/obsidian
# update FROM line in Dockerfile, then:
docker compose build && test/e2e_test.sh
```

CI workflows: update `go-version` in `.github/workflows/lint.yaml` and `test.yaml`.

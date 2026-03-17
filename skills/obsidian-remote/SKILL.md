---
name: obsidian-remote
description: Manage a remote Obsidian vault via MCP (Headless Go Edition).
compatibility: Requires Obsidian Remote (Go) container.
allowed-tools: mcp__obsidian-remote__*
---

# Obsidian Remote Skill (Headless Go)

This skill enables high-performance interaction with a remote Obsidian vault using the Model Context Protocol.

## Configuration

The server is configured via environment variables. Key variables include:

- `PUBLIC_HOST`: The external URL of your MCP server.
- `OAUTH_ISSUER`: Your OAuth provider's issuer URL.
- `OAUTH_JWKS_URL`: Your OAuth provider's JWKS endpoint.
- `OAUTH_AUDIENCE`: Your OAuth Client ID.
- `OAUTH_ALLOWED_EMAIL`: Authorized email for access.

## Tools

- `obsidian_list_notes`: List all files in the vault.
- `obsidian_read_note`: Retrieve the full content of a specific note.
- `obsidian_update_note`: Create or overwrite a note.
- `obsidian_global_search`: Search for text or regex across all notes.
- `obsidian_manage_frontmatter`: Get or set YAML frontmatter keys (via `/metadata/` endpoint).

## Usage

Configure your MCP client to connect to the server's SSE endpoint (Port 4000). The Gemini CLI will automatically handle OIDC discovery and authentication.

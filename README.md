# Obsidian Remote

A high-performance MCP server for your Obsidian vault, written in Go. This server acts as a bridge between MCP clients and the Obsidian [Local REST API](https://github.com/coddingtonbear/obsidian-local-rest-api).

## Features

- **Headless & Fast:** Direct filesystem and API access with zero GUI overhead.
- **RFC 9728 Compliant:** Implements OAuth-protected resource discovery.
- **Dual Transport:** Supports both Streamable HTTP and SSE transports.
- **Secure:** Integrated JWT and OAuth access token validation with email-based access control.
- **Server-Side Token Proxy:** Clients never need the OAuth client secret — the server injects it during the token exchange.
- **Provider Agnostic:** Supports any OpenID Connect (OIDC) provider (Google, GitHub, Auth0, etc.).

## Prerequisites

- An Obsidian vault with the [Local REST API](https://github.com/coddingtonbear/obsidian-local-rest-api) plugin installed and configured.
- A public URL (HTTPS) if accessing from outside your local network.
- An OAuth client ID (and secret) from your OIDC provider.

## Setup

1. **Configure Environment:**

   ```bash
   cp .env.example .env
   ```

   Edit `.env` with your configuration:
   - `PUBLIC_HOST`: The external URL of this server (e.g., `https://obsidian.yourdomain.com`).
   - `OAUTH_ISSUER`: Your OIDC issuer (e.g., `https://accounts.google.com`).
   - `OAUTH_JWKS_URL`: The URL to fetch public signing keys (e.g., `https://www.googleapis.com/oauth2/v3/certs`).
   - `OAUTH_AUDIENCE`: Your OAuth Client ID.
   - `OAUTH_CLIENT_SECRET`: Your OAuth Client Secret (used server-side for the token exchange proxy).
   - `OAUTH_ALLOWED_EMAIL`: The specific email address authorized to access the vault.

2. **Run with Docker:**
   ```bash
   docker-compose up -d --build
   ```

## Client Configuration

### Gemini CLI

Add to your `~/.config/gemini/settings.json` (or `.gemini/settings.json` in the project):

```json
{
  "mcpServers": {
    "obsidian-remote": {
      "httpUrl": "https://obsidian.yourdomain.com/mcp",
      "oauth": {
        "clientId": "your-client-id.apps.googleusercontent.com",
        "scopes": ["openid", "email", "profile"]
      }
    }
  }
}
```

Then authenticate:

```
/mcp auth obsidian-remote
```

The server handles OAuth discovery automatically via `/.well-known/oauth-protected-resource` and proxies the token exchange so clients never need the client secret.

### Other Clients (SSE)

Clients that support the SSE transport can connect to `/sse` instead:

```json
{
  "mcpServers": {
    "obsidian-remote": {
      "url": "https://obsidian.yourdomain.com/sse"
    }
  }
}
```

## Endpoints

| Path                                      | Method          | Description                                  |
| :---------------------------------------- | :-------------- | :------------------------------------------- |
| `/mcp`                                    | POST/GET/DELETE | Streamable HTTP transport (recommended)      |
| `/sse`                                    | GET             | SSE transport                                |
| `/message`                                | POST            | SSE message endpoint                         |
| `/token`                                  | POST            | Token exchange proxy (injects client secret) |
| `/.well-known/oauth-protected-resource`   | GET             | RFC 9728 resource metadata                   |
| `/.well-known/oauth-authorization-server` | GET             | RFC 8414 authorization server metadata       |
| `/register`                               | POST            | Dynamic client registration (RFC 7591)       |

## Available Tools

- `obsidian_list_notes`: List all notes in the vault.
- `obsidian_read_note`: Read the content of a specific note.
- `obsidian_update_note`: Create or update a note.
- `obsidian_global_search`: Search for text or regex across the vault.
- `obsidian_search_replace`: Targeted search and replace within a specific file.
- `obsidian_manage_frontmatter`: Manage YAML frontmatter keys.
- `obsidian_manage_tags`: Add or remove tags from a note.

## Architecture

The server supports both Streamable HTTP (`/mcp`) and SSE (`/sse`) transports, converting MCP tool calls into HTTP requests for the Obsidian Local REST API. Authentication is handled via OAuth 2.0 with support for both JWT (ID tokens) and opaque access tokens validated against the provider's tokeninfo endpoint.

# Obsidian Remote MCP Setup

This reference provides configuration examples for connecting MCP clients to your Obsidian Remote server.

Replace `<server-url>` with your server's public URL and `<client-id>` with your OAuth Client ID.

## Gemini CLI (Streamable HTTP)

**Config File:** `~/.config/gemini/settings.json`

```json
{
  "mcpServers": {
    "obsidian-remote": {
      "httpUrl": "<server-url>/mcp",
      "oauth": {
        "clientId": "<client-id>",
        "scopes": ["openid", "email", "profile"]
      }
    }
  }
}
```

Then authenticate inside Gemini:

```
/mcp auth obsidian-remote
```

## Cursor

Cursor supports OAuth discovery via RFC 9728. Add to your MCP config:

```json
{
  "mcpServers": {
    "obsidian-remote": {
      "url": "<server-url>/mcp"
    }
  }
}
```

## Amp (Sourcegraph)

**Config File:** `~/.config/agents/skills/obsidian-remote/mcp.json`

```json
{
  "obsidian-remote": {
    "url": "<server-url>/sse"
  }
}
```

## Other SSE Clients

Clients that support the SSE transport can connect to `/sse`:

```json
{
  "mcpServers": {
    "obsidian-remote": {
      "url": "<server-url>/sse"
    }
  }
}
```

## Tools

| Tool | Description |
| :--- | :--- |
| `read_note` | Retrieve content and metadata (path, tags, frontmatter). |
| `update_note` | Create or overwrite notes. |
| `delete_note` | Permanently delete a note. |
| `list_notes` | List files and folders in the vault. |
| `global_search` | Search for text or regex across the entire vault. |
| `search_replace` | Targeted search and replace within a specific file. |
| `manage_frontmatter` | Get, set, or delete specific YAML frontmatter keys. |
| `manage_tags` | Add or remove tags from a note. |

## Troubleshooting

- **"No client ID provided":** Add `oauth.clientId` to your MCP server config. Gemini CLI does not yet support dynamic client registration.
- **"client_secret is missing":** The server's `/token` proxy handles this. Make sure `OAUTH_CLIENT_SECRET` is set in the server's `.env`.
- **"Protected resource does not match":** Clear cached tokens and re-authenticate.
- **401 Unauthorized:** Your token is invalid/expired, or `OAUTH_ALLOWED_EMAIL` doesn't match your Google account.
- **Connection Refused:** Ensure the container is running and the port is open.

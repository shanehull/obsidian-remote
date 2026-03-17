# Obsidian Remote MCP Setup

This reference provides configuration examples for connecting popular MCP clients to your Obsidian Remote server.

## Configure MCP Client

Replace `<server-url>` with your server's endpoint and `<client-id>` with your OAuth Client ID.

### Gemini CLI (Streamable HTTP — recommended)
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

### Cursor / Other SSE Clients
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
| `obsidian_read_note` | Retrieve content and metadata (path, tags, frontmatter). |
| `obsidian_update_note` | Create, append, prepend, or overwrite notes. |
| `obsidian_list_notes` | List files and folders in the vault. |
| `obsidian_global_search` | Search for text or regex across the entire vault. |
| `obsidian_search_replace` | Targeted search and replace within a specific file. |
| `obsidian_manage_frontmatter` | Get, set, or delete specific YAML frontmatter keys. |
| `obsidian_manage_tags` | Add or remove tags from a note. |

## Troubleshooting

- **"No client ID provided":** Add `oauth.clientId` to your MCP server config. Gemini CLI does not yet support dynamic client registration.
- **"client_secret is missing":** The server's `/token` proxy handles this. Make sure `OAUTH_CLIENT_SECRET` is set in the server's `.env`.
- **"Protected resource does not match":** Clear cached tokens in `~/.gemini/mcp-oauth-tokens.json` and re-authenticate.
- **401 Unauthorized:** Your token is invalid/expired, or `OAUTH_ALLOWED_EMAIL` doesn't match your Google account.
- **Connection Refused:** Ensure the container is running and the port is open.

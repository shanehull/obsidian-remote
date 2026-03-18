# Obsidian Remote MCP Setup

This reference provides configuration examples for connecting MCP clients to your Obsidian Remote server.

Replace `<server-url>` with your server's public URL.

## Gemini CLI (Streamable HTTP)

**Config File:** `~/.config/gemini/settings.json`

```json
{
  "mcpServers": {
    "obsidian-remote": {
      "httpUrl": "<server-url>/mcp"
    }
  }
}
```

Then authenticate inside Gemini:

```
/mcp auth obsidian-remote
```

## Cursor (SSE)

Cursor supports OAuth discovery via RFC 9728. Add to your MCP config:

```json
{
  "mcpServers": {
    "obsidian-remote": {
      "url": "<server-url>/sse"
    }
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

| Tool                 | Description                                              |
| :------------------- | :------------------------------------------------------- |
| `read_note`          | Retrieve content and metadata (path, tags, frontmatter). |
| `update_note`        | Create or overwrite notes.                               |
| `append_note`        | Append content to the end of an existing note.           |
| `delete_note`        | Permanently delete a note.                               |
| `list_notes`         | List files and folders in the vault.                     |
| `global_search`      | Search for text or regex across the entire vault.        |
| `search_replace`     | Targeted search and replace within a specific file.      |
| `manage_frontmatter` | Get, set, or delete specific YAML frontmatter keys.      |
| `manage_tags`        | Add or remove tags from a note.                          |

## Troubleshooting

- **401 Unauthorized:** Your token is invalid/expired, or `OAUTH_ALLOWED_EMAIL` doesn't match your Google account.
- **Connection Refused:** Ensure the container is running and the port is open.

# Obsidian Remote MCP Setup

This reference provides configuration examples for connecting popular MCP clients to your Obsidian Remote server.

## Configure MCP Client

Replace `<host-ip>` with your VPS IP address or domain name.

### Gemini CLI
File: `~/.config/gemini/settings.json`

```json
{
  "mcpServers": {
    "obsidian-remote": {
      "url": "http://<host-ip>:4000/mcp"
    }
  }
}
```

### Claude Desktop
File: `~/Library/Application Support/Claude/claude_desktop_config.json`

```json
{
  "mcpServers": {
    "obsidian-remote": {
      "url": "http://<host-ip>:4000/mcp"
    }
  }
}
```

### Cursor
1. Open Cursor Settings > MCP.
2. Click **+ Add New MCP Server**.
3. Name: `obsidian-remote`
4. Type: `SSE`
5. URL: `http://<host-ip>:4000/mcp`

### Gemini CLI Skill
File: `~/.config/agents/skills/obsidian-remote/mcp.json`

```json
{
  "obsidian-remote": {
    "url": "http://<host-ip>:4000/mcp"
  }
}
```

## Tools

The following tools are available once connected:

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

- **Connection Refused:** Ensure the container is running (`docker-compose ps`) and port 4000 is open in your firewall.
- **404 Not Found:** Verify the URL ends in `/mcp`.
- **401 Unauthorized:** Ensure the `OBSIDIAN_API_KEY` in your `.env` matches the one generated in the Obsidian Web UI.
- **REST API Not Active:** You must manually click **"Trust"** once in the VNC Web UI (`http://<host-ip>:3000`) for the vault to open and the plugin to start.

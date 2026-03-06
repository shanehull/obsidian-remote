# Obsidian Remote

A fully containerized Obsidian instance with a built-in MCP server for automated note management.

> [!CAUTION]
> **Security Warning:** By default, the MCP server on port 4000 is **publicly accessible**. You MUST configure `MCP_AUTH_MODE=jwt` and `MCP_AUTH_SECRET_KEY` in your `.env` file before exposing this to the internet.

## Quick Start

1.  **Configure environment:**
    ```bash
    cp .env.example .env
    ```
    Edit `.env` and set your credentials. **Security Recommendation:**
    - `MCP_AUTH_MODE=jwt`
    - `MCP_AUTH_SECRET_KEY=your_strong_secret`
    - `OBSIDIAN_API_KEY`: (Optional) Leave blank to have one generated automatically.

2.  **Start the server:**
    ```bash
    docker-compose up -d --build
    ```

3.  **Retrieve API Key (if generated):**
    If you left `OBSIDIAN_API_KEY` blank, retrieve the generated key:
    ```bash
    docker exec obsidian cat /config/.obsidian_api_key
    ```

## Client Configuration

Replace `<server-endpoint>` with your server's address. Use an environment variable (e.g., `$OBSIDIAN_REMOTE_JWT`) for your token.

### Gemini CLI
**Config File:** `~/.gemini/settings.json`
```json
{
  "mcpServers": {
    "obsidian-remote": {
      "url": "<server-endpoint>",
      "headers": {
        "Authorization": "Bearer $OBSIDIAN_REMOTE_JWT"
      }
    }
  }
}
```
**CLI command:**
```bash
gemini mcp add obsidian-remote <server-endpoint> --transport http --header "Authorization: Bearer $OBSIDIAN_REMOTE_JWT"
```

### Amp (Sourcegraph)
**Config File:** `~/.config/agents/skills/obsidian-remote/mcp.json`
```json
{
  "obsidian-remote": {
    "url": "<server-endpoint>",
    "headers": {
      "Authorization": "Bearer $OBSIDIAN_REMOTE_JWT"
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
      "url": "<server-endpoint>",
      "headers": {
        "Authorization": "Bearer $OBSIDIAN_REMOTE_JWT"
      }
    }
  }
}
```

## Architecture

- **Zero-Click Startup:** Automatically handles vault cloning and plugin configuration.
- **VNC Management:** Port 3000 (protected by `VNC_PASSWORD`).
- **MCP Server:** Port 4000 (protected by `MCP_AUTH_SECRET_KEY`).
- **Resource Optimized:** 256MB SHM for stability.

## Security

- **Web UI:** Protected by `VNC_PASSWORD`.
- **MCP Server:** Protected by `MCP_AUTH_SECRET_KEY` (JWT).
- **Obsidian API:** Protected by `OBSIDIAN_API_KEY` (Internal only).

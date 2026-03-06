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

## Client Setup (Skill + MCP)

To use this with AI agents, you must register the **Skill** (documentation) and the **MCP Server** (connection).

### 1. Register the Skill

Install the skill documentation into your agent CLI:

#### Gemini CLI
```bash
gemini skills install https://github.com/shanehull/obsidian-remote --path skills/obsidian-remote
```

#### Amp (Sourcegraph)
Amp automatically discovers skills in your workspace. If you are using it globally, ensure the `obsidian-remote` directory is in your `~/.config/agents/skills/` path.

### 2. Configure the MCP Server

Replace `<server-endpoint>` with your server's address and `<your-jwt>` with a token signed by your `MCP_AUTH_SECRET_KEY`.

#### Gemini CLI
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

#### Amp (Sourcegraph)
The MCP server configuration is already included in the skill's `mcp.json`. Ensure the URL in `~/.config/agents/skills/obsidian-remote/mcp.json` points to your endpoint.

#### Cursor
Add to `~/.cursor/mcp.json`:
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

- **Zero-Click Startup:** Automatically handles vault cloning, plugin configuration, and "Trusting" the vault.
- **MCP Server:** Programmatic bridge for AI agents listening on Port 4000.
- **Resource Optimized:** 256MB SHM for stability.

## Security

- **Web UI:** Protected by `VNC_PASSWORD`.
- **MCP Server:** Protected by `MCP_AUTH_SECRET_KEY` (JWT).
- **Obsidian API:** Protected by `OBSIDIAN_API_KEY` (Internal only).

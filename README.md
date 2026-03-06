# Obsidian Remote

A fully containerized Obsidian instance with a built-in MCP server for automated note management. Designed for resource-constrained environments like e2-micro.

## Quick Start

You can pull the pre-built image from the GitHub Container Registry:

```bash
docker pull ghcr.io/shanehull/obsidian-remote:latest
```

### Run with Docker Compose (Recommended)
1.  **Configure environment:**
    ```bash
    cp .env.example .env
    ```
    Edit `.env` and set your credentials.

2.  **Start the server:**
    ```bash
    docker-compose up -d --build
    ```

### Run with Docker CLI
```bash
docker run -d \
  --name obsidian \
  --shm-size="256mb" \
  -p 3000:3000 \
  -p 4000:4000 \
  -p 27123:27123 \
  -v $(pwd)/config:/config \
  -v $(pwd)/vaults:/vaults \
  -e TEST_MODE=true \
  -e PASSWORD=your_vnc_password \
  -e OBSIDIAN_API_KEY=your_api_key \
  -e GIT_REPO_URL=git@github.com:user/repo.git \
  -e GITHUB_PAT=your_github_pat \
  ghcr.io/shanehull/obsidian-remote:latest
```

## Client Configuration

Replace `<server-endpoint>` with your server's address (e.g., `http://<ip>:4000/mcp` or `https://obsidian.domain.com/mcp`).

### Gemini CLI
File: `~/.config/gemini/settings.json`
```json
{
  "mcpServers": {
    "obsidian-remote": {
      "url": "<server-endpoint>"
    }
  }
}
```

### Cursor
File: `~/.cursor/mcp.json`
```json
{
  "mcpServers": {
    "obsidian-remote": {
      "url": "<server-endpoint>"
    }
  }
}
```

### Amp (Sourcegraph)
File: `~/.config/agents/skills/obsidian-remote/mcp.json`
```json
{
  "obsidian-remote": {
    "url": "<server-endpoint>"
  }
}
```

### Claude Desktop
File: `~/Library/Application Support/Claude/claude_desktop_config.json`
```json
{
  "mcpServers": {
    "obsidian-remote": {
      "url": "<server-endpoint>"
    }
  }
}
```

## Architecture

- **Zero-Click Startup:** Automatically handles vault cloning, plugin configuration, and "Trusting" the vault.
- **Web UI:** Manual vault management (if needed).
- **MCP Server:** Programmatic bridge for AI agents.
- **Persistence:** Vault and config data are stored in local `./vaults` and `./config` directories.

## Security

- Access is protected by `VNC_PASSWORD` (Web UI) and `OBSIDIAN_API_KEY` (API/MCP).
- Communication is plaintext by default; use a reverse proxy for TLS in production.

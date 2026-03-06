# Obsidian Remote

A fully containerized Obsidian instance with a built-in MCP server for automated note management.

## Prerequisites

Before using this in Production Mode (`TEST_MODE=false`), your Obsidian vault **must** have the following configured and committed to your Git repository:

1.  **Obsidian Git Plugin:** Installed, enabled, and configured for auto-sync.
2.  **Local REST API Plugin:** Installed, enabled, and configured.
3.  **Config committed:** Ensure your `.obsidian/` folder (including `community-plugins.json` and `plugins/`) is committed to Git so the server can initialize itself.

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
  -v $(pwd)/config:/config \
  -e TEST_MODE=true \
  -e PASSWORD=your_vnc_password \
  -e OBSIDIAN_API_KEY=your_api_key \
  -e GIT_REPO_URL=git@github.com:user/repo.git \
  -e GITHUB_PAT=your_github_pat \
  ghcr.io/shanehull/obsidian-remote:latest
```

## Client Setup (Skill + MCP)

To use this with AI agents, you must register the **Skill** (the documentation) and the **MCP Server** (the connection).

Replace `<server-endpoint>` with your server's address (e.g., `http://<ip>:4000/mcp` or `https://obsidian.domain.com/mcp`).

### 1. Register the Skill

Install the skill documentation into your agent CLI:

#### Gemini CLI

```bash
gemini skills install https://github.com/shanehull/obsidian-remote --path skills/obsidian-remote
```

#### Amp (Sourcegraph)

Amp automatically discovers skills in your workspace. If you are using it globally, ensure the `obsidian-remote` directory is in your `~/.config/agents/skills/` path.

### 2. Configure the MCP Server

#### Gemini CLI

```bash
gemini mcp add obsidian-remote --transport http <server-endpoint>
```

#### Amp (Sourcegraph)
The MCP server configuration is already included in the skill's `mcp.json`. Once you've installed the skill, simply ensure the URL in `~/.config/agents/skills/obsidian-remote/mcp.json` points to your endpoint. No `amp mcp add` command is required.

#### Claude Desktop

Add to `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "obsidian-remote": {
      "url": "<server-endpoint>"
    }
  }
}
```

#### Cursor

Add to `~/.cursor/mcp.json`:

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

- **Zero-Click Startup:** Automatically handles vault cloning and plugin configuration.
- **Mandatory Manual Step:** You MUST manually click **"Trust"** once in the VNC Web UI (`http://<host-ip>:3000`) for the vault to open and the plugin to start.
- **MCP Server:** Programmatic bridge for AI agents listening on Port 4000.
- **Resource Optimized:** Configured for stability with 256MB SHM.

## Security

- Access is protected by `VNC_PASSWORD` (Web UI) and `OBSIDIAN_API_KEY` (API/MCP).
- Communication is plaintext by default; use a reverse proxy for TLS in production.

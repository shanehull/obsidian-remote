# Obsidian Remote

A containerized Obsidian instance with a built-in MCP server for automated note management.

## Setup

1.  **Configure environment:**
    ```bash
    cp .env.example .env
    ```
    Edit `.env` and set:
    - `TEST_MODE`: `true` (isolated dummy vault) or `false` (real Git vault).
    - `VNC_PASSWORD`: Password for the Web UI.
    - `GIT_REPO_URL`: Your private vault repository.
    - `SSH_PRIVATE_KEY`: Your SSH key (base64 encoded recommended).
    - `OBSIDIAN_API_KEY`: A strong token for the REST API.

2.  **Start the server:**
    ```bash
    docker-compose up -d --build
    ```

3.  **One-time Manual Setup (Mandatory):**
    - Access the Web UI at `http://<host-ip>:3000`.
    - You MUST manually click **"Trust"** when the vault opens.
    - Go to **Settings > Community Plugins** and ensure the Local REST API plugin is enabled.
    - Once port 27123 is active, the MCP server will start working.

4.  **Client configuration:**
    Connect your MCP client to the server's endpoint:
    ```json
    {
      "mcpServers": {
        "obsidian-remote": {
          "url": "http://<host-ip>:4000/mcp"
        }
      }
    }
    ```

## Architecture

- **Web UI (Port 3000):** Manual vault management.
- **MCP Server (Port 4000):** Programmatic bridge for AI agents.
- **Persistence:** Vault and config data are stored in local `./vaults` and `./config` directories.
- **Sync:** Automatically clones and manages your Git-backed vault.

## Security

- Access is protected by `VNC_PASSWORD` (Web UI) and `OBSIDIAN_API_KEY` (API/MCP).
- Communication is plaintext by default; use a reverse proxy for TLS in production.

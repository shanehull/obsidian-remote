# Obsidian Remote

A containerized Obsidian instance with a built-in MCP server for automated note management.

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
    docker-compose up -d
    ```

### Run with Docker CLI
If you prefer not to use Compose, you can run the image directly:

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

## Setup

### Operational Modes
- **TEST_MODE=true (Default):** Automated "Zero-Click" seeding. The container will automatically download, enable, and configure the Local REST API plugin in an isolated dummy vault for immediate testing.
- **TEST_MODE=false:** Production mode. The container clones your real vault from Git and assumes the Local REST API plugin is already committed and configured in your repository.

### Vault Prerequisites
When using your real vault (`TEST_MODE=false`), ensure the following are already committed to your repository:
- **Local REST API Plugin:** Installed and enabled.
- **Obsidian Git Plugin:** Configured for auto-sync.
- **API Key:** Matches the `OBSIDIAN_API_KEY` in your `.env`.

1.  **Configure environment:**
    ```bash
    cp .env.example .env
    ```
    Edit `.env` and set:
    - `TEST_MODE`: `true` (isolated dummy vault) or `false` (real Git vault).
    - `VNC_PASSWORD`: Password for the Web UI.
    - `GIT_REPO_URL`: Your private vault repository (e.g. `git@github.com:user/repo.git`).
    - `GITHUB_PAT`: Your GitHub Personal Access Token with `repo` scope.
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

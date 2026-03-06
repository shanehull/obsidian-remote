---
name: obsidian-remote
description: Manage a remote Obsidian vault via MCP.
compatibility: Requires Obsidian Remote container with MCP enabled.
allowed-tools: mcp__obsidian-remote__*
---

# Obsidian Remote Skill

This skill enables interaction with a remote Obsidian vault using the Model Context Protocol.

## Tools

### Note Management

- `obsidian_read_note`: Retrieve note content and metadata.
- `obsidian_update_note`: Create, append, or overwrite notes.
- `obsidian_list_notes`: List files and folders.

### Search

- `obsidian_global_search`: Search for text or regex across the vault.
- `obsidian_search_replace`: Perform search-and-replace within a note.

### Metadata

- `obsidian_manage_frontmatter`: Atomic YAML key management.
- `obsidian_manage_tags`: Add or remove tags.

## Usage

Configure your MCP client to connect to the server's endpoint as described in the project `README.md`.

---
name: obsidian-remote
description: Manage a remote Obsidian vault via MCP.
compatibility: Requires Obsidian Remote container with MCP enabled.
allowed-tools: mcp__obsidian-remote__*
---

# Obsidian Remote Skill

This skill enables interaction with a remote Obsidian vault using the Model Context Protocol. The server is a Go bridge that translates MCP tool calls into HTTP requests for the Obsidian Local REST API.

## Configuration

The server is configured via environment variables. See `.env.example` for the full list. Key variables:

- `PUBLIC_HOST`: The external URL of your MCP server.
- `OAUTH_ISSUER`: Your OAuth provider's issuer URL.
- `OAUTH_JWKS_URL`: Your OAuth provider's JWKS endpoint.
- `OAUTH_AUDIENCE`: Your OAuth Client ID.
- `OAUTH_CLIENT_SECRET`: Your OAuth Client Secret (used server-side for the token exchange proxy).
- `OAUTH_ALLOWED_EMAIL`: Authorized email for access.

## Tools

### Note Management

- `read_note`: Retrieve note content and metadata.
- `update_note`: Create or overwrite notes.
- `append_note`: Append content to the end of an existing note.
- `delete_note`: Permanently delete a note.
- `list_notes`: List files and folders.

### Search

- `global_search`: Search for text or regex across the vault.
- `search_replace`: Perform search-and-replace within a note.

### Metadata

- `manage_frontmatter`: Atomic YAML key management.
- `manage_tags`: Add or remove tags.

## CRITICAL: Behavioral Rules

### Search Results Formatting

The `global_search` tool returns raw JSON. You MUST format these results into a clean, human-readable Markdown list.

- Group results by file.
- Provide a brief context snippet for each match.
- Use a concise summary if there are many matches.

### Show diffs for `search_replace`

You MUST display the old and new text **in your response text before invoking** `search_replace`. The user needs to see the change before approving the tool call.

Format:

**Before:**

```markdown
(exact old text)
```

**After:**

```markdown
(exact new text)
```

If the replacement is large (>30 lines), summarise the key changes in a bullet list instead. Never skip the diff — the user will reject the call without it.

## Usage

Configure your MCP client to connect to the server's endpoint. Both Streamable HTTP (`/mcp`) and SSE (`/sse`) transports are supported. See `references/mcp-setup.md` for client-specific examples.

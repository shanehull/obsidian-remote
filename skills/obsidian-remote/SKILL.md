---
name: obsidian-remote
description: Manage a remote Obsidian vault via MCP.
compatibility: Requires Obsidian Remote container with MCP enabled.
allowed-tools: mcp__obsidian-remote__*
---

# Obsidian Remote Skill

This skill enables interaction with a remote Obsidian vault using the Model Context Protocol. The server is a Go bridge that translates MCP tool calls into HTTP requests for the Obsidian Local REST API.

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

### Show Content/Diff Before Modification (ALWAYS)

You MUST display the content or diff in your response text **before invoking** any tool that modifies a note (`update_note`, `append_note`, `search_replace`). The user needs to see exactly what will be written before approving the tool call. **This is mandatory regardless of the size of the change.**

- **For `update_note` / `append_note`**: Display the full content or the block being written/appended.
- **For `search_replace`**: Display the old and new text.

Format for `search_replace`:

**Before:**

```markdown
(exact old text)
```

**After:**

```markdown
(exact new text)
```

Never skip this step — the user will reject the call without it.

### Search Results Formatting

Search results should be displayed in a readable manner, not in the raw JSON response format received.

When presenting search results:
- Extract and display the matched text snippets with context
- Show the file path and line number for each match
- Highlight the search term within the matched text if possible
- Avoid dumping raw JSON — parse and present as structured text

## Usage

Configure your MCP client to connect to the server's endpoint. Both Streamable HTTP (`/mcp`) and SSE (`/sse`) transports are supported.

- **Streamable HTTP (Gemini CLI):** Use `httpUrl` (e.g., `https://<server-url>/mcp`).
- **SSE (Cursor, Amp):** Use `url` (e.g., `https://<server-url>/sse`).

See `references/mcp-setup.md` for client-specific examples.

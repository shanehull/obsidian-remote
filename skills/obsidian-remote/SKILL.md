---
name: obsidian-remote
description: Manage an Obsidian vault — read, search, create, update, and delete notes with metadata tagging and full-text search. Use this skill when the user wants to interact with their Obsidian vault content, even if they don't explicitly say Obsidian or vault.
compatibility: Requires Obsidian Remote container with MCP enabled.
allowed-tools: mcp__obsidian-remote__*
---

# Obsidian Remote Skill

This skill enables interaction with a remote Obsidian vault via the Model Context Protocol.

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

### Confirmation Required for All Write Actions

Before invoking ANY tool that writes to a note (`update_note`, `append_note`, `delete_note`, `search_replace`), you MUST:

1. **Display the exact content or diff** in your response
2. **Ask for explicit confirmation** from the user
3. **WAIT** for the user to confirm before invoking the tool

**For `update_note`**: AVOID when possible. Use `search_replace` instead for targeted changes. If `update_note` is necessary, display a diff (old vs new) by using `read_note` first, then show only changed lines.

**For `append_note`**: Display only the block being appended.

**For `search_replace`**: Display the old and new text.

Format for `search_replace`:

**Before:**

```markdown
(exact old text)
```

**After:**

```markdown
(exact new text)
```

**Then ask:** "Proceed with this change?"

**Never skip confirmation** — the user will reject the call without it. No exceptions, regardless of change size.

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

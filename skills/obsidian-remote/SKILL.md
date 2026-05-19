---
name: obsidian-remote
description: Read, write, and search an Obsidian vault — surgically target specific headings, blocks, or frontmatter fields in notes without reading full files. Use this skill when the user wants to interact with their notes or vault, even if they don't name Obsidian.
compatibility: Requires Obsidian Remote container with MCP enabled.
allowed-tools: mcp__obsidian-remote__*
---

## Gotchas

- **`prepend` requires a target** — you cannot prepend to an entire file. Use `operation=append` with no target to add content at the end.
- **`target_type` and `target` are a pair** — providing one without the other is an error.
- **`target_scope` only applies to `heading` and `block`** — it is ignored for `frontmatter` targets.
- **Boolean params are strings** — `create_target_if_missing`, `reject_if_content_preexists`, and `trim_target_whitespace` expect `"true"` (a string), not `true` (a boolean).
- **Nested headings use `::`** — target `"Projects::Active"` not `"Projects/Active"` or `"Projects > Active"`.
- **`search_replace` count defaults to `1`** — only the first occurrence is replaced unless you set `count` to `-1`.

## Tools

| Tool                 | Writes? |
| -------------------- | ------- |
| `list_notes`         | No      |
| `read_note`          | No      |
| `global_search`      | No      |
| `update_note`        | Yes     |
| `delete_note`        | Yes     |
| `search_replace`     | Yes     |
| `manage_frontmatter` | Yes     |
| `manage_tags`        | Yes     |

### `read_note`

Read all or part of a note.

| Parameter          | Values                                            |
| ------------------ | ------------------------------------------------- |
| `target_type`      | `heading`, `block`, or `frontmatter`              |
| `target`           | Heading text, block ref, or frontmatter key       |
| `target_scope`     | `content` (default), `marker`, `markerAndContent` |
| `target_delimiter` | Default `::` — separates nested heading levels    |

Omit target params to read the full note. Nest headings with `::`, e.g. `"Section::Subsection"`.

### `update_note`

Create, overwrite, append, or prepend content within a note.

| Parameter                     | Values                                            |
| ----------------------------- | ------------------------------------------------- |
| `operation`                   | `replace` (default), `append`, or `prepend`       |
| `target_type`                 | `heading`, `block`, or `frontmatter`              |
| `target`                      | Heading text, block ref, or frontmatter key       |
| `target_scope`                | `content` (default), `marker`, `markerAndContent` |
| `target_delimiter`            | Default `::`                                      |
| `create_target_if_missing`    | `"true"` to create target if absent               |
| `reject_if_content_preexists` | `"true"` to reject if target has content          |
| `trim_target_whitespace`      | `"true"` to trim whitespace before operation      |

Behavior by combination:

| Operation | Target? | Result                                      |
| --------- | ------- | ------------------------------------------- |
| `replace` | No      | Overwrites the entire file.                 |
| `replace` | Yes     | Replaces only the target section's content. |
| `append`  | No      | Appends to end of file.                     |
| `append`  | Yes     | Appends within the target section.          |
| `prepend` | No      | **Error** — requires a target.              |
| `prepend` | Yes     | Prepends within the target section.         |

### `search_replace`

| Parameter | Description                                  |
| --------- | -------------------------------------------- |
| `count`   | Default `1` (first occurrence), `-1` for all |

### `manage_frontmatter`

| Parameter     | Values                |
| ------------- | --------------------- |
| `operation`   | `get` or `set`        |
| `jsonPayload` | JSON object for `set` |

### `manage_tags`

| Parameter   | Values                        |
| ----------- | ----------------------------- |
| `operation` | `add` or `remove`             |
| `tag`       | Tag value without leading `#` |

## Behavioral Rules

### Exclusive Write Path

Use **only** `obsidian-remote` tools for vault writes — never local `edit`/`write`/`bash`. Mixing paths on the same file creates git conflicts.

### Confirmation Required

Before any write (`update_note`, `delete_note`, `search_replace`):

1. Show the exact content or diff.
2. Ask for explicit confirmation.
3. Wait for the user to confirm.

- **`update_note` replace**: Show old vs new via `read_note` with same target, then diffs.
- **`update_note` append/prepend**: Show the block being inserted.
- **`search_replace`**: Show old and new text in "Before/After" format.

Never skip confirmation. No exceptions.

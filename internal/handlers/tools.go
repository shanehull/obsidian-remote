package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/shanehull/obsidian-remote/internal/obsidian"
)

func RegisterTools(s *server.MCPServer, client *obsidian.Client) {
	registerListNotes(s, client)
	registerReadNote(s, client)
	registerUpdateNote(s, client)
	registerDeleteNote(s, client)
	registerGlobalSearch(s, client)
	registerSearchReplace(s, client)
	registerManageFrontmatter(s, client)
	registerManageTags(s, client)
}

func normalizePath(path string) string {
	return strings.TrimPrefix(strings.TrimSuffix(path, "/"), "/")
}

func registerListNotes(s *server.MCPServer, client *obsidian.Client) {
	s.AddTool(mcp.NewTool("list_notes",
		mcp.WithDescription("List files in the vault"),
		mcp.WithString("dirPath", mcp.Description("Subdirectory")),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		subDir := normalizePath(req.GetString("dirPath", ""))
		res, err := client.Call("GET", "/vault/"+subDir, nil)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(string(res)), nil
	})
}

func registerReadNote(s *server.MCPServer, client *obsidian.Client) {
	s.AddTool(mcp.NewTool("read_note",
		mcp.WithDescription("Read a note"),
		mcp.WithString("path", mcp.Required()),
		mcp.WithString("target_type", mcp.Description("Section type: heading, block, or frontmatter")),
		mcp.WithString("target", mcp.Description("Heading text, block reference, or frontmatter key")),
		mcp.WithString("target_scope", mcp.Description("Scope for heading/block targets: content, marker, or markerAndContent")),
		mcp.WithString("target_delimiter", mcp.Description("Delimiter for nested headings (default: ::)")),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path, err := req.RequireString("path")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		path = normalizePath(path)

		headers, headerErr := buildTargetHeaders(req)
		if headerErr != nil {
			return mcp.NewToolResultError(headerErr.Error()), nil
		}
		res, err := client.Call("GET", "/vault/"+path, nil, headers)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(string(res)), nil
	})
}

func buildTargetHeaders(req mcp.CallToolRequest) (map[string]string, error) {
	targetType := req.GetString("target_type", "")
	target := req.GetString("target", "")
	if targetType == "" && target == "" {
		return nil, nil
	}
	if targetType == "" || target == "" {
		return nil, fmt.Errorf("both target_type and target are required when targeting a section")
	}
	if targetType != "heading" && targetType != "block" && targetType != "frontmatter" {
		return nil, fmt.Errorf("target_type must be 'heading', 'block', or 'frontmatter'")
	}
	headers := map[string]string{
		"Target-Type": targetType,
		"Target":      target,
	}
	if scope := req.GetString("target_scope", ""); scope != "" {
		if scope != "content" && scope != "marker" && scope != "markerAndContent" {
			return nil, fmt.Errorf("target_scope must be 'content', 'marker', or 'markerAndContent'")
		}
		headers["Target-Scope"] = scope
	}
	if delimiter := req.GetString("target_delimiter", ""); delimiter != "" {
		headers["Target-Delimiter"] = delimiter
	}
	if req.GetString("create_target_if_missing", "") == "true" {
		headers["Create-Target-If-Missing"] = "true"
	}
	if req.GetString("reject_if_content_preexists", "") == "true" {
		headers["Reject-If-Content-Preexists"] = "true"
	}
	if req.GetString("trim_target_whitespace", "") == "true" {
		headers["Trim-Target-Whitespace"] = "true"
	}
	return headers, nil
}

func registerUpdateNote(s *server.MCPServer, client *obsidian.Client) {
	s.AddTool(mcp.NewTool("update_note",
		mcp.WithDescription("Create, update, or append content within a note"),
		mcp.WithString("path", mcp.Required()),
		mcp.WithString("content", mcp.Required()),
		mcp.WithString("operation", mcp.Description("replace (default), append, or prepend")),
		mcp.WithString("target_type", mcp.Description("Section type: heading, block, or frontmatter")),
		mcp.WithString("target", mcp.Description("Heading text, block reference, or frontmatter key")),
		mcp.WithString("target_scope", mcp.Description("Scope for heading/block targets: content, marker, or markerAndContent")),
		mcp.WithString("target_delimiter", mcp.Description("Delimiter for nested headings (default: ::)")),
		mcp.WithString("create_target_if_missing", mcp.Description("Create the target if it does not exist (true or false)")),
		mcp.WithString("reject_if_content_preexists", mcp.Description("Reject if target already has content (true or false)")),
		mcp.WithString("trim_target_whitespace", mcp.Description("Trim whitespace from target content before operation (true or false)")),
		mcp.WithReadOnlyHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(true),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path, err := req.RequireString("path")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		path = normalizePath(path)
		content, err := req.RequireString("content")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		op := req.GetString("operation", "replace")
		if op != "replace" && op != "append" && op != "prepend" {
			return mcp.NewToolResultError("operation must be 'replace', 'append', or 'prepend'"), nil
		}
		targetType := req.GetString("target_type", "")
		target := req.GetString("target", "")
		hasTarget := targetType != "" && target != ""

		if op == "prepend" && !hasTarget {
			return mcp.NewToolResultError("prepend requires target_type and target"), nil
		}

		var method string
		headers := make(map[string]string)
		switch {
		case op == "replace" && !hasTarget:
			method = "PUT"
		case op == "replace" && hasTarget:
			method = "PUT"
			h, hErr := buildTargetHeaders(req)
			if hErr != nil {
				return mcp.NewToolResultError(hErr.Error()), nil
			}
			headers = h
		case op == "append" && !hasTarget:
			method = "POST"
		case op == "append" && hasTarget:
			method = "POST"
			h, hErr := buildTargetHeaders(req)
			if hErr != nil {
				return mcp.NewToolResultError(hErr.Error()), nil
			}
			headers = h
		case op == "prepend" && hasTarget:
			method = "PATCH"
			h, hErr := buildTargetHeaders(req)
			if hErr != nil {
				return mcp.NewToolResultError(hErr.Error()), nil
			}
			headers = h
			headers["Operation"] = "prepend"
		}

		_, err = client.Call(method, "/vault/"+path, []byte(content), headers)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Successfully %sed note: %s", op, path)), nil
	})
}

func registerDeleteNote(s *server.MCPServer, client *obsidian.Client) {
	s.AddTool(mcp.NewTool("delete_note",
		mcp.WithDescription("Permanently delete a note from the vault"),
		mcp.WithString("path", mcp.Required(), mcp.Description("Path to the note to delete")),
		mcp.WithReadOnlyHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(true),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path, err := req.RequireString("path")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		path = normalizePath(path)
		_, err = client.Call("DELETE", "/vault/"+path, nil)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Successfully deleted note: %s", path)), nil
	})
}

func registerGlobalSearch(s *server.MCPServer, client *obsidian.Client) {
	s.AddTool(mcp.NewTool("global_search",
		mcp.WithDescription("Search for text across all notes"),
		mcp.WithString("query", mcp.Required()),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		query, err := req.RequireString("query")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		res, err := client.Call("POST", "/search/simple/?query="+url.QueryEscape(query), nil)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(string(res)), nil
	})
}

func registerSearchReplace(s *server.MCPServer, client *obsidian.Client) {
	s.AddTool(mcp.NewTool("search_replace",
		mcp.WithDescription("Search and replace text within a specific note"),
		mcp.WithString("path", mcp.Required(), mcp.Description("Path to the note")),
		mcp.WithString("search", mcp.Required(), mcp.Description("Text to find")),
		mcp.WithString("replace", mcp.Required(), mcp.Description("Replacement text")),
		mcp.WithNumber("count", mcp.Description("Maximum occurrences to replace (default: 1, set to -1 for all)")),
		mcp.WithReadOnlyHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(true),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path, err := req.RequireString("path")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		path = normalizePath(path)
		search, err := req.RequireString("search")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		replace, err := req.RequireString("replace")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		count := req.GetInt("count", 1)

		content, err := client.Call("GET", "/vault/"+path, nil)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		original := string(content)
		if !strings.Contains(original, search) {
			return mcp.NewToolResultError("search text not found in note"), nil
		}

		updated := strings.Replace(original, search, replace, count)

		replaced := count
		if count < 0 {
			replaced = strings.Count(original, search)
		}

		if _, err := client.Call("PUT", "/vault/"+path, []byte(updated)); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully replaced %d occurrence(s) in %s", replaced, path)), nil
	})
}

func normalizeTags(tags []string) []string {
	result := make([]string, 0, len(tags))
	for _, t := range tags {
		result = append(result, strings.TrimPrefix(t, "#"))
	}
	return result
}

func addTagToSlice(tags []string, tag string) ([]string, bool) {
	for _, t := range tags {
		if t == tag {
			return tags, false
		}
	}
	return append(tags, tag), true
}

func removeTagFromSlice(tags []string, tag string) ([]string, bool) {
	found := false
	filtered := tags[:0]
	for _, t := range tags {
		if t == tag {
			found = true
			continue
		}
		filtered = append(filtered, t)
	}
	if !found {
		return tags, false
	}
	return filtered, true
}

func registerManageTags(s *server.MCPServer, client *obsidian.Client) {
	s.AddTool(mcp.NewTool("manage_tags",
		mcp.WithDescription("Add or remove tags from a note's frontmatter"),
		mcp.WithString("path", mcp.Required(), mcp.Description("Path to the note")),
		mcp.WithString("operation", mcp.Required(), mcp.Description("add or remove")),
		mcp.WithString("tag", mcp.Required(), mcp.Description("Tag value (without leading #)")),
		mcp.WithReadOnlyHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(true),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path, err := req.RequireString("path")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		path = normalizePath(path)
		op, err := req.RequireString("operation")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		tag, err := req.RequireString("tag")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		tag = strings.TrimPrefix(tag, "#")

		if op != "add" && op != "remove" {
			return mcp.NewToolResultError("operation must be 'add' or 'remove'"), nil
		}

		res, err := client.Call("GET", "/vault/"+path, nil,
			map[string]string{"Accept": "application/vnd.olrapi.note+json"})
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		var note struct {
			Tags []string `json:"tags"`
		}
		if jsonErr := json.Unmarshal(res, &note); jsonErr != nil {
			return mcp.NewToolResultError("failed to parse note metadata: " + jsonErr.Error()), nil
		}

		tags := normalizeTags(note.Tags)

		var ok bool
		switch op {
		case "add":
			tags, ok = addTagToSlice(tags, tag)
			if !ok {
				return mcp.NewToolResultText(fmt.Sprintf("Tag '%s' already exists in %s", tag, path)), nil
			}
		case "remove":
			tags, ok = removeTagFromSlice(tags, tag)
			if !ok {
				return mcp.NewToolResultError(fmt.Sprintf("Tag '%s' not found in %s", tag, path)), nil
			}
		}

		tagsJSON, err := json.Marshal(tags)
		if err != nil {
			return mcp.NewToolResultError("failed to marshal tags: " + err.Error()), nil
		}
		if _, err := client.Call("PATCH", "/vault/"+path, tagsJSON,
			map[string]string{
				"Content-Type":             "application/json",
				"Operation":                "replace",
				"Target-Type":              "frontmatter",
				"Target":                   "tags",
				"Create-Target-If-Missing": "true",
			}); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		msg := fmt.Sprintf("Successfully added tag '%s' to %s", tag, path)
		if op == "remove" {
			msg = fmt.Sprintf("Successfully removed tag '%s' from %s", tag, path)
		}
		return mcp.NewToolResultText(msg), nil
	})
}

func registerManageFrontmatter(s *server.MCPServer, client *obsidian.Client) {
	s.AddTool(mcp.NewTool("manage_frontmatter",
		mcp.WithDescription("Get or set YAML frontmatter keys"),
		mcp.WithString("path", mcp.Required()),
		mcp.WithString("operation", mcp.Required(), mcp.Description("get or set")),
		mcp.WithString("jsonPayload", mcp.Description("JSON object of keys to set (required for 'set')")),
		mcp.WithReadOnlyHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(true),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path, err := req.RequireString("path")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		path = normalizePath(path)
		op, err := req.RequireString("operation")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		if op == "get" {
			res, err := client.Call("GET", "/vault/"+path, nil,
				map[string]string{"Accept": "application/vnd.olrapi.note+json"})
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(string(res)), nil
		}

		if op == "set" {
			payload, err := req.RequireString("jsonPayload")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			var kvs map[string]interface{}
			if jsonErr := json.Unmarshal([]byte(payload), &kvs); jsonErr != nil {
				return mcp.NewToolResultError("jsonPayload must be a valid JSON object: " + jsonErr.Error()), nil
			}

			var errs []string
			for k, v := range kvs {
				b, marshalErr := json.Marshal(v)
				if marshalErr != nil {
					errs = append(errs, k+": failed to marshal value: "+marshalErr.Error())
					continue
				}
				_, patchErr := client.Call("PATCH", "/vault/"+path, b,
					map[string]string{
						"Content-Type":             "application/json",
						"Operation":                "replace",
						"Target-Type":              "frontmatter",
						"Target":                   k,
						"Create-Target-If-Missing": "true",
					})
				if patchErr != nil {
					errs = append(errs, k+": "+patchErr.Error())
				}
			}
			if len(errs) > 0 {
				return mcp.NewToolResultError("errors: " + strings.Join(errs, "; ")), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("Successfully updated frontmatter for: %s", path)), nil
		}

		return mcp.NewToolResultError("Invalid operation. Use 'get' or 'set'."), nil
	})
}

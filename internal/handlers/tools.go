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
	registerGlobalSearch(s, client)
	registerSearchReplace(s, client)
	registerManageFrontmatter(s, client)
	registerManageTags(s, client)
}

func registerListNotes(s *server.MCPServer, client *obsidian.Client) {
	s.AddTool(mcp.NewTool("obsidian_list_notes",
		mcp.WithDescription("List files in the vault"),
		mcp.WithString("dirPath", mcp.Description("Subdirectory")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		subDir := req.GetString("dirPath", "")
		res, err := client.Call("GET", "/vault/"+subDir, nil)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(string(res)), nil
	})
}

func registerReadNote(s *server.MCPServer, client *obsidian.Client) {
	s.AddTool(mcp.NewTool("obsidian_read_note",
		mcp.WithDescription("Read a note"),
		mcp.WithString("path", mcp.Required()),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path, err := req.RequireString("path")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		res, err := client.Call("GET", "/vault/"+path, nil)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(string(res)), nil
	})
}

func registerUpdateNote(s *server.MCPServer, client *obsidian.Client) {
	s.AddTool(mcp.NewTool("obsidian_update_note",
		mcp.WithDescription("Create or update a note"),
		mcp.WithString("path", mcp.Required()),
		mcp.WithString("content", mcp.Required()),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path, err := req.RequireString("path")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		content, err := req.RequireString("content")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		res, err := client.Call("PUT", "/vault/"+path, []byte(content))
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(string(res)), nil
	})
}

func registerGlobalSearch(s *server.MCPServer, client *obsidian.Client) {
	s.AddTool(mcp.NewTool("obsidian_global_search",
		mcp.WithDescription("Search for text across all notes"),
		mcp.WithString("query", mcp.Required()),
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
	s.AddTool(mcp.NewTool("obsidian_search_replace",
		mcp.WithDescription("Search and replace text within a specific note"),
		mcp.WithString("path", mcp.Required(), mcp.Description("Path to the note")),
		mcp.WithString("search", mcp.Required(), mcp.Description("Text to find")),
		mcp.WithString("replace", mcp.Required(), mcp.Description("Replacement text")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path, err := req.RequireString("path")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		search, err := req.RequireString("search")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		replace, err := req.RequireString("replace")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		content, err := client.Call("GET", "/vault/"+path, nil)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		original := string(content)
		if !strings.Contains(original, search) {
			return mcp.NewToolResultError("search text not found in note"), nil
		}

		updated := strings.ReplaceAll(original, search, replace)
		count := strings.Count(original, search)

		if _, err := client.Call("PUT", "/vault/"+path, []byte(updated)); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("replaced %d occurrence(s)", count)), nil
	})
}

func registerManageTags(s *server.MCPServer, client *obsidian.Client) {
	s.AddTool(mcp.NewTool("obsidian_manage_tags",
		mcp.WithDescription("Add or remove tags from a note's frontmatter"),
		mcp.WithString("path", mcp.Required(), mcp.Description("Path to the note")),
		mcp.WithString("operation", mcp.Required(), mcp.Description("add or remove")),
		mcp.WithString("tag", mcp.Required(), mcp.Description("Tag value (without leading #)")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path, err := req.RequireString("path")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
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

		var tags []string
		for _, t := range note.Tags {
			tags = append(tags, strings.TrimPrefix(t, "#"))
		}

		switch op {
		case "add":
			for _, t := range tags {
				if t == tag {
					return mcp.NewToolResultText("tag already exists"), nil
				}
			}
			tags = append(tags, tag)
		case "remove":
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
				return mcp.NewToolResultError("tag not found"), nil
			}
			tags = filtered
		}

		tagsJSON, _ := json.Marshal(tags)
		if _, err := client.Call("PATCH", "/vault/"+path, tagsJSON,
			map[string]string{
				"Content-Type":            "application/json",
				"Operation":               "replace",
				"Target-Type":             "frontmatter",
				"Target":                  "tags",
				"Create-Target-If-Missing": "true",
			}); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		msg := fmt.Sprintf("tag '%s' added", tag)
		if op == "remove" {
			msg = fmt.Sprintf("tag '%s' removed", tag)
		}
		return mcp.NewToolResultText(msg), nil
	})
}

func registerManageFrontmatter(s *server.MCPServer, client *obsidian.Client) {
	s.AddTool(mcp.NewTool("obsidian_manage_frontmatter",
		mcp.WithDescription("Get or set YAML frontmatter keys"),
		mcp.WithString("path", mcp.Required()),
		mcp.WithString("operation", mcp.Required(), mcp.Description("get or set")),
		mcp.WithString("jsonPayload", mcp.Description("JSON object of keys to set (required for 'set')")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path, err := req.RequireString("path")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
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
				b, _ := json.Marshal(v)
				_, patchErr := client.Call("PATCH", "/vault/"+path, b,
					map[string]string{
						"Content-Type":            "application/json",
						"Operation":               "replace",
						"Target-Type":             "frontmatter",
						"Target":                  k,
						"Create-Target-If-Missing": "true",
					})
				if patchErr != nil {
					errs = append(errs, k+": "+patchErr.Error())
				}
			}
			if len(errs) > 0 {
				return mcp.NewToolResultError("errors: " + strings.Join(errs, "; ")), nil
			}
			return mcp.NewToolResultText("frontmatter updated"), nil
		}

		return mcp.NewToolResultError("Invalid operation. Use 'get' or 'set'."), nil
	})
}

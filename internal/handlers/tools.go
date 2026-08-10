package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/shanehull/obsidian-remote/internal/obsidian"
)

func RegisterTools(s *mcp.Server, client *obsidian.Client) {
	registerListNotes(s, client)
	registerReadNote(s, client)
	registerUpdateNote(s, client)
	registerDeleteNote(s, client)
	registerMoveNote(s, client)
	registerGlobalSearch(s, client)
	registerSearchReplace(s, client)
	registerManageFrontmatter(s, client)
	registerManageTags(s, client)
}

var validTargetTypes = map[string]bool{
	"heading":     true,
	"block":       true,
	"frontmatter": true,
}

var validTargetScopes = map[string]bool{
	"content":          true,
	"marker":           true,
	"markerAndContent": true,
}

func normalizePath(path string) string {
	return strings.TrimPrefix(strings.TrimSuffix(path, "/"), "/")
}

func encodePath(p string) string {
	parts := strings.Split(p, "/")
	for i, part := range parts {
		parts[i] = url.PathEscape(part)
	}
	return strings.Join(parts, "/")
}

func unmarshalArgs(req *mcp.CallToolRequest) (map[string]any, error) {
	var args map[string]any
	if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
		return nil, err
	}
	return args, nil
}

func getStringArg(req *mcp.CallToolRequest, key, defaultVal string) string {
	args, _ := unmarshalArgs(req)
	if v, ok := args[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return defaultVal
}

func requireStringArg(req *mcp.CallToolRequest, key string) (string, error) {
	args, err := unmarshalArgs(req)
	if err != nil {
		return "", err
	}
	v, ok := args[key]
	if !ok {
		return "", fmt.Errorf("missing required parameter: %s", key)
	}
	s, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("parameter %s must be a string", key)
	}
	return s, nil
}

func getIntArg(req *mcp.CallToolRequest, key string, defaultVal int) int {
	args, _ := unmarshalArgs(req)
	if v, ok := args[key]; ok {
		switch n := v.(type) {
		case float64:
			return int(n)
		}
	}
	return defaultVal
}

func boolPtr(b bool) *bool { return &b }

func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}
}

func errorResult(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{&mcp.TextContent{Text: msg}},
	}
}

func handleListNotes(client *obsidian.Client) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		subDir := normalizePath(getStringArg(req, "dirPath", ""))
		res, err := client.Call("GET", "/vault/"+subDir, nil)
		if err != nil {
			return errorResult(err.Error()), nil
		}
		return textResult(string(res)), nil
	}
}

func registerListNotes(s *mcp.Server, client *obsidian.Client) {
	s.AddTool(&mcp.Tool{
		Name:        "list_notes",
		Description: "List files and directories in the vault. Use dirPath to list a specific subdirectory. Returns an array of filenames.",
		InputSchema: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"dirPath": {Type: "string", Description: "Subdirectory"},
			},
		},
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, DestructiveHint: boolPtr(false)},
	}, handleListNotes(client))
}

func parseTarget(req *mcp.CallToolRequest) (targetType, target string, err error) {
	targetType = getStringArg(req, "target_type", "")
	target = getStringArg(req, "target", "")
	if targetType == "" && target == "" {
		return "", "", nil
	}
	if targetType == "" || target == "" {
		return "", "", fmt.Errorf("both target_type and target are required when targeting a section")
	}
	if !validTargetTypes[targetType] {
		return "", "", fmt.Errorf("target_type must be 'heading', 'block', or 'frontmatter'")
	}
	return targetType, target, nil
}

func targetPathSegment(req *mcp.CallToolRequest) (string, error) {
	targetType, target, err := parseTarget(req)
	if err != nil {
		return "", err
	}
	if targetType == "" {
		return "", nil
	}
	delimiter := getStringArg(req, "target_delimiter", "")
	if delimiter == "" {
		delimiter = "::"
	}
	segments := targetSegments(target, delimiter)
	for i, s := range segments {
		segments[i] = url.PathEscape(s)
	}
	return "/" + targetType + "/" + strings.Join(segments, "/"), nil
}

func targetSegments(target, delimiter string) []string {
	if delimiter == "" {
		delimiter = "::"
	}
	return strings.Split(target, delimiter)
}

func patchBody(targetType, target, delimiter, operation, content string, value any, req *mcp.CallToolRequest) ([]byte, error) {
	var t any
	if targetType == "heading" {
		t = targetSegments(target, delimiter)
	} else {
		t = target
	}
	body := map[string]any{
		"targetType": targetType,
		"target":     t,
		"operation":  operation,
	}
	if value != nil {
		body["value"] = value
	} else {
		body["content"] = content
	}
	if scope := getStringArg(req, "target_scope", ""); scope != "" {
		body["targetScope"] = scope
	}
	if getStringArg(req, "create_target_if_missing", "") == "true" {
		body["createTargetIfMissing"] = true
	}
	if getStringArg(req, "reject_if_content_preexists", "") == "true" {
		body["rejectIfContentPreexists"] = true
	}
	if getStringArg(req, "trim_target_whitespace", "") == "true" {
		body["trimTargetWhitespace"] = true
	}
	return json.Marshal(body)
}

func handleReadNote(client *obsidian.Client) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path, err := requireStringArg(req, "path")
		if err != nil {
			return errorResult(err.Error()), nil
		}
		path = normalizePath(path)

		targetPath, err := targetPathSegment(req)
		if err != nil {
			return errorResult(err.Error()), nil
		}
		var headers map[string]string
		if scope := getStringArg(req, "target_scope", ""); scope != "" {
			if !validTargetScopes[scope] {
				return errorResult("target_scope must be 'content', 'marker', or 'markerAndContent'"), nil
			}
			headers = map[string]string{"Target-Scope": scope}
		}
		res, err := client.Call("GET", "/vault/"+path+targetPath, nil, headers)
		if err != nil {
			return errorResult(err.Error()), nil
		}
		return textResult(string(res)), nil
	}
}

func registerReadNote(s *mcp.Server, client *obsidian.Client) {
	s.AddTool(&mcp.Tool{
		Name:        "read_note",
		Description: "Read a note or a specific section within it. Use target_type=heading with target=heading text to read one section, target_type=block for block references, or target_type=frontmatter for a YAML key. Omitting target returns the full note.",
		InputSchema: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"path":             {Type: "string"},
				"target_type":      {Type: "string", Description: "Section type: heading, block, or frontmatter"},
				"target":           {Type: "string", Description: "Heading text, block reference, or frontmatter key"},
				"target_scope":     {Type: "string", Description: "Scope for heading/block targets: content, marker, or markerAndContent"},
				"target_delimiter": {Type: "string", Description: "Delimiter for nested headings (default: ::)"},
			},
			Required: []string{"path"},
		},
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, DestructiveHint: boolPtr(false)},
	}, handleReadNote(client))
}

func resolveUpdateMethod(op string, hasTarget bool) (string, error) {
	switch op {
	case "replace":
		if hasTarget {
			return "PATCH", nil
		}
		return "PUT", nil
	case "append":
		if hasTarget {
			return "PATCH", nil
		}
		return "POST", nil
	case "prepend":
		if !hasTarget {
			return "", fmt.Errorf("prepend requires target_type and target")
		}
		return "PATCH", nil
	default:
		return "", fmt.Errorf("operation must be 'replace', 'append', or 'prepend'")
	}
}

type updateNoteParams struct {
	path   string
	method string
	body   []byte
}

func parseUpdateNote(req *mcp.CallToolRequest) (*updateNoteParams, error) {
	path, err := requireStringArg(req, "path")
	if err != nil {
		return nil, err
	}
	path = normalizePath(path)
	content, err := requireStringArg(req, "content")
	if err != nil {
		return nil, err
	}
	op := getStringArg(req, "operation", "replace")
	targetType, target, err := parseTarget(req)
	if err != nil {
		return nil, err
	}
	hasTarget := targetType != ""

	method, err := resolveUpdateMethod(op, hasTarget)
	if err != nil {
		return nil, err
	}

	var body []byte
	if hasTarget {
		delimiter := getStringArg(req, "target_delimiter", "")
		body, err = patchBody(targetType, target, delimiter, op, content, nil, req)
		if err != nil {
			return nil, err
		}
	} else {
		body = []byte(content)
	}

	return &updateNoteParams{path, method, body}, nil
}

func handleUpdateNote(client *obsidian.Client) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		params, err := parseUpdateNote(req)
		if err != nil {
			return errorResult(err.Error()), nil
		}
	var headers map[string]string
	if params.method == "PATCH" {
		headers = map[string]string{"Content-Type": "application/json"}
	}
	_, err = client.Call(params.method, "/vault/"+params.path, params.body, headers)
		if err != nil {
			return errorResult(err.Error()), nil
		}
		return textResult(fmt.Sprintf("Successfully %sed note: %s", getStringArg(req, "operation", "replace"), params.path)), nil
	}
}

func registerUpdateNote(s *mcp.Server, client *obsidian.Client) {
	s.AddTool(&mcp.Tool{
		Name:        "update_note",
		Description: "Create, update, or append content within a note. Read the note first with read_note, show the current content to the user as a preview of what will change, and confirm before making any modifications.",
		InputSchema: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"path":                         {Type: "string"},
				"content":                      {Type: "string"},
				"operation":                    {Type: "string", Description: "replace (default), append, or prepend"},
				"target_type":                  {Type: "string", Description: "Section type: heading, block, or frontmatter"},
				"target":                       {Type: "string", Description: "Heading text, block reference, or frontmatter key"},
				"target_scope":                 {Type: "string", Description: "Scope for heading/block targets: content, marker, or markerAndContent"},
				"target_delimiter":             {Type: "string", Description: "Delimiter for nested headings (default: ::)"},
				"create_target_if_missing":     {Type: "string", Description: "Create the target if it does not exist (true or false)"},
				"reject_if_content_preexists":  {Type: "string", Description: "Reject if target already has content (true or false)"},
				"trim_target_whitespace":       {Type: "string", Description: "Trim whitespace from target content before operation (true or false)"},
			},
			Required: []string{"path", "content"},
		},
		Annotations: &mcp.ToolAnnotations{DestructiveHint: boolPtr(true)},
	}, handleUpdateNote(client))
}

func handleDeleteNote(client *obsidian.Client) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path, err := requireStringArg(req, "path")
		if err != nil {
			return errorResult(err.Error()), nil
		}
		path = normalizePath(path)
		_, err = client.Call("DELETE", "/vault/"+path, nil)
		if err != nil {
			return errorResult(err.Error()), nil
		}
		return textResult(fmt.Sprintf("Successfully deleted note: %s", path)), nil
	}
}

func registerDeleteNote(s *mcp.Server, client *obsidian.Client) {
	s.AddTool(&mcp.Tool{
		Name:        "delete_note",
		Description: "Permanently delete a note from the vault. This cannot be undone. Read the note first with read_note, show its content to the user, and confirm before deleting.",
		InputSchema: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"path": {Type: "string", Description: "Path to the note to delete"},
			},
			Required: []string{"path"},
		},
		Annotations: &mcp.ToolAnnotations{DestructiveHint: boolPtr(true)},
	}, handleDeleteNote(client))
}

func handleMoveNote(client *obsidian.Client) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path, err := requireStringArg(req, "path")
		if err != nil {
			return errorResult(err.Error()), nil
		}
		newPath, err := requireStringArg(req, "newPath")
		if err != nil {
			return errorResult(err.Error()), nil
		}
		path = normalizePath(path)
		newPath = normalizePath(newPath)

		headers := map[string]string{
			"Destination": encodePath(newPath),
		}
		if getStringArg(req, "allowOverwrite", "") == "true" {
			headers["Allow-Overwrite"] = "true"
		}

		_, err = client.Call("MOVE", "/vault/"+path, nil, headers)
		if err != nil {
			return errorResult(err.Error()), nil
		}
		return textResult(fmt.Sprintf("Successfully moved note from %s to %s", path, newPath)), nil
	}
}

func registerMoveNote(s *mcp.Server, client *obsidian.Client) {
	s.AddTool(&mcp.Tool{
		Name:        "move_note",
		Description: "Move or rename a note. Read the source note first with read_note, show it to the user, and confirm before moving. Read the destination note first if allowOverwrite is true to avoid accidental overwrite.",
		InputSchema: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"path":           {Type: "string", Description: "Current path of the note"},
				"newPath":        {Type: "string", Description: "New path for the note"},
				"allowOverwrite": {Type: "string", Description: "Allow overwrite if destination exists (true or false, default false)"},
			},
			Required: []string{"path", "newPath"},
		},
		Annotations: &mcp.ToolAnnotations{DestructiveHint: boolPtr(true)},
	}, handleMoveNote(client))
}

func handleGlobalSearch(client *obsidian.Client) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		query, err := requireStringArg(req, "query")
		if err != nil {
			return errorResult(err.Error()), nil
		}
		res, err := client.Call("POST", "/search/simple/?query="+url.QueryEscape(query), nil)
		if err != nil {
			return errorResult(err.Error()), nil
		}
		return textResult(string(res)), nil
	}
}

func registerGlobalSearch(s *mcp.Server, client *obsidian.Client) {
	s.AddTool(&mcp.Tool{
		Name:        "global_search",
		Description: "Search for text across all notes",
		InputSchema: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"query": {Type: "string"},
			},
			Required: []string{"query"},
		},
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, DestructiveHint: boolPtr(false)},
	}, handleGlobalSearch(client))
}

func handleSearchReplace(client *obsidian.Client) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path, err := requireStringArg(req, "path")
		if err != nil {
			return errorResult(err.Error()), nil
		}
		path = normalizePath(path)
		search, err := requireStringArg(req, "search")
		if err != nil {
			return errorResult(err.Error()), nil
		}
		replace, err := requireStringArg(req, "replace")
		if err != nil {
			return errorResult(err.Error()), nil
		}
		count := getIntArg(req, "count", 1)

		content, err := client.Call("GET", "/vault/"+path, nil)
		if err != nil {
			return errorResult(err.Error()), nil
		}

		original := string(content)
		if !strings.Contains(original, search) {
			return errorResult("search text not found in note"), nil
		}

		updated := strings.Replace(original, search, replace, count)

		replaced := count
		if count < 0 {
			replaced = strings.Count(original, search)
		}

		if _, err := client.Call("PUT", "/vault/"+path, []byte(updated)); err != nil {
			return errorResult(err.Error()), nil
		}

		return textResult(fmt.Sprintf("Successfully replaced %d occurrence(s) in %s", replaced, path)), nil
	}
}

func registerSearchReplace(s *mcp.Server, client *obsidian.Client) {
	s.AddTool(&mcp.Tool{
		Name:        "search_replace",
		Description: "Search and replace text within a note. Read the note first with read_note, show the matching text to the user, and confirm before replacing. Replaces only within a single note.",
		InputSchema: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"path":    {Type: "string", Description: "Path to the note"},
				"search":  {Type: "string", Description: "Text to find"},
				"replace": {Type: "string", Description: "Replacement text"},
				"count":   {Type: "number", Description: "Maximum occurrences to replace (default: 1, set to -1 for all)"},
			},
			Required: []string{"path", "search", "replace"},
		},
		Annotations: &mcp.ToolAnnotations{DestructiveHint: boolPtr(true)},
	}, handleSearchReplace(client))
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

func handleManageTags(client *obsidian.Client) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path, err := requireStringArg(req, "path")
		if err != nil {
			return errorResult(err.Error()), nil
		}
		path = normalizePath(path)
		op, err := requireStringArg(req, "operation")
		if err != nil {
			return errorResult(err.Error()), nil
		}
		tag, err := requireStringArg(req, "tag")
		if err != nil {
			return errorResult(err.Error()), nil
		}
		tag = strings.TrimPrefix(tag, "#")

		if op != "add" && op != "remove" {
			return errorResult("operation must be 'add' or 'remove'"), nil
		}

		res, err := client.Call("GET", "/vault/"+path, nil,
			map[string]string{"Accept": "application/vnd.olrapi.note+json"})
		if err != nil {
			return errorResult(err.Error()), nil
		}

		var note struct {
			Tags []string `json:"tags"`
		}
		if jsonErr := json.Unmarshal(res, &note); jsonErr != nil {
			return errorResult("failed to parse note metadata: " + jsonErr.Error()), nil
		}

		tags := normalizeTags(note.Tags)

		var ok bool
		switch op {
		case "add":
			tags, ok = addTagToSlice(tags, tag)
			if !ok {
				return textResult(fmt.Sprintf("Tag '%s' already exists in %s", tag, path)), nil
			}
		case "remove":
			tags, ok = removeTagFromSlice(tags, tag)
			if !ok {
				return errorResult(fmt.Sprintf("Tag '%s' not found in %s", tag, path)), nil
			}
		}

		body, err := json.Marshal(map[string]any{
			"targetType":             "frontmatter",
			"target":                 "tags",
			"operation":              "replace",
			"value":                  tags,
			"createTargetIfMissing":  true,
		})
		if err != nil {
			return errorResult("failed to marshal patch body: " + err.Error()), nil
		}
		if _, err := client.Call("PATCH", "/vault/"+path, body,
			map[string]string{"Content-Type": "application/json"}); err != nil {
			return errorResult(err.Error()), nil
		}

		msg := fmt.Sprintf("Successfully added tag '%s' to %s", tag, path)
		if op == "remove" {
			msg = fmt.Sprintf("Successfully removed tag '%s' from %s", tag, path)
		}
		return textResult(msg), nil
	}
}

func registerManageTags(s *mcp.Server, client *obsidian.Client) {
	s.AddTool(&mcp.Tool{
		Name:        "manage_tags",
		Description: "Add or remove tags from a note's frontmatter. Use read_note or manage_frontmatter with operation=get to view the current tags first, show them to the user, and confirm before modifying.",
		InputSchema: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"path":      {Type: "string", Description: "Path to the note"},
				"operation": {Type: "string", Description: "add or remove"},
				"tag":       {Type: "string", Description: "Tag value (without leading #)"},
			},
			Required: []string{"path", "operation", "tag"},
		},
		Annotations: &mcp.ToolAnnotations{DestructiveHint: boolPtr(true)},
	}, handleManageTags(client))
}

func handleManageFrontmatter(client *obsidian.Client) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path, err := requireStringArg(req, "path")
		if err != nil {
			return errorResult(err.Error()), nil
		}
		path = normalizePath(path)
		op, err := requireStringArg(req, "operation")
		if err != nil {
			return errorResult(err.Error()), nil
		}

		if op == "get" {
			res, err := client.Call("GET", "/vault/"+path, nil,
				map[string]string{"Accept": "application/vnd.olrapi.note+json"})
			if err != nil {
				return errorResult(err.Error()), nil
			}
			return textResult(string(res)), nil
		}

		if op == "set" {
			payload, err := requireStringArg(req, "jsonPayload")
			if err != nil {
				return errorResult(err.Error()), nil
			}

			var kvs map[string]any
			if jsonErr := json.Unmarshal([]byte(payload), &kvs); jsonErr != nil {
				return errorResult("jsonPayload must be a valid JSON object: " + jsonErr.Error()), nil
			}

			var errs []string
			for k, v := range kvs {
				body, marshalErr := json.Marshal(map[string]any{
					"targetType":             "frontmatter",
					"target":                 k,
					"operation":              "replace",
					"value":                  v,
					"createTargetIfMissing":  true,
				})
				if marshalErr != nil {
					errs = append(errs, k+": failed to marshal patch body: "+marshalErr.Error())
					continue
				}
				_, patchErr := client.Call("PATCH", "/vault/"+path, body,
					map[string]string{"Content-Type": "application/json"})
				if patchErr != nil {
					errs = append(errs, k+": "+patchErr.Error())
				}
			}
			if len(errs) > 0 {
				return errorResult("errors: " + strings.Join(errs, "; ")), nil
			}
			return textResult(fmt.Sprintf("Successfully updated frontmatter for: %s", path)), nil
		}

		return errorResult("Invalid operation. Use 'get' or 'set'."), nil
	}
}

func registerManageFrontmatter(s *mcp.Server, client *obsidian.Client) {
	s.AddTool(&mcp.Tool{
		Name:        "manage_frontmatter",
		Description: "Get or set YAML frontmatter keys. Use operation=get first to read the current frontmatter, show it to the user, and confirm before using operation=set to make changes.",
		InputSchema: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"path":        {Type: "string"},
				"operation":   {Type: "string", Description: "get or set"},
				"jsonPayload": {Type: "string", Description: "JSON object of keys to set (required for 'set')"},
			},
			Required: []string{"path", "operation"},
		},
		Annotations: &mcp.ToolAnnotations{DestructiveHint: boolPtr(true)},
	}, handleManageFrontmatter(client))
}

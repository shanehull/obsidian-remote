package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/shanehull/obsidian-remote/internal/config"
	"github.com/shanehull/obsidian-remote/internal/obsidian"
)

func TestNormalizeTags(t *testing.T) {
	tests := []struct {
		name string
		tags []string
		want []string
	}{
		{"empty", nil, nil},
		{"no_hash", []string{"tag1", "tag2"}, []string{"tag1", "tag2"}},
		{"with_hash", []string{"#tag1", "#tag2"}, []string{"tag1", "tag2"}},
		{"mixed", []string{"#tag1", "tag2", "#tag3"}, []string{"tag1", "tag2", "tag3"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeTags(tt.tags)
			if tt.tags == nil && got == nil {
				return
			}
			if len(got) != len(tt.want) {
				t.Fatalf("normalizeTags(%v) = %v, want %v", tt.tags, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("normalizeTags(%v) = %v, want %v", tt.tags, got, tt.want)
				}
			}
		})
	}
}

func TestAddTagToSlice(t *testing.T) {
	tests := []struct {
		name   string
		tags   []string
		tag    string
		wantOK bool
	}{
		{"new_tag", []string{"a", "b"}, "c", true},
		{"existing_tag", []string{"a", "b"}, "a", false},
		{"empty_slice", nil, "a", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := addTagToSlice(tt.tags, tt.tag)
			if ok != tt.wantOK {
				t.Fatalf("addTagToSlice(%v, %q) ok = %v, want %v", tt.tags, tt.tag, ok, tt.wantOK)
			}
			if ok {
				found := false
				for _, r := range result {
					if r == tt.tag {
						found = true
						break
					}
				}
				if !found {
					t.Fatalf("addTagToSlice(%v, %q) = %v, missing new tag", tt.tags, tt.tag, result)
				}
				if len(result) != len(tt.tags)+1 {
					t.Fatalf("addTagToSlice(%v, %q) len = %d, want %d", tt.tags, tt.tag, len(result), len(tt.tags)+1)
				}
			}
		})
	}
}

func TestRemoveTagFromSlice(t *testing.T) {
	tests := []struct {
		name   string
		tags   []string
		tag    string
		wantOK bool
	}{
		{"existing_tag", []string{"a", "b"}, "a", true},
		{"missing_tag", []string{"a", "b"}, "c", false},
		{"last_tag", []string{"a"}, "a", true},
		{"multiple_occurrences", []string{"a", "b", "a"}, "a", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := removeTagFromSlice(tt.tags, tt.tag)
			if ok != tt.wantOK {
				t.Fatalf("removeTagFromSlice(%v, %q) ok = %v, want %v", tt.tags, tt.tag, ok, tt.wantOK)
			}
			if ok {
				for _, r := range result {
					if r == tt.tag {
						t.Fatalf("removeTagFromSlice(%v, %q) = %v, tag still present", tt.tags, tt.tag, result)
					}
				}
			}
		})
	}
}

func assertHeader(t *testing.T, headers map[string]string, key, want string) {
	t.Helper()
	if want != "" && headers[key] != want {
		t.Fatalf("%s = %q, want %q", key, headers[key], want)
	}
}

type buildTargetHeadersTestCase struct {
	name    string
	args    map[string]any
	wantNil bool
	wantErr bool
	want    map[string]string
}

func testBuildTargetHeadersCase(t *testing.T, tt buildTargetHeadersTestCase) {
	t.Helper()
	rawArgs, _ := json.Marshal(tt.args)
	req := &mcp.CallToolRequest{Params: &mcp.CallToolParamsRaw{Arguments: rawArgs}}
	headers, err := buildTargetHeaders(req)
	if tt.wantErr {
		if err == nil {
			t.Fatalf("expected error for %v", tt.args)
		}
		return
	}
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tt.wantNil {
		if headers != nil {
			t.Fatalf("expected nil headers, got %v", headers)
		}
		return
	}
	for k, v := range tt.want {
		assertHeader(t, headers, k, v)
	}
}

func TestBuildTargetHeaders(t *testing.T) {
	tests := []buildTargetHeadersTestCase{
		{"no_targeting", map[string]any{}, true, false, nil},
		{"heading_target", map[string]any{
			"target_type": "heading", "target": "My Section",
		}, false, false, map[string]string{"Target-Type": "heading", "Target": "My%20Section"}},
		{"with_scope", map[string]any{
			"target_type": "heading", "target": "Heading", "target_scope": "content",
		}, false, false, map[string]string{"Target-Type": "heading", "Target": "Heading", "Target-Scope": "content"}},
		{"with_delimiter", map[string]any{
			"target_type": "heading", "target": "H1::H2", "target_delimiter": "::",
		}, false, false, map[string]string{"Target-Type": "heading", "Target": "H1::H2", "Target-Delimiter": "::"}},
		{"with_create_if_missing", map[string]any{
			"target_type": "heading", "target": "H", "create_target_if_missing": "true",
		}, false, false, map[string]string{"Target-Type": "heading", "Target": "H", "Create-Target-If-Missing": "true"}},
		{"with_reject", map[string]any{
			"target_type": "heading", "target": "H", "reject_if_content_preexists": "true",
		}, false, false, map[string]string{"Target-Type": "heading", "Target": "H", "Reject-If-Content-Preexists": "true"}},
		{"with_trim", map[string]any{
			"target_type": "heading", "target": "H", "trim_target_whitespace": "true",
		}, false, false, map[string]string{"Target-Type": "heading", "Target": "H", "Trim-Target-Whitespace": "true"}},
		{"missing_target", map[string]any{"target_type": "heading"}, false, true, nil},
		{"missing_target_type", map[string]any{"target": "H"}, false, true, nil},
		{"invalid_target_type", map[string]any{
			"target_type": "invalid", "target": "H",
		}, false, true, nil},
		{"invalid_scope", map[string]any{
			"target_type": "heading", "target": "H", "target_scope": "bad",
		}, false, true, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testBuildTargetHeadersCase(t, tt)
		})
	}
}

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{"no_change", "notes/mynote.md", "notes/mynote.md"},
		{"leading_slash", "/notes/mynote.md", "notes/mynote.md"},
		{"trailing_slash", "notes/", "notes"},
		{"both_slashes", "/notes/mynote.md/", "notes/mynote.md"},
		{"empty", "", ""},
		{"root_path", "/", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizePath(tt.path); got != tt.want {
				t.Fatalf("normalizePath(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func newTestClient(t *testing.T, handler http.HandlerFunc) *obsidian.Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return obsidian.NewClient(&config.Config{
		ObsidianURL: srv.URL,
		ObsidianKey: "test-key",
	})
}

func mustMarshalArgs(args map[string]any) json.RawMessage {
	data, err := json.Marshal(args)
	if err != nil {
		panic(err)
	}
	return data
}

func getText(result *mcp.CallToolResult) string {
	return result.Content[0].(*mcp.TextContent).Text
}

func TestHandleMoveNote(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		var gotMethod, gotPath, gotDest string
		client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			gotMethod = r.Method
			gotPath = r.URL.Path
			gotDest = r.Header.Get("Destination")
			w.WriteHeader(http.StatusNoContent)
		})

		handler := handleMoveNote(client)
		req := &mcp.CallToolRequest{Params: &mcp.CallToolParamsRaw{Arguments: mustMarshalArgs(map[string]any{
			"path":    "notes/old.md",
			"newPath": "notes/new.md",
		})}}
		result, err := handler(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.IsError {
			t.Fatalf("expected success, got error: %v", getText(result))
		}
		if gotMethod != "PATCH" {
			t.Fatalf("method = %q, want PATCH", gotMethod)
		}
		if gotPath != "/vault/notes/old.md" {
			t.Fatalf("path = %q, want /vault/notes/old.md", gotPath)
		}
		if gotDest != "notes/new.md" {
			t.Fatalf("Destination = %q, want notes/new.md", gotDest)
		}
		if !strings.Contains(getText(result), "Successfully moved") {
			t.Fatalf("unexpected result text: %s", getText(result))
		}
	})

	t.Run("success with allowOverwrite", func(t *testing.T) {
		var gotOverwrite string
		client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			gotOverwrite = r.Header.Get("Allow-Overwrite")
			w.WriteHeader(http.StatusNoContent)
		})

		handler := handleMoveNote(client)
		req := &mcp.CallToolRequest{Params: &mcp.CallToolParamsRaw{Arguments: mustMarshalArgs(map[string]any{
			"path":           "notes/old.md",
			"newPath":        "notes/exists.md",
			"allowOverwrite": "true",
		})}}
		result, err := handler(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.IsError {
			t.Fatalf("expected success, got error: %v", getText(result))
		}
		if gotOverwrite != "true" {
			t.Fatalf("Allow-Overwrite = %q, want true", gotOverwrite)
		}
	})

	t.Run("missing path", func(t *testing.T) {
		handler := handleMoveNote(nil)
		req := &mcp.CallToolRequest{Params: &mcp.CallToolParamsRaw{Arguments: mustMarshalArgs(map[string]any{
			"newPath": "notes/new.md",
		})}}
		result, err := handler(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.IsError {
			t.Fatal("expected error for missing path")
		}
	})

	t.Run("missing newPath", func(t *testing.T) {
		handler := handleMoveNote(nil)
		req := &mcp.CallToolRequest{Params: &mcp.CallToolParamsRaw{Arguments: mustMarshalArgs(map[string]any{
			"path": "notes/old.md",
		})}}
		result, err := handler(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.IsError {
			t.Fatal("expected error for missing newPath")
		}
	})

	t.Run("api error", func(t *testing.T) {
		client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`not found`))
		})

		handler := handleMoveNote(client)
		req := &mcp.CallToolRequest{Params: &mcp.CallToolParamsRaw{Arguments: mustMarshalArgs(map[string]any{
			"path":    "notes/missing.md",
			"newPath": "notes/target.md",
		})}}
		result, err := handler(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.IsError {
			t.Fatal("expected error wrapping api error")
		}
	})
}

func TestHandleListNotes(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		var gotMethod, gotPath string
		client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			gotMethod = r.Method
			gotPath = r.URL.Path
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`["note.md"]`))
		})

		handler := handleListNotes(client)
		req := &mcp.CallToolRequest{Params: &mcp.CallToolParamsRaw{Arguments: mustMarshalArgs(map[string]any{})}}
		result, err := handler(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.IsError {
			t.Fatalf("expected success, got error: %v", getText(result))
		}
		if gotMethod != "GET" {
			t.Fatalf("method = %q, want GET", gotMethod)
		}
		if gotPath != "/vault/" {
			t.Fatalf("path = %q, want /vault/", gotPath)
		}
	})

	t.Run("with subdir", func(t *testing.T) {
		var gotPath string
		client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`["note.md"]`))
		})

		handler := handleListNotes(client)
		req := &mcp.CallToolRequest{Params: &mcp.CallToolParamsRaw{Arguments: mustMarshalArgs(map[string]any{
			"dirPath": "notes/sub",
		})}}
		_, _ = handler(context.Background(), req)
		if gotPath != "/vault/notes/sub" {
			t.Fatalf("path = %q, want /vault/notes/sub", gotPath)
		}
	})

	t.Run("api error", func(t *testing.T) {
		client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		})

		handler := handleListNotes(client)
		req := &mcp.CallToolRequest{Params: &mcp.CallToolParamsRaw{Arguments: mustMarshalArgs(map[string]any{})}}
		result, err := handler(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.IsError {
			t.Fatal("expected error wrapping api error")
		}
	})
}

func TestHandleReadNote(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		var gotMethod, gotPath string
		client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			gotMethod = r.Method
			gotPath = r.URL.Path
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("# Hello"))
		})

		handler := handleReadNote(client)
		req := &mcp.CallToolRequest{Params: &mcp.CallToolParamsRaw{Arguments: mustMarshalArgs(map[string]any{
			"path": "notes/test.md",
		})}}
		result, err := handler(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.IsError {
			t.Fatalf("expected success, got error: %v", getText(result))
		}
		if gotMethod != "GET" {
			t.Fatalf("method = %q, want GET", gotMethod)
		}
		if gotPath != "/vault/notes/test.md" {
			t.Fatalf("path = %q, want /vault/notes/test.md", gotPath)
		}
	})

	t.Run("missing path", func(t *testing.T) {
		handler := handleReadNote(nil)
		req := &mcp.CallToolRequest{Params: &mcp.CallToolParamsRaw{Arguments: mustMarshalArgs(map[string]any{})}}
		result, err := handler(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.IsError {
			t.Fatal("expected error for missing path")
		}
	})

	t.Run("api error", func(t *testing.T) {
		client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		})

		handler := handleReadNote(client)
		req := &mcp.CallToolRequest{Params: &mcp.CallToolParamsRaw{Arguments: mustMarshalArgs(map[string]any{
			"path": "notes/missing.md",
		})}}
		result, err := handler(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.IsError {
			t.Fatal("expected error wrapping api error")
		}
	})
}

func TestHandleDeleteNote(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		var gotMethod, gotPath string
		client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			gotMethod = r.Method
			gotPath = r.URL.Path
			w.WriteHeader(http.StatusNoContent)
		})

		handler := handleDeleteNote(client)
		req := &mcp.CallToolRequest{Params: &mcp.CallToolParamsRaw{Arguments: mustMarshalArgs(map[string]any{
			"path": "notes/test.md",
		})}}
		result, err := handler(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.IsError {
			t.Fatalf("expected success, got error: %v", getText(result))
		}
		if gotMethod != "DELETE" {
			t.Fatalf("method = %q, want DELETE", gotMethod)
		}
		if gotPath != "/vault/notes/test.md" {
			t.Fatalf("path = %q, want /vault/notes/test.md", gotPath)
		}
	})

	t.Run("missing path", func(t *testing.T) {
		handler := handleDeleteNote(nil)
		req := &mcp.CallToolRequest{Params: &mcp.CallToolParamsRaw{Arguments: mustMarshalArgs(map[string]any{})}}
		result, err := handler(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.IsError {
			t.Fatal("expected error for missing path")
		}
	})

	t.Run("api error", func(t *testing.T) {
		client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		})

		handler := handleDeleteNote(client)
		req := &mcp.CallToolRequest{Params: &mcp.CallToolParamsRaw{Arguments: mustMarshalArgs(map[string]any{
			"path": "notes/missing.md",
		})}}
		result, err := handler(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.IsError {
			t.Fatal("expected error wrapping api error")
		}
	})
}

func TestHandleGlobalSearch(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		var gotMethod, gotPath string
		client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			gotMethod = r.Method
			gotPath = r.URL.RequestURI()
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[]`))
		})

		handler := handleGlobalSearch(client)
		req := &mcp.CallToolRequest{Params: &mcp.CallToolParamsRaw{Arguments: mustMarshalArgs(map[string]any{
			"query": "test",
		})}}
		result, err := handler(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.IsError {
			t.Fatalf("expected success, got error: %v", getText(result))
		}
		if gotMethod != "POST" {
			t.Fatalf("method = %q, want POST", gotMethod)
		}
		if gotPath != "/search/simple/?query=test" {
			t.Fatalf("path = %q, want /search/simple/?query=test", gotPath)
		}
	})

	t.Run("missing query", func(t *testing.T) {
		handler := handleGlobalSearch(nil)
		req := &mcp.CallToolRequest{Params: &mcp.CallToolParamsRaw{Arguments: mustMarshalArgs(map[string]any{})}}
		result, err := handler(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.IsError {
			t.Fatal("expected error for missing query")
		}
	})
}

func TestHandleSearchReplace(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		var getCalled, putCalled bool
		var putBody string
		client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				getCalled = true
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("hello world"))
			default:
				putCalled = true
				buf := make([]byte, 1024)
				n, _ := r.Body.Read(buf)
				putBody = string(buf[:n])
				w.WriteHeader(http.StatusNoContent)
			}
		})

		handler := handleSearchReplace(client)
		req := &mcp.CallToolRequest{Params: &mcp.CallToolParamsRaw{Arguments: mustMarshalArgs(map[string]any{
			"path":    "notes/test.md",
			"search":  "hello",
			"replace": "hi",
			"count":   float64(-1),
		})}}
		result, err := handler(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.IsError {
			t.Fatalf("expected success, got error: %v", getText(result))
		}
		if !getCalled || !putCalled {
			t.Fatal("expected both GET and PUT calls")
		}
		if putBody != "hi world" {
			t.Fatalf("putBody = %q, want %q", putBody, "hi world")
		}
	})

	t.Run("search text not found", func(t *testing.T) {
		client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("hello world"))
		})

		handler := handleSearchReplace(client)
		req := &mcp.CallToolRequest{Params: &mcp.CallToolParamsRaw{Arguments: mustMarshalArgs(map[string]any{
			"path":    "notes/test.md",
			"search":  "xyz",
			"replace": "hi",
		})}}
		result, err := handler(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.IsError {
			t.Fatal("expected error when search text not found")
		}
	})

	t.Run("missing params", func(t *testing.T) {
		handler := handleSearchReplace(nil)
		req := &mcp.CallToolRequest{Params: &mcp.CallToolParamsRaw{Arguments: mustMarshalArgs(map[string]any{})}}
		result, err := handler(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.IsError {
			t.Fatal("expected error for missing params")
		}
	})
}

func TestHandleManageTags(t *testing.T) {
	t.Run("add tag success", func(t *testing.T) {
		var patchBody string
		client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"tags":["existing"]}`))
			default:
				buf := make([]byte, 1024)
				n, _ := r.Body.Read(buf)
				patchBody = string(buf[:n])
				w.WriteHeader(http.StatusNoContent)
			}
		})

		handler := handleManageTags(client)
		req := &mcp.CallToolRequest{Params: &mcp.CallToolParamsRaw{Arguments: mustMarshalArgs(map[string]any{
			"path":      "notes/test.md",
			"operation": "add",
			"tag":       "newtag",
		})}}
		result, err := handler(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.IsError {
			t.Fatalf("expected success, got error: %v", getText(result))
		}
		if patchBody != `["existing","newtag"]` {
			t.Fatalf("patchBody = %q, want [\"existing\",\"newtag\"]", patchBody)
		}
	})

	t.Run("remove tag success", func(t *testing.T) {
		var patchBody string
		client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"tags":["a","b"]}`))
			default:
				buf := make([]byte, 1024)
				n, _ := r.Body.Read(buf)
				patchBody = string(buf[:n])
				w.WriteHeader(http.StatusNoContent)
			}
		})

		handler := handleManageTags(client)
		req := &mcp.CallToolRequest{Params: &mcp.CallToolParamsRaw{Arguments: mustMarshalArgs(map[string]any{
			"path":      "notes/test.md",
			"operation": "remove",
			"tag":       "a",
		})}}
		result, err := handler(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.IsError {
			t.Fatalf("expected success, got error: %v", getText(result))
		}
		if patchBody != `["b"]` {
			t.Fatalf("patchBody = %q, want [\"b\"]", patchBody)
		}
	})

	t.Run("invalid operation", func(t *testing.T) {
		handler := handleManageTags(nil)
		req := &mcp.CallToolRequest{Params: &mcp.CallToolParamsRaw{Arguments: mustMarshalArgs(map[string]any{
			"path":      "notes/test.md",
			"operation": "invalid",
			"tag":       "x",
		})}}
		result, err := handler(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.IsError {
			t.Fatal("expected error for invalid operation")
		}
	})

	t.Run("missing params", func(t *testing.T) {
		handler := handleManageTags(nil)
		req := &mcp.CallToolRequest{Params: &mcp.CallToolParamsRaw{Arguments: mustMarshalArgs(map[string]any{})}}
		result, err := handler(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.IsError {
			t.Fatal("expected error for missing params")
		}
	})
}

func TestHandleManageFrontmatter(t *testing.T) {
	t.Run("get success", func(t *testing.T) {
		client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"title":"test"}`))
		})

		handler := handleManageFrontmatter(client)
		req := &mcp.CallToolRequest{Params: &mcp.CallToolParamsRaw{Arguments: mustMarshalArgs(map[string]any{
			"path":      "notes/test.md",
			"operation": "get",
		})}}
		result, err := handler(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.IsError {
			t.Fatalf("expected success, got error: %v", getText(result))
		}
	})

	t.Run("set success", func(t *testing.T) {
		var patchMethod, patchTarget string
		client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			patchMethod = r.Method
			patchTarget = r.Header.Get("Target")
			w.WriteHeader(http.StatusNoContent)
		})

		handler := handleManageFrontmatter(client)
		req := &mcp.CallToolRequest{Params: &mcp.CallToolParamsRaw{Arguments: mustMarshalArgs(map[string]any{
			"path":        "notes/test.md",
			"operation":   "set",
			"jsonPayload": `{"title":"hello"}`,
		})}}
		result, err := handler(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.IsError {
			t.Fatalf("expected success, got error: %v", getText(result))
		}
		if patchMethod != "PATCH" {
			t.Fatalf("method = %q, want PATCH", patchMethod)
		}
		if patchTarget != "title" {
			t.Fatalf("Target = %q, want title", patchTarget)
		}
	})

	t.Run("set with invalid json", func(t *testing.T) {
		handler := handleManageFrontmatter(nil)
		req := &mcp.CallToolRequest{Params: &mcp.CallToolParamsRaw{Arguments: mustMarshalArgs(map[string]any{
			"path":        "notes/test.md",
			"operation":   "set",
			"jsonPayload": `not json`,
		})}}
		result, err := handler(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.IsError {
			t.Fatal("expected error for invalid json")
		}
	})

	t.Run("invalid operation", func(t *testing.T) {
		handler := handleManageFrontmatter(nil)
		req := &mcp.CallToolRequest{Params: &mcp.CallToolParamsRaw{Arguments: mustMarshalArgs(map[string]any{
			"path":      "notes/test.md",
			"operation": "delete",
		})}}
		result, err := handler(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.IsError {
			t.Fatal("expected error for invalid operation")
		}
	})

	t.Run("missing params", func(t *testing.T) {
		handler := handleManageFrontmatter(nil)
		req := &mcp.CallToolRequest{Params: &mcp.CallToolParamsRaw{Arguments: mustMarshalArgs(map[string]any{})}}
		result, err := handler(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.IsError {
			t.Fatal("expected error for missing params")
		}
	})
}

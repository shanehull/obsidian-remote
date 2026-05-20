package handlers

import (
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
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

func TestBuildTargetHeaders(t *testing.T) {
	tests := []struct {
		name    string
		args    map[string]any
		wantNil bool
		wantErr bool
		want    map[string]string
	}{
		{"no_targeting", map[string]any{}, true, false, nil},
		{"heading_target", map[string]any{
			"target_type": "heading", "target": "My Section",
		}, false, false, map[string]string{"Target-Type": "heading", "Target": "My Section"}},
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
			req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: tt.args}}
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

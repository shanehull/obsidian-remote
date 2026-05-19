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

func TestBuildTargetHeaders(t *testing.T) {
	tests := []struct {
		name     string
		args     map[string]any
		wantNil  bool
		wantErr  bool
		wantType string
		wantTarg string
		wantScop string
		wantDel  string
		wantCrea string
		wantReje string
		wantTrim string
	}{
		{"no_targeting", map[string]any{}, true, false, "", "", "", "", "", "", ""},
		{"heading_target", map[string]any{
			"target_type": "heading", "target": "My Section",
		}, false, false, "heading", "My Section", "", "", "", "", ""},
		{"with_scope", map[string]any{
			"target_type": "heading", "target": "Heading", "target_scope": "content",
		}, false, false, "heading", "Heading", "content", "", "", "", ""},
		{"with_delimiter", map[string]any{
			"target_type": "heading", "target": "H1::H2", "target_delimiter": "::",
		}, false, false, "heading", "H1::H2", "", "::", "", "", ""},
		{"with_create_if_missing", map[string]any{
			"target_type": "heading", "target": "H", "create_target_if_missing": "true",
		}, false, false, "heading", "H", "", "", "true", "", ""},
		{"with_reject", map[string]any{
			"target_type": "heading", "target": "H", "reject_if_content_preexists": "true",
		}, false, false, "heading", "H", "", "", "", "true", ""},
		{"with_trim", map[string]any{
			"target_type": "heading", "target": "H", "trim_target_whitespace": "true",
		}, false, false, "heading", "H", "", "", "", "", "true"},
		{"missing_target", map[string]any{"target_type": "heading"}, false, true, "", "", "", "", "", "", ""},
		{"missing_target_type", map[string]any{"target": "H"}, false, true, "", "", "", "", "", "", ""},
		{"invalid_target_type", map[string]any{
			"target_type": "invalid", "target": "H",
		}, false, true, "", "", "", "", "", "", ""},
		{"invalid_scope", map[string]any{
			"target_type": "heading", "target": "H", "target_scope": "bad",
		}, false, true, "", "", "", "", "", "", ""},
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
			if headers["Target-Type"] != tt.wantType {
				t.Fatalf("Target-Type = %q, want %q", headers["Target-Type"], tt.wantType)
			}
			if headers["Target"] != tt.wantTarg {
				t.Fatalf("Target = %q, want %q", headers["Target"], tt.wantTarg)
			}
			if tt.wantScop != "" && headers["Target-Scope"] != tt.wantScop {
				t.Fatalf("Target-Scope = %q, want %q", headers["Target-Scope"], tt.wantScop)
			}
			if tt.wantDel != "" && headers["Target-Delimiter"] != tt.wantDel {
				t.Fatalf("Target-Delimiter = %q, want %q", headers["Target-Delimiter"], tt.wantDel)
			}
			if tt.wantCrea != "" && headers["Create-Target-If-Missing"] != tt.wantCrea {
				t.Fatalf("Create-Target-If-Missing = %q, want %q", headers["Create-Target-If-Missing"], tt.wantCrea)
			}
			if tt.wantReje != "" && headers["Reject-If-Content-Preexists"] != tt.wantReje {
				t.Fatalf("Reject-If-Content-Preexists = %q, want %q", headers["Reject-If-Content-Preexists"], tt.wantReje)
			}
			if tt.wantTrim != "" && headers["Trim-Target-Whitespace"] != tt.wantTrim {
				t.Fatalf("Trim-Target-Whitespace = %q, want %q", headers["Trim-Target-Whitespace"], tt.wantTrim)
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

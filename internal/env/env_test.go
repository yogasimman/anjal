// Copyright (c) 2026 Yogasimman Ravisagar
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package env

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolve(t *testing.T) {
	vars := map[string]string{
		"BASE_URL": "https://api.example.com",
		"TOKEN":    "abc123",
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"single placeholder", "{{.BASE_URL}}/users", "https://api.example.com/users"},
		{"whitespace inside", "{{ .BASE_URL }}/users", "https://api.example.com/users"},
		{"trailing space", "{{.BASE_URL }}/users", "https://api.example.com/users"},
		{"leading space", "{{ .BASE_URL}}/users", "https://api.example.com/users"},
		{"multiple placeholders", "{{.BASE_URL}}/users?token={{.TOKEN}}", "https://api.example.com/users?token=abc123"},
		{"no placeholders", "https://example.com", "https://example.com"},
		{"empty string", "", ""},
		{"unmatched placeholder", "{{.MISSING}}/path", "{{.MISSING}}/path"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Resolve(tt.input, vars)
			if got != tt.expected {
				t.Errorf("Resolve(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestResolveNilVars(t *testing.T) {
	input := "{{.KEY}}/path"
	got := Resolve(input, nil)
	if got != input {
		t.Errorf("Resolve with nil vars should return input unchanged, got %q", got)
	}
}

func TestLoadGlobalEnv(t *testing.T) {
	workspace := t.TempDir()
	t.Setenv("HOME", workspace)

	home := filepath.Join(workspace, ".anjal")
	os.MkdirAll(home, 0755)
	os.WriteFile(filepath.Join(home, ".env"), []byte("GLOBAL_KEY=global_value\n"), 0644)

	vars, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if vars["GLOBAL_KEY"] != "global_value" {
		t.Errorf("GLOBAL_KEY = %q, want %q", vars["GLOBAL_KEY"], "global_value")
	}
}

func TestLoadForCollectionOverride(t *testing.T) {
	workspace := t.TempDir()
	t.Setenv("HOME", workspace)

	home := filepath.Join(workspace, ".anjal")
	os.MkdirAll(home, 0755)

	// Global env
	os.WriteFile(filepath.Join(home, ".env"), []byte("KEY=global\nGLOBAL_ONLY=hello\n"), 0644)

	// Collection-specific env (collection.md -> .collection.env)
	os.WriteFile(filepath.Join(home, ".myapi.env"), []byte("KEY=override\n"), 0644)

	vars, err := LoadForCollection("myapi")
	if err != nil {
		t.Fatalf("LoadForCollection() error: %v", err)
	}

	// Collection overrides global
	if vars["KEY"] != "override" {
		t.Errorf("KEY = %q, want %q (collection should override global)", vars["KEY"], "override")
	}
	// Global-only variable still present
	if vars["GLOBAL_ONLY"] != "hello" {
		t.Errorf("GLOBAL_ONLY = %q, want %q", vars["GLOBAL_ONLY"], "hello")
	}
}

func TestLoadForCollectionWithMdExtension(t *testing.T) {
	workspace := t.TempDir()
	t.Setenv("HOME", workspace)

	home := filepath.Join(workspace, ".anjal")
	os.MkdirAll(home, 0755)

	os.WriteFile(filepath.Join(home, ".env"), []byte("KEY=global\n"), 0644)
	os.WriteFile(filepath.Join(home, ".myapi.env"), []byte("KEY=collection\n"), 0644)

	// collection name with .md suffix should be stripped
	vars, err := LoadForCollection("myapi.md")
	if err != nil {
		t.Fatalf("LoadForCollection() error: %v", err)
	}
	if vars["KEY"] != "collection" {
		t.Errorf("KEY = %q, want %q", vars["KEY"], "collection")
	}
}

func TestSaveNewKey(t *testing.T) {
	workspace := t.TempDir()
	t.Setenv("HOME", workspace)

	home := filepath.Join(workspace, ".anjal")
	os.MkdirAll(home, 0755)

	err := Save("myapi", "NEW_KEY", "new_value")
	if err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(home, ".myapi.env"))
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}
	content := string(data)
	if content != "NEW_KEY=new_value\n" {
		t.Errorf("file content = %q, want %q", content, "NEW_KEY=new_value\n")
	}
}

func TestSaveExistingKey(t *testing.T) {
	workspace := t.TempDir()
	t.Setenv("HOME", workspace)

	home := filepath.Join(workspace, ".anjal")
	os.MkdirAll(home, 0755)

	os.WriteFile(filepath.Join(home, ".myapi.env"), []byte("KEY=old_value\nOTHER=keep\n"), 0644)

	err := Save("myapi", "KEY", "updated_value")
	if err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	vars, err := LoadForCollection("myapi")
	if err != nil {
		t.Fatalf("LoadForCollection error: %v", err)
	}
	if vars["KEY"] != "updated_value" {
		t.Errorf("KEY = %q, want %q", vars["KEY"], "updated_value")
	}
	if vars["OTHER"] != "keep" {
		t.Errorf("OTHER = %q, want %q", vars["OTHER"], "keep")
	}
}

func TestDeleteKey(t *testing.T) {
	workspace := t.TempDir()
	t.Setenv("HOME", workspace)

	home := filepath.Join(workspace, ".anjal")
	os.MkdirAll(home, 0755)

	os.WriteFile(filepath.Join(home, ".myapi.env"), []byte("KEY=value\nOTHER=keep\nKEEP=stay\n"), 0644)

	err := Delete("myapi", "KEY")
	if err != nil {
		t.Fatalf("Delete() error: %v", err)
	}

	vars, err := LoadForCollection("myapi")
	if err != nil {
		t.Fatalf("LoadForCollection error: %v", err)
	}
	if _, exists := vars["KEY"]; exists {
		t.Errorf("KEY should have been deleted but still exists")
	}
	if vars["OTHER"] != "keep" {
		t.Errorf("OTHER = %q, want %q", vars["OTHER"], "keep")
	}
	if vars["KEEP"] != "stay" {
		t.Errorf("KEEP = %q, want %q", vars["KEEP"], "stay")
	}
}

func TestDeleteKeyNonexistentFile(t *testing.T) {
	workspace := t.TempDir()
	t.Setenv("HOME", workspace)

	home := filepath.Join(workspace, ".anjal")
	os.MkdirAll(home, 0755)

	err := Delete("nonexistent", "KEY")
	if err != nil {
		t.Fatalf("Delete() on non-existent file should not error: %v", err)
	}
}

func TestSaveEmptyCollectionName(t *testing.T) {
	err := Save("", "KEY", "value")
	if err == nil {
		t.Fatal("Save with empty collection name should return an error")
	}
}

func TestDeleteEmptyCollectionName(t *testing.T) {
	err := Delete("", "KEY")
	if err == nil {
		t.Fatal("Delete with empty collection name should return an error")
	}
}

func TestQuotedValues(t *testing.T) {
	workspace := t.TempDir()
	t.Setenv("HOME", workspace)

	home := filepath.Join(workspace, ".anjal")
	os.MkdirAll(home, 0755)

	os.WriteFile(filepath.Join(home, ".env"), []byte("KEY=\"quoted value\"\nOTHER='single quoted'\nPLAIN=unquoted\n"), 0644)

	vars, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if vars["KEY"] != "quoted value" {
		t.Errorf("KEY = %q, want %q", vars["KEY"], "quoted value")
	}
	if vars["OTHER"] != "single quoted" {
		t.Errorf("OTHER = %q, want %q", vars["OTHER"], "single quoted")
	}
	if vars["PLAIN"] != "unquoted" {
		t.Errorf("PLAIN = %q, want %q", vars["PLAIN"], "unquoted")
	}
}

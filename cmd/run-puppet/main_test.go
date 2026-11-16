package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBranchFromFile(t *testing.T) {
	dir := t.TempDir()
	write := func(name, content string) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		return p
	}

	tests := []struct {
		name string
		path string
		want string
	}{
		{"missing file", filepath.Join(dir, "missing"), "main"},
		{"branch with newline", write("branch", "feature-x\n"), "feature-x"},
		{"empty file", write("empty", " \n"), "main"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := branchFromFile(tt.path, "main"); got != tt.want {
				t.Errorf("branchFromFile() = %q, want %q", got, tt.want)
			}
		})
	}
}

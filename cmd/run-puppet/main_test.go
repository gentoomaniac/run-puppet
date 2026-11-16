package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alecthomas/kong"
)

func TestFileResolver(t *testing.T) {
	dir := t.TempDir()
	write := func(name, content string) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		return p
	}

	tests := []struct {
		name    string
		path    string
		args    []string
		want    string
		wantErr bool
	}{
		{"missing file uses default", filepath.Join(dir, "missing"), nil, "main", false},
		{"file value with newline", write("branch", "feature-x\n"), nil, "feature-x", false},
		{"empty file uses default", write("empty", " \n"), nil, "main", false},
		{"flag overrides file", write("branch2", "feature-x\n"), []string{"--branch=cli"}, "cli", false},
		{"unreadable file fails", dir, nil, "", true},
		{"flag skips unreadable file", dir, []string{"--branch=cli"}, "cli", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var c struct {
				Branch string `default:"main"`
				Other  string `default:"other"`
			}
			parser, err := kong.New(&c, kong.Resolvers(fileResolver("branch", tt.path)))
			if err != nil {
				t.Fatal(err)
			}
			_, err = parser.Parse(tt.args)
			if tt.wantErr {
				if err == nil || !strings.Contains(err.Error(), tt.path) {
					t.Fatalf("Parse() error = %v, want error naming %s", err, tt.path)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if c.Branch != tt.want {
				t.Errorf("Branch = %q, want %q", c.Branch, tt.want)
			}
			if c.Other != "other" {
				t.Errorf("Other = %q, resolver touched an unrelated flag", c.Other)
			}
		})
	}
}

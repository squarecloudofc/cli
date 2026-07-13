package squareconfig

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadFromReaderAndSavePreservesUnknownKeys(t *testing.T) {
	content := "MAIN=index.js\nMEMORY=512\nID=abc123\n# comment line\n"

	config := New(filepath.Join(t.TempDir(), "squarecloud.app"))
	if err := config.LoadFromReader(strings.NewReader(content)); err != nil {
		t.Fatal(err)
	}

	if config.ID != "abc123" {
		t.Fatalf("ID = %q, want abc123", config.ID)
	}

	config.ID = "newid456"
	if err := config.Save(); err != nil {
		t.Fatal(err)
	}

	saved, err := os.ReadFile(config.filename)
	if err != nil {
		t.Fatal(err)
	}

	got := string(saved)
	for _, want := range []string{"MAIN=index.js", "MEMORY=512", "ID=newid456", "# comment line"} {
		if !strings.Contains(got, want) {
			t.Errorf("saved file missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "abc123") {
		t.Errorf("old ID survived the save:\n%s", got)
	}
}

func TestSaveAppendsIDWhenFileWasEmpty(t *testing.T) {
	config := New(filepath.Join(t.TempDir(), "squarecloud.app"))
	config.ID = "abc123"

	if err := config.Save(); err != nil {
		t.Fatal(err)
	}

	saved, _ := os.ReadFile(config.filename)
	if strings.TrimSpace(string(saved)) != "ID=abc123" {
		t.Errorf("unexpected content: %q", saved)
	}
}

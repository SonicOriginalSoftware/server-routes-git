package repo_test

import (
	"testing"

	"git.sonicoriginal.software/routes/git/repo"
	"github.com/go-git/go-billy/v5/memfs"
)

const configPath = "config"

func TestCreateRepo(t *testing.T) {
	memoryFS := memfs.New()
	err := repo.Create(memoryFS, "/")
	if err != nil {
		t.Fatalf("Could not initialize repository: %v", err)
	}

	if _, err := memoryFS.Stat(configPath); err != nil {
		t.Fatalf("Repository initialized incorrectly: %v", err)
	}
}

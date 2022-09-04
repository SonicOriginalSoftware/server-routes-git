package repo_test

import (
	"testing"

	"git.sonicoriginal.software/routes/git/cgi/internal/repo"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
)

const configPath = "config"

func TestCreateRepo(t *testing.T) {
	var (
		err      error
		memoryFS billy.Filesystem
	)

	if memoryFS, _, err = repo.Create(memfs.New()); err != nil {
		t.Fatalf("Could not initialize repository: %v", err)
	}

	if _, err := memoryFS.Stat(configPath); err != nil {
		t.Fatalf("Repository initialized incorrectly: %v", err)
	}
}

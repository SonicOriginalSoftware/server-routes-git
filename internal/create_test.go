package internal_test

import (
	"testing"

	"git.nathanblair.rocks/routes/git/internal"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
)

const (
	configPath = "config"
)

func TestCreateRepo(t *testing.T) {
	var (
		err      error
		memoryFS billy.Filesystem
	)

	if memoryFS, _, err = internal.CreateRepo(memfs.New()); err != nil {
		t.Fatalf("Could not initialize repository: %v", err)
	}

	if _, err := memoryFS.Stat(configPath); err != nil {
		t.Fatalf("Repository initialized incorrectly: %v", err)
	}
}

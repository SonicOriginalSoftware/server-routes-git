package git_test

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	go_git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"

	"git.nathanblair.rocks/routes/git"
	"git.nathanblair.rocks/routes/git/internal/repo"
	lib "git.nathanblair.rocks/server"
)

const (
	port       = "4430"
	configPath = "config"
	localHost  = "localhost"
	remoteName = "go-git-test"
)

var certs []tls.Certificate

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

func TestPush(t *testing.T) {
	var (
		err        error
		memoryFS   billy.Filesystem
		repository *go_git.Repository
	)

	if memoryFS, repository, err = repo.Create(memfs.New()); err != nil {
		t.Fatalf("Could not initialize repository: %v", err)
	}

	route := fmt.Sprintf("localhost/%v/", git.Name)
	t.Setenv(fmt.Sprintf("%v_SERVE_ADDRESS", strings.ToUpper(git.Name)), route)
	t.Setenv("PORT", port)

	remoteURL := fmt.Sprintf("http://%v:%v/%v", localHost, port, git.Name)
	if _, err = repository.CreateRemote(&config.RemoteConfig{
		Name: remoteName,
		URLs: []string{remoteURL},
	}); err != nil {
		t.Fatalf("Could not create remote: %v", err)
	}

	options := &go_git.PushOptions{
		RemoteName: remoteName,
		Force:      true,
		RefSpecs:   []config.RefSpec{},
	}

	_ = git.New(memoryFS)

	ctx, cancelFunction := context.WithCancel(context.Background())

	exitCode, _ := lib.Run(ctx, certs)
	defer close(exitCode)

	err = repository.Push(options)

	cancelFunction()

	if returnCode := <-exitCode; returnCode != 0 {
		t.Fatalf("Server errored: %v", returnCode)
	}

	if err != nil && !errors.Is(err, go_git.NoErrAlreadyUpToDate) {
		t.Fatalf("Could not sync repository with remote: %v", err)
	}
}

package backend_test

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"testing"

	"git.sonicoriginal.software/routes/git/backend"
	"git.sonicoriginal.software/routes/git/internal/repo"
	lib "git.sonicoriginal.software/server"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	go_git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
)

const (
	port       = "4430"
	localHost  = "localhost"
	remoteName = "go-git-test"
)

var certs []tls.Certificate

func TestPush(t *testing.T) {
	var (
		err        error
		memoryFS   billy.Filesystem
		repository *go_git.Repository
	)

	if memoryFS, repository, err = repo.Create(memfs.New()); err != nil {
		t.Fatalf("Could not initialize repository: %v", err)
	}

	t.Setenv("PORT", port)

	remoteURL := fmt.Sprintf("http://%v:%v/", localHost, port)
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

	backend.Register(memoryFS)

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

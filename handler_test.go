package git_test

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"testing"

	"git.sonicoriginal.software/routes/git"
	"git.sonicoriginal.software/routes/git/repo"
	lib "git.sonicoriginal.software/server"

	"github.com/go-git/go-billy/v5/memfs"
	go_git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/storage/memory"
)

const (
	port       = "4430"
	localHost  = "localhost"
	remoteName = "go-git-test"
)

var certs []tls.Certificate

func TestPush(t *testing.T) {
	memoryFS := memfs.New()

	err := repo.Create(memoryFS, "/")
	if err != nil {
		t.Fatalf("Could not initialize repository: %v", err)
	}

	repository, err := go_git.Init(memory.NewStorage(), nil)

	t.Setenv("PORT", port)

	remoteURL := fmt.Sprintf("http://%v:%v/", localHost, port)
	_, err = repository.CreateRemote(&config.RemoteConfig{
		Name: remoteName,
		URLs: []string{remoteURL},
	})
	if err != nil {
		t.Fatalf("Could not create remote: %v", err)
	}

	server := git.NewServer(memoryFS)
	git.New(server)

	ctx, cancelFunction := context.WithCancel(context.Background())
	exitCode, _ := lib.Run(ctx, certs)
	defer close(exitCode)

	err = repository.Push(&go_git.PushOptions{RemoteName: remoteName})

	cancelFunction()

	if returnCode := <-exitCode; returnCode != 0 {
		t.Fatalf("Server errored: %v", returnCode)
	}

	if err != nil && !errors.Is(err, go_git.NoErrAlreadyUpToDate) {
		t.Fatalf("Could not sync repository with remote: %v", err)
	}
}

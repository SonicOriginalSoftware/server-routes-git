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
	"github.com/go-git/go-git/v5/storage/filesystem/dotgit"
	"github.com/go-git/go-git/v5/storage/memory"

	"git.nathanblair.rocks/routes/git"
	lib "git.nathanblair.rocks/server"
)

var certs []tls.Certificate

func createRepo() (repoFS billy.Filesystem, repo *go_git.Repository, err error) {
	dotGit := dotgit.New(memfs.New())
	if err = dotGit.Initialize(); err != nil {
		return
	}
	if _, err = dotGit.ConfigWriter(); err != nil {
		return
	}
	if err = dotGit.Close(); err != nil {
		return
	}
	repoFS = dotGit.Fs()
	repo, err = go_git.Init(memory.NewStorage(), nil)
	return
}

func TestHandler(t *testing.T) {
	const (
		port       = "4430"
		localHost  = "localhost"
		remoteName = "go-git-test"
		repoPath   = "repo"
	)
	var (
		err      error
		memoryFS billy.Filesystem
		repo     *go_git.Repository
	)

	if memoryFS, repo, err = createRepo(); err != nil {
		t.Fatalf("Could not initialize git in memory: %v", err)
	}

	route := fmt.Sprintf("localhost/%v/", git.Name)
	t.Setenv(fmt.Sprintf("%v_SERVE_ADDRESS", strings.ToUpper(git.Name)), route)
	t.Setenv("PORT", port)

	remoteURL := fmt.Sprintf("http://%v:%v/%v", localHost, port, git.Name)
	if _, err = repo.CreateRemote(&config.RemoteConfig{
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

	err = repo.Push(options)

	cancelFunction()

	returnCode := <-exitCode
	if returnCode != 0 {
		t.Fatalf("Server errored: %v", returnCode)
	}

	if err != nil && !errors.Is(err, go_git.NoErrAlreadyUpToDate) {
		t.Fatalf("Could not sync repository with remote: %v", err)
	}
}

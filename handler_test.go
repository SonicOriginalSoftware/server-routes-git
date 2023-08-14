package git_test

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"git.sonicoriginal.software/routes/git"

	"git.sonicoriginal.software/server.git/v2"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	go_git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	git_server "github.com/go-git/go-git/v5/plumbing/transport/server"
	"github.com/go-git/go-git/v5/storage/memory"
)

const (
	portEnvKey = "TEST_PORT"
	remoteName = "go-git-test"
	port       = "4430"
)

var (
	remoteAddress       string
	certs               []tls.Certificate
	testMux             = http.NewServeMux()
	ctx, cancelFunction = context.WithCancel(context.Background())
)

func setup(t *testing.T, filesystem billy.Filesystem, remoteStorage *memory.Storage) (
	remoteURL string,
	localRepo *go_git.Repository,
	remoteRepo *go_git.Remote,
) {
	localRepo, err := go_git.Init(memory.NewStorage(), filesystem)
	if err != nil {
		t.Fatalf("Could not initialize local repository: %v", err)
	}

	path := fmt.Sprintf("%v", t.Name())
	remoteURL = fmt.Sprintf("http://%v/%v/", remoteAddress, path)

	route := git.New(path, git_server.MapLoader{remoteURL: remoteStorage}, testMux)
	t.Logf("Handler registered for route [%v]\n", route)

	remoteRepo, err = localRepo.CreateRemote(
		&config.RemoteConfig{
			Name: remoteName,
			URLs: []string{strings.TrimSuffix(remoteURL, "/")},
		},
	)
	if err != nil {
		t.Fatalf("Could not create remote: %v", err)
	}
	t.Logf("Created remote with URL [%v]\n", remoteRepo.Config().URLs[0])

	if numberOfObjects := len(remoteStorage.Objects); numberOfObjects != 0 {
		t.Fatalf("Remote object count incorrect: %v", numberOfObjects)
	}

	return
}

func checkServer(t *testing.T, serverErrorChannel chan server.Error) {
	serverError := <-serverErrorChannel
	contextError := serverError.Context.Error()
	t.Logf("Server [%v] stopped: %v", remoteAddress, contextError)

	if serverError.Close != nil {
		t.Fatalf("Error closing server: %v", serverError.Close.Error())
	}

	if contextError != server.ErrContextCancelled.Error() {
		t.Fatalf("Server failed unexpectedly: %v", contextError)
	}
}

func noContent(t *testing.T) {
	remoteURL, localRepo, _ := setup(t, memfs.New(), memory.NewStorage())

	t.Logf("Pushing to remote [%v]\n", remoteURL)
	err := localRepo.Push(&go_git.PushOptions{RemoteName: remoteName})

	if err != nil && !errors.Is(err, go_git.NoErrAlreadyUpToDate) {
		t.Fatalf("Could not sync repository with remote: %v", err)
	}
}

func withContent(t *testing.T) {
	localFilesystem := memfs.New()

	file, err := localFilesystem.Create("dummy")
	if err != nil {
		t.Fatalf("Could not create local repo file: %v", err)
	}
	fileName := file.Name()
	if _, err = file.Write([]byte("Test content")); err != nil {
		t.Fatalf("Could not write local repo file: %v", err)
	}
	if err = file.Close(); err != nil {
		t.Fatalf("Could not close local repo file: %v", err)
	}

	remoteStorage := memory.NewStorage()

	remoteURL, localRepo, _ := setup(t, localFilesystem, remoteStorage)

	wt, err := localRepo.Worktree()
	if err != nil {
		t.Fatalf("Could not obtain local worktree: %v", err)
	}
	if _, err = wt.Add(fileName); err != nil {
		t.Fatalf("Could not add [%v]: %v", fileName, err)
	}

	commit, err := wt.Commit("test commit", &go_git.CommitOptions{
		Author: &object.Signature{Name: "Test", Email: "test@example.com"},
	})
	if err != nil {
		t.Fatalf("Could not commit worktree: %v", err)
	}

	t.Logf("Pushing [%v] to remote [%v]\n", commit, remoteURL)
	if err = localRepo.Push(&go_git.PushOptions{RemoteName: remoteName}); err != nil {
		t.Fatalf("%v", err)
	}

	_, found := remoteStorage.ObjectStorage.Commits[commit]
	if !found {
		t.Fatalf("Could not get commit [%v] from remote", commit)
	}
}

func TestPush(t *testing.T) {
	t.Setenv(portEnvKey, port)

	var serverErrorChannel chan server.Error
	remoteAddress, serverErrorChannel = server.Run(ctx, &certs, testMux, portEnvKey)
	t.Logf("Serving on [%v]\n", remoteAddress)

	t.Run("No Content", noContent)
	t.Run("With Content", withContent)

	cancelFunction()

	checkServer(t, serverErrorChannel)
}

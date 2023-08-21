//revive:disable:package-comments

package git_test

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	git "git.sonicoriginal.software/server-routes-git.git"
	"git.sonicoriginal.software/server.git/v2"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	go_git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	git_server "github.com/go-git/go-git/v5/plumbing/transport/server"
	"github.com/go-git/go-git/v5/storage/memory"
)

const (
	defaultBranch = plumbing.Main
	remoteName    = "go-git-test"
	portEnvKey    = "TEST_PORT"
	port          = "4430"
)

var (
	remoteAddress       string
	certs               []tls.Certificate
	testMux             = http.NewServeMux()
	ctx, cancelFunction = context.WithCancel(context.Background())
)

func setupLocalRepo(t *testing.T, worktree billy.Filesystem) (localRepo *go_git.Repository) {
	localRepo, err := go_git.InitWithOptions(
		memory.NewStorage(),
		worktree,
		go_git.InitOptions{
			DefaultBranch: defaultBranch,
		},
	)
	if err != nil {
		t.Fatalf("Could not initialize local repository: %v", err)
	}

	return
}

func setupRemoteRepo(t *testing.T, sourceRepo *go_git.Repository, remoteStorage *memory.Storage) (remoteRepo *go_git.Remote) {
	path := fmt.Sprintf("%v", t.Name())
	remoteURL := fmt.Sprintf("http://%v/%v/", remoteAddress, path)

	route := git.New(path, git_server.MapLoader{remoteURL: remoteStorage}, testMux)
	t.Logf("Handler registered for route [%v]\n", route)

	remoteRepo, err := sourceRepo.CreateRemote(
		&config.RemoteConfig{
			Name: remoteName,
			URLs: []string{strings.TrimSuffix(remoteURL, "/")},
			// Fetch: []config.RefSpec{},
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

func populateFilesystem(t *testing.T) (fileSystem billy.Filesystem) {
	fileSystem = memfs.New()

	file, err := fileSystem.Create("dummy")
	if err != nil {
		t.Fatalf("Could not create local repo file: %v", err)
	}
	if _, err = file.Write([]byte("Test content")); err != nil {
		t.Fatalf("Could not write local repo file: %v", err)
	}
	if err = file.Close(); err != nil {
		t.Fatalf("Could not close local repo file: %v", err)
	}

	return
}

func addCommitInRepo(t *testing.T, repo *go_git.Repository) (commit plumbing.Hash) {
	wt, err := repo.Worktree()
	if err != nil {
		t.Fatalf("Could not obtain local worktree: %v", err)
	}

	if err = wt.AddWithOptions(&go_git.AddOptions{All: true}); err != nil {
		t.Fatalf("Could not add files: %v", err)
	}

	commit, err = wt.Commit("test commit", &go_git.CommitOptions{
		Author: &object.Signature{Name: "Test", Email: "test@example.com"},
	})
	if err != nil {
		t.Fatalf("Could not commit worktree: %v", err)
	}

	return
}

func noContent(t *testing.T) {
	localRepo := setupLocalRepo(t, memfs.New())
	remoteRepo := setupRemoteRepo(t, localRepo, memory.NewStorage())

	t.Logf("Pushing to remote [%v]\n", remoteRepo.Config().URLs[0])
	err := localRepo.Push(&go_git.PushOptions{RemoteName: remoteName})

	if err != nil && !errors.Is(err, go_git.NoErrAlreadyUpToDate) {
		t.Fatalf("Could not sync repository with remote: %v", err)
	}
}

func withContent(t *testing.T) {
	remoteStorage := memory.NewStorage()
	localFilesystem := populateFilesystem(t)

	localRepo := setupLocalRepo(t, localFilesystem)
	remoteRepo := setupRemoteRepo(t, localRepo, remoteStorage)
	remoteURL := remoteRepo.Config().URLs[0]

	commit := addCommitInRepo(t, localRepo)

	t.Logf("Pushing [%v] to remote [%v]\n", commit, remoteURL)
	if err := localRepo.Push(&go_git.PushOptions{RemoteName: remoteName}); err != nil {
		t.Fatalf("%v", err)
	}

	if _, found := remoteStorage.Commits[commit]; !found {
		t.Fatalf("Could not get commit [%v] from remote", commit)
	}
}

func clone(t *testing.T) {
	remoteStorage := memory.NewStorage()
	localFilesystem := populateFilesystem(t)

	sourceRepo := setupLocalRepo(t, localFilesystem)
	remoteRepo := setupRemoteRepo(t, sourceRepo, remoteStorage)
	remoteURL := remoteRepo.Config().URLs[0]

	commit := addCommitInRepo(t, sourceRepo)

	t.Logf("Pushing [%v] to remote [%v]\n", commit, remoteURL)
	if err := sourceRepo.Push(&go_git.PushOptions{RemoteName: remoteName}); err != nil {
		t.Fatalf("%v", err)
	}

	if _, found := remoteStorage.Commits[commit]; !found {
		t.Fatalf("Could not get commit [%v] from remote", commit)
	}

	clonedStorage := memory.NewStorage()
	clonedWorktree := memfs.New()

	_, err := go_git.Clone(
		clonedStorage,
		clonedWorktree,
		&go_git.CloneOptions{
			URL:           remoteURL,
			RemoteName:    remoteName,
			ReferenceName: defaultBranch,
		},
	)
	if err != nil {
		t.Fatalf("Could not clone remote [%v]: %v", remoteURL, err)
	}

	if _, found := clonedStorage.Commits[commit]; !found {
		t.Fatalf("Could not get commit [%v] from remote", commit)
	}
}

func TestHandler(t *testing.T) {
	var serverErrorChannel chan server.Error
	// Handle server startup
	{
		t.Setenv(portEnvKey, port)

		remoteAddress, serverErrorChannel = server.Run(ctx, &certs, testMux, portEnvKey)
		t.Logf("Serving on [%v]\n", remoteAddress)
	}

	t.Run("Push with no content", noContent)
	t.Run("Push with content", withContent)
	t.Run("Clone", clone)

	cancelFunction()

	// Handle server shutdown
	{
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
}

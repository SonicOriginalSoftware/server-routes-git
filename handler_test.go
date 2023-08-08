package git_test

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"git.sonicoriginal.software/routes/git"
	"git.sonicoriginal.software/routes/git/repo"

	"git.sonicoriginal.software/server.git/v2"

	"github.com/go-git/go-billy/v5/memfs"
	go_git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/storage/memory"
)

const (
	portEnvKey = "TEST_PORT"
	remoteName = "go-git-test"
	port       = "4430"
)

var (
	certs []tls.Certificate
	mux   *http.ServeMux = nil
)

func TestPush(t *testing.T) {
	t.Setenv(portEnvKey, port)
	memoryFS := memfs.New()
	gitServer := git.NewServer(memoryFS)
	route := git.New(gitServer, mux)

	t.Logf("Handler registered for route [%v]\n", route)

	ctx, cancelFunction := context.WithCancel(context.Background())
	address, serverErrorChannel := server.Run(ctx, &certs, mux, portEnvKey)

	t.Logf("Serving on [%v]\n", address)

	err := repo.Create(memoryFS, "/")
	if err != nil {
		t.Fatalf("Could not create repository: %v", err)
	}

	repository, err := go_git.Init(memory.NewStorage(), nil)
	if err != nil {
		t.Fatalf("Could not initialize repository: %v", err)
	}

	remoteURL := fmt.Sprintf("http://%v%v", address, route)

	t.Logf("Creating remote with URL [%v]\n", remoteURL)

	if _, err = repository.CreateRemote(&config.RemoteConfig{Name: remoteName, URLs: []string{remoteURL}}); err != nil {
		t.Fatalf("Could not create remote: %v", err)
	}

	t.Logf("Pushing to remote [%v]\n", remoteURL)

	err = repository.Push(&go_git.PushOptions{RemoteName: remoteName})

	cancelFunction()

	serverError := <-serverErrorChannel
	if serverError.Close != nil {
		t.Fatalf("Error closing server: %v", serverError.Close.Error())
	}
	contextError := serverError.Context.Error()

	t.Logf("%v\n", contextError)

	if contextError != server.ErrContextCancelled.Error() {
		t.Fatalf("Server failed unexpectedly: %v", contextError)
	}

	if err != nil && !errors.Is(err, go_git.NoErrAlreadyUpToDate) {
		t.Fatalf("Could not sync repository with remote: %v", err)
	}
}

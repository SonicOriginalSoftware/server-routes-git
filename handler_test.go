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

	"github.com/go-git/go-billy/v5/memfs"
	go_git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	git_server "github.com/go-git/go-git/v5/plumbing/transport/server"
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

	localWorktree := memfs.New()
	localRepo, err := go_git.Init(memory.NewStorage(), localWorktree)
	if err != nil {
		t.Fatalf("Could not initialize local repository: %v", err)
	}

	ctx, cancelFunction := context.WithCancel(context.Background())
	address, serverErrorChannel := server.Run(ctx, &certs, mux, portEnvKey)
	t.Logf("Serving on [%v]\n", address)

	const gitRoute = "/git/"
	remoteURL := fmt.Sprintf("http://%v%v", address, gitRoute)

	serverLoader := git_server.MapLoader{remoteURL: memory.NewStorage()}

	// FIXME This needs an actual remote filesystem
	// More importantly, it needs the `config` file in the .git directory (i.e. a bare repo)
	route := git.New(serverLoader, mux)
	if route != gitRoute {
		cancelFunction()
		t.Fatalf("%v != %v\n", route, remoteURL)
	}

	t.Logf("Handler registered for route [%v]\n", route)

	remoteURL = strings.TrimSuffix(remoteURL, "/")
	t.Logf("Creating remote with URL [%v]\n", remoteURL)
	_, err = localRepo.CreateRemote(&config.RemoteConfig{Name: remoteName, URLs: []string{remoteURL}})
	if err != nil {
		cancelFunction()
		t.Fatalf("Could not create remote: %v", err)
	}

	t.Logf("Pushing to remote [%v]\n", remoteURL)
	err = localRepo.Push(&go_git.PushOptions{RemoteName: remoteName})

	cancelFunction()

	serverError := <-serverErrorChannel
	if serverError.Close != nil {
		t.Fatalf("Error closing server: %v", serverError.Close.Error())
	}
	contextError := serverError.Context.Error()

	t.Logf("%v\n", contextError)

	if contextError != server.ErrContextCancelled.Error() {
		t.Fatalf("Server failed unexpectedly: %v", contextError)
	} else if err != nil && !errors.Is(err, go_git.NoErrAlreadyUpToDate) {
		t.Fatalf("Could not sync repository with remote: %v", err)
	}
}

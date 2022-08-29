package git_test

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"
	"testing"

	"git.nathanblair.rocks/routes/git"
	lib "git.nathanblair.rocks/server"
)

var certs []tls.Certificate

func TestHandler(t *testing.T) {
	route := fmt.Sprintf("localhost/%v/", git.Name)
	t.Setenv(fmt.Sprintf("%v_SERVE_ADDRESS", strings.ToUpper(git.Name)), route)

	_ = git.New()

	ctx, cancelFunction := context.WithCancel(context.Background())

	exitCode, _ := lib.Run(ctx, certs)
	defer close(exitCode)

	// TODO send a git command to the server

	cancelFunction()

	if returnCode := <-exitCode; returnCode != 0 {
		t.Fatalf("Server errored: %v", returnCode)
	}
}

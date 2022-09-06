//revive:disable:package-comments

package git

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/server"
)

// NewServer returns an implementation for go-git transport
func NewServer(fsys billy.Filesystem) transport.Transport {
	return server.NewServer(server.NewFilesystemLoader(fsys))
}

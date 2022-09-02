//revive:disable:package-comments

package backend

import (
	"git.sonicoriginal.software/routes/git/internal"
	info "git.sonicoriginal.software/routes/git/internal/handlers/info_refs"
	receive "git.sonicoriginal.software/routes/git/internal/handlers/receive_pack"
	upload "git.sonicoriginal.software/routes/git/internal/handlers/upload_pack"
	"git.sonicoriginal.software/server/logging"

	"github.com/go-git/go-billy/v5"
)

// Register registers the handlers for git requests
func Register(fsys billy.Filesystem) {
	server := internal.NewServer(fsys)
	logger := logging.New(internal.Name)

	_ = info.New(logger, server)
	_ = receive.New(logger, server)
	_ = upload.New(logger, server)

	return
}

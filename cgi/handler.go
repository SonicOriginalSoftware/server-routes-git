//revive:disable:package-comments

package cgi

import (
	"net/http"

	"git.sonicoriginal.software/routes/git/internal"
	info "git.sonicoriginal.software/routes/git/internal/handlers/info_refs"
	receive "git.sonicoriginal.software/routes/git/internal/handlers/receive_pack"
	upload "git.sonicoriginal.software/routes/git/internal/handlers/upload_pack"
	"git.sonicoriginal.software/server/handlers"
	"git.sonicoriginal.software/server/logging"

	"github.com/go-git/go-billy/v5"
)

// Register registers the handlers for git requests
func Register(fsys billy.Filesystem) {
	server := internal.NewServer(fsys)
	logger := logging.New(internal.Name)

	var handler http.Handler

	handler = info.New(server, logger)
	handlers.Register(internal.Name, "", internal.InfoRefsPath, handler, logger)

	handler = receive.New(server, logger)
	handlers.Register(internal.Name, "", internal.ReceivePackPath, handler, logger)

	handler = upload.New(server, logger)
	handlers.Register(internal.Name, "", internal.UploadPackPath, handler, logger)

	return
}

//revive:disable:package-comments

package git

import (
	"fmt"
	"net/http"
	"os"

	"git.sonicoriginal.software/routes/git/internal"
	"git.sonicoriginal.software/routes/git/internal/pack"

	"git.sonicoriginal.software/logger.git"
	"git.sonicoriginal.software/server.git/v2"

	"github.com/go-git/go-git/v5/plumbing/transport"
	git_server "github.com/go-git/go-git/v5/plumbing/transport/server"
)

const (
	name = "git"

	// contentTypeHeaderKey is the key for the Content-Type header
	contentTypeHeaderKey = "Content-Type"
)

// Handler handles git requests
type handler struct {
	logger logger.Log
	server transport.Transport
}

func (h *handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	h.logger.Info("%v %v\n", request.Method, request.RequestURI)

	writer.Header().Set("Cache-Control", "no-cache")

	service, err := internal.RetrieveService(request.URL.Path)
	if err != nil {
		h.logger.Error("%s", err)
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	transportEndpoint, err := internal.RetrieveTransportEndpoint(
		request.Host,
		request.URL.Path,
		service,
		request.TLS != nil,
	)
	if err != nil {
		h.logger.Error("%s", err)
		http.Error(writer, err.Error(), http.StatusNotAcceptable)
		return
	}

	switch service {
	case internal.InfoRefsPath:
		service = request.URL.Query().Get("service")
		contentType := fmt.Sprintf("application/x-%v-advertisement", service)
		writer.Header().Set(contentTypeHeaderKey, contentType)

		err = internal.Advertise(request.Context(), service, transportEndpoint, h.server, writer)
	case internal.ReceivePackPath:
		contentType := fmt.Sprintf("application/x-%v-result", internal.ReceivePackPath)
		writer.Header().Set(contentTypeHeaderKey, contentType)
		err = pack.Receive(request.Context(), h.server, transportEndpoint, request.Body, writer)
		break
	case internal.UploadPackPath:
		contentType := fmt.Sprintf("application/x-%v-result", internal.UploadPackPath)
		writer.Header().Set(contentTypeHeaderKey, contentType)
		err = pack.Upload(request.Context(), h.server, transportEndpoint, request.Body, writer)
	default:
		err = fmt.Errorf("Invalid request: %v", service)
	}

	if err != nil {
		h.logger.Error("%s", err)
		http.Error(writer, err.Error(), http.StatusBadRequest)
	}
}

// New generates a new git Handler
func New(serverLoader git_server.Loader, mux *http.ServeMux) (route string) {
	logger := logger.New(
		name,
		logger.DefaultSeverity,
		os.Stdout,
		os.Stderr,
	)

	// serverLoader := git_server.NewFilesystemLoader(fsys)
	// serverLoader := git_server.MapLoader{}
	gitServer := git_server.NewServer(serverLoader)
	h := &handler{logger, gitServer}
	return server.RegisterHandler(name, h, mux)
}

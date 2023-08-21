//revive:disable:package-comments

package git

import (
	"fmt"
	"net/http"
	"os"

	"git.sonicoriginal.software/server-routes-git.git/internal"
	"git.sonicoriginal.software/server-routes-git.git/internal/pack"

	"git.sonicoriginal.software/logger.git"
	"git.sonicoriginal.software/server.git/v2"

	"github.com/go-git/go-git/v5/plumbing/transport"
	git_server "github.com/go-git/go-git/v5/plumbing/transport/server"
)

// Handler handles git requests
type handler struct {
	logger logger.Log
	server transport.Transport
}

func (h *handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	h.logger.Info("%v %v\n", request.Method, request.RequestURI)

	writer.Header().Set("Cache-Control", "no-cache")

	requestPath, err := internal.ParsePath(request.URL.Path)
	if err != nil {
		h.logger.Error("%s", err)
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	transportEndpoint, err := internal.RetrieveTransportEndpoint(
		request.Host,
		request.URL.Path,
		requestPath,
		request.TLS != nil,
	)
	if err != nil {
		h.logger.Error("%s", err)
		http.Error(writer, err.Error(), http.StatusNotAcceptable)
		return
	}

	contentTypeRequest := requestPath
	contentTypeSuffix := "result"

	var service string
	if requestPath == internal.InfoRefsPath {
		service = request.URL.Query().Get("service")
		contentTypeRequest = service
		contentTypeSuffix = "advertisement"
	}

	contentType := fmt.Sprintf("application/x-%v-%v", contentTypeRequest, contentTypeSuffix)
	writer.Header().Set("Content-Type", contentType)

	switch requestPath {
	case internal.InfoRefsPath:
		err = internal.Advertise(request.Context(), service, transportEndpoint, h.server, writer)
	case internal.ReceivePackPath:
		err = pack.Receive(request.Context(), h.server, transportEndpoint, request.Body, writer)
	case internal.UploadPackPath:
		err = pack.Upload(request.Context(), h.server, transportEndpoint, request.Body, writer)
	default:
		err = fmt.Errorf("Invalid request: %v", requestPath)
	}

	if err != nil {
		h.logger.Error("%s", err)
		http.Error(writer, err.Error(), http.StatusBadRequest)
	}
}

// New generates a new git Handler
func New(path string, serverLoader git_server.Loader, mux *http.ServeMux) (route string) {
	logger := logger.New(
		fmt.Sprintf("%v", path),
		logger.DefaultSeverity,
		os.Stdout,
		os.Stderr,
	)

	return server.RegisterHandler(
		path,
		&handler{
			logger,
			git_server.NewServer(serverLoader),
		},
		mux,
	)
}

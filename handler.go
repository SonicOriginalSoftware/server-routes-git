//revive:disable:package-comments

package git

import (
	"fmt"
	"net/http"

	"git.sonicoriginal.software/routes/git/internal"
	ci "git.sonicoriginal.software/routes/git/internal"
	info "git.sonicoriginal.software/routes/git/internal/info_refs"
	"git.sonicoriginal.software/routes/git/internal/pack"
	"git.sonicoriginal.software/server/handlers"
	"git.sonicoriginal.software/server/logging"

	"github.com/go-git/go-git/v5/plumbing/transport"
)

const rootPath = "/"

// Handler handles git requests
type Handler struct {
	logger logging.Log
	server transport.Transport
}

func (handler *Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	handler.logger.Info("%v %v\n", request.Method, request.RequestURI)

	writer.Header().Set("Cache-Control", "no-cache")

	service, err := ci.RetrieveService(request.URL.Path)
	if err != nil {
		handler.logger.Error("%s", err)
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	transportEndpoint, err := ci.RetrieveTransportEndpoint(
		request.Host,
		request.URL.Path,
		service,
		request.TLS != nil,
	)
	if err != nil {
		handler.logger.Error("%s", err)
		http.Error(writer, err.Error(), http.StatusNotAcceptable)
		return
	}

	switch service {
	case ci.InfoRefsPath:
		service = request.URL.Query().Get("service")
		contentType := fmt.Sprintf("application/x-%v-advertisement", service)
		writer.Header().Set(internal.ContentTypeHeaderKey, contentType)

		err = info.Advertise(request.Context(), service, transportEndpoint, handler.server, writer)
	case ci.ReceivePackPath:
		contentType := fmt.Sprintf("application/x-%v-result", ci.ReceivePackPath)
		writer.Header().Set(internal.ContentTypeHeaderKey, contentType)
		err = pack.Receive(request.Context(), handler.server, transportEndpoint, request.Body, writer)
		break
	case ci.UploadPackPath:
		contentType := fmt.Sprintf("application/x-%v-result", ci.UploadPackPath)
		writer.Header().Set(internal.ContentTypeHeaderKey, contentType)
		err = pack.Upload(request.Context(), handler.server, transportEndpoint, request.Body, writer)
	default:
		err = fmt.Errorf("Invalid request: %v", service)
	}

	if err != nil {
		handler.logger.Error("%s", err)
		http.Error(writer, err.Error(), http.StatusBadRequest)
	}
}

// New generates a new git Handler
func New(server transport.Transport) (handler *Handler) {
	logger := logging.New(internal.Name)
	handler = &Handler{logger, server}
	handlers.Register(internal.Name, "", rootPath, handler, logger)
	return
}

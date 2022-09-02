//revive:disable:package-comments

package receive

import (
	"fmt"
	"net/http"
	"strings"

	"git.sonicoriginal.software/routes/git/internal"
	"git.sonicoriginal.software/routes/git/internal/pack"
	"git.sonicoriginal.software/server/handlers"
	"git.sonicoriginal.software/server/logging"

	"github.com/go-git/go-git/v5/plumbing/transport"
)

// Handler handles Git requests
type Handler struct {
	server transport.Transport
	logger *logging.Logger
	path   string
}

func (handler *Handler) ServeHTTP(writer http.ResponseWriter, r *http.Request) {
	handler.logger.Info("(%v) %v %v\n", r.Host, r.Method, r.RequestURI)

	writer.Header().Set("Cache-Control", "no-cache")
	writer.Header().Set(internal.ContentTypeHeaderKey, internal.PlainContentValue)

	repoPath := strings.TrimSuffix(r.URL.Path, handler.path)
	proto := "http"
	if r.TLS != nil {
		proto = "https"
	}
	endpoint := fmt.Sprintf("%v://%v/%v", proto, r.Host, repoPath)

	transportEndpoint, err := transport.NewEndpoint(endpoint)
	if err != nil {
		handler.logger.Error("%s", err)
		http.Error(writer, err.Error(), http.StatusNotAcceptable)
		return
	}

	session, err := handler.server.NewReceivePackSession(transportEndpoint, nil)
	if err != nil {
		handler.logger.Error("%s", err)
		http.Error(writer, err.Error(), http.StatusNotAcceptable)
		return
	}

	contentType := fmt.Sprintf("application/x-%v-result", internal.ReceivePackPath)
	writer.Header().Set(internal.ContentTypeHeaderKey, contentType)

	err = pack.Receive(r.Context(), session, r.Body, writer)
	if err != nil {
		handler.logger.Error("%s", err)
		writer.Header().Set(internal.ContentTypeHeaderKey, internal.PlainContentValue)
		http.Error(writer, err.Error(), http.StatusNotAcceptable)
		return
	}
}

// New returns a new Handler
func New(logger *logging.Logger, server transport.Transport) (handler *Handler) {
	handlerPath := fmt.Sprintf("%v/", internal.ReceivePackPath)
	handler = &Handler{server, logger, handlerPath}
	handlers.Register(internal.Name, "", handlerPath, handler, logger)

	return
}

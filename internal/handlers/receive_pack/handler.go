//revive:disable:package-comments

package receive

import (
	"fmt"
	"net/http"

	"git.sonicoriginal.software/routes/git/internal"
	"git.sonicoriginal.software/routes/git/internal/pack"
	"git.sonicoriginal.software/server/logging"

	"github.com/go-git/go-git/v5/plumbing/transport"
)

// Handler handles Git requests
type Handler struct {
	server transport.Transport
	logger logging.Log
}

func (handler *Handler) ServeHTTP(writer http.ResponseWriter, r *http.Request) {
	handler.logger.Info("(%v) %v %v\n", r.Host, r.Method, r.RequestURI)

	writer.Header().Set("Cache-Control", "no-cache")

	transportEndpoint, err := internal.RetrieveTransportEndpoint(
		r.Host,
		r.URL.Path,
		internal.ReceivePackPath,
		r.TLS != nil,
	)
	if err != nil {
		handler.logger.Error("%s", err)
		http.Error(writer, err.Error(), http.StatusNotAcceptable)
		return
	}

	contentType := fmt.Sprintf("application/x-%v-result", internal.UploadPackPath)
	writer.Header().Set(internal.ContentTypeHeaderKey, contentType)

	err = pack.Receive(r.Context(), handler.server, transportEndpoint, r.Body, writer)
	if err != nil {
		handler.logger.Error("%s", err)
		http.Error(writer, err.Error(), http.StatusNotAcceptable)
		return
	}
}

// New creates a new valid instance of the Handler
func New(server transport.Transport, logger logging.Log) *Handler {
	return &Handler{server: server, logger: logger}
}

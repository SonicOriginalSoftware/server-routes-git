//revive:disable:package-comments

package info

import (
	"fmt"
	"net/http"

	"git.sonicoriginal.software/routes/git/internal"
	"git.sonicoriginal.software/server/logging"

	"github.com/go-git/go-git/v5/plumbing/format/pktline"
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
		internal.InfoRefsPath,
		r.TLS != nil,
	)
	if err != nil {
		handler.logger.Error("%s", err)
		http.Error(writer, err.Error(), http.StatusNotAcceptable)
		return
	}

	service := r.URL.Query().Get("service")
	contentType := fmt.Sprintf("application/x-%v-advertisement", service)
	writer.Header().Set(internal.ContentTypeHeaderKey, contentType)

	var session transport.Session
	if service == internal.ReceivePackPath {
		session, err = handler.server.NewReceivePackSession(transportEndpoint, nil)
	} else if service == internal.UploadPackPath {
		session, err = handler.server.NewUploadPackSession(transportEndpoint, nil)
	}
	if err != nil {
		handler.logger.Error("%s", err)
		http.Error(writer, err.Error(), http.StatusNotAcceptable)
		return
	}

	refs, err := session.AdvertisedReferencesContext(r.Context())
	if err != nil {
		handler.logger.Error("%s", err)
		http.Error(writer, err.Error(), http.StatusNotAcceptable)
		return
	}

	serviceLine := []byte(fmt.Sprintf("# service=%v", service))
	refs.Prefix = [][]byte{serviceLine, pktline.Flush}
	if err = refs.Encode(writer); err != nil {
		handler.logger.Error("%s", err)
		http.Error(writer, err.Error(), http.StatusBadRequest)
	}
}

// New creates a new valid instance of the Handler
func New(server transport.Transport, logger logging.Log) *Handler {
	return &Handler{server, logger}
}

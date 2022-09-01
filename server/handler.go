//revive:disable:package-comments

package server

import (
	"fmt"
	"net/http"
	"strings"

	"git.nathanblair.rocks/routes/git"
	"git.nathanblair.rocks/routes/git/internal"
	"git.nathanblair.rocks/routes/git/internal/request"
	"git.nathanblair.rocks/server/handlers"
	"git.nathanblair.rocks/server/logging"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
)

const contentTypeHeaderKey = "Content-Type"
const plainContentValue = "text/plain; charset=utf-8"

// Handler handles Git requests
type Handler struct {
	logger *logging.Logger
	server transport.Transport
}

func (handler *Handler) ServeHTTP(writer http.ResponseWriter, r *http.Request) {
	handler.logger.Info("(%v) %v %v\n", r.Host, r.Method, r.RequestURI)
	requestPath := strings.TrimPrefix(r.URL.Path, fmt.Sprintf("/%v/", git.Name))
	writer.Header().Set("Cache-Control", "no-cache")

	isInfoRefsRequest,
		service,
		contentType,
		endpoint,
		err := request.Parse(r.TLS != nil, r.Host, requestPath, r.URL.Query())
	if err != nil {
		handler.logger.Error("%s", err)
		writer.Header().Set(contentTypeHeaderKey, plainContentValue)
		http.Error(writer, err.Error(), http.StatusBadRequest)
	}

	writer.Header().Set(contentTypeHeaderKey, contentType)

	err = request.Do(
		r.Context(),
		isInfoRefsRequest,
		service,
		endpoint,
		handler.server,
		r.Body,
		writer,
	)
	if err != nil {
		handler.logger.Error("%s", err)
		writer.Header().Set(contentTypeHeaderKey, plainContentValue)
		http.Error(writer, err.Error(), http.StatusBadRequest)
	}
}

// New returns a new Handler
func New(fsys billy.Filesystem) (handler *Handler) {
	logger := logging.New(git.Name)
	handler = &Handler{logger, internal.NewServer(fsys)}
	handlers.Register(git.Name, handler, logger)

	return
}

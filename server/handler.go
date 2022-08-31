//revive:disable:package-comments

package server

import (
	"context"
	"io"
	"net/http"
	"net/url"

	"git.nathanblair.rocks/routes/git"
	"git.nathanblair.rocks/routes/git/internal"
	"git.nathanblair.rocks/routes/git/internal/pack"
	"git.nathanblair.rocks/routes/git/internal/request"
	"git.nathanblair.rocks/server/handlers"
	"git.nathanblair.rocks/server/logging"

	billy "github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
)

// Handler handles Git requests
type Handler struct {
	logger *logging.Logger
	server transport.Transport
}

func handleRequest(
	context context.Context,
	isSecure bool,
	host, requestPath string,
	content io.ReadCloser,
	writer http.ResponseWriter,
	server transport.Transport,
	query url.Values,
) (statusCode int, err error) {
	contentType,
		isInfoRefsRequest,
		isReceiveRequest,
		isUploadRequest,
		session,
		err := request.Initialize(
		context,
		isSecure,
		host,
		requestPath,
		server,
		writer,
		query,
	)
	if err != nil {
		statusCode = http.StatusForbidden
		return
	}

	writer.Header().Set("Content-Type", contentType)

	if isReceiveRequest && !isInfoRefsRequest {
		err = pack.Receive(context, session, content, writer)
	} else if isUploadRequest && !isInfoRefsRequest {
		err = pack.Upload(context, session, content, writer)
	}
	if err != nil {
		statusCode = http.StatusBadRequest
		return
	}

	return
}

func (handler *Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	requestPath := request.URL.Path
	handler.logger.Info("(%v) %v %v\n", request.Host, request.Method, requestPath)
	writer.Header().Set("Cache-Control", "no-cache")

	isSecure := request.TLS != nil

	if statusCode, err := handleRequest(
		request.Context(),
		isSecure,
		request.Host,
		requestPath,
		request.Body,
		writer,
		handler.server,
		request.URL.Query(),
	); err != nil {
		handler.logger.Error("%s", err)
		writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
		http.Error(writer, err.Error(), statusCode)
	}
}

// New returns a new Handler
func New(fsys billy.Filesystem) (handler *Handler) {
	logger := logging.New(git.Name)
	handler = &Handler{logger, internal.NewServer(fsys)}
	handlers.Register(git.Name, handler, logger)

	return
}

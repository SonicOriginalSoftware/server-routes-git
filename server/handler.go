//revive:disable:package-comments

package server

import (
	"fmt"
	"net/http"
	"strings"

	"git.nathanblair.rocks/routes/git"
	"git.nathanblair.rocks/routes/git/internal"
	"git.nathanblair.rocks/server/handlers"
	"git.nathanblair.rocks/server/logging"

	billy "github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5/plumbing/format/pktline"
	"github.com/go-git/go-git/v5/plumbing/protocol/packp"
	"github.com/go-git/go-git/v5/plumbing/transport"
	go_git_server "github.com/go-git/go-git/v5/plumbing/transport/server"
)

const (
	receiveService = "git-receive-pack"
	uploadService  = "git-upload-pack"
)

// Handler handles Git requests
type Handler struct {
	logger *logging.Logger
	server transport.Transport
}

func handleRequest(
	writer http.ResponseWriter,
	request *http.Request,
	requestPath string,
	server transport.Transport,
) (statusCode int, err error) {
	service, infoRefsRequest := internal.Parse(requestPath)

	if service != receiveService && service != uploadService && !infoRefsRequest {
		err = fmt.Errorf("Invalid request: %v", requestPath)
		statusCode = http.StatusForbidden
		return
	}

	contentTypeSuffix := "result"
	trimParts := 1
	if isInfoRefsRequest {
		trimParts = 2
		contentTypeSuffix = "advertisement"
		service = request.URL.Query().Get("service")
	}

	proto := "http"
	if request.TLS != nil {
		proto = "https"
	}
	repoPath := strings.Join(pathParts[0:len(pathParts)-trimParts], "/")
	endpoint := fmt.Sprintf("%v://%v/%v", proto, request.Host, repoPath)
	transportEndpoint, err := transport.NewEndpoint(endpoint)
	if err != nil {
		statusCode = http.StatusBadRequest
		return
	}

	writer.Header().Set("Content-Type", fmt.Sprintf("application/x-%v-%v", service, contentTypeSuffix))

	var session transport.Session
	if service == receiveService {
		session, err = server.NewReceivePackSession(transportEndpoint, nil)
		if !isInfoRefsRequest && err == nil {
			err = internal.ReceivePack(request.Context(), session, request.Body, writer)
		}
	} else if service == uploadService {
		session, err = server.NewUploadPackSession(transportEndpoint, nil)
		if !isInfoRefsRequest && err == nil {
			err = internal.UploadPack(request.Context(), session, request.Body, writer)
		}
	}

	if isInfoRefsRequest && err == nil {
		var refs *packp.AdvRefs
		if refs, err = session.AdvertisedReferencesContext(request.Context()); err == nil {
			refs.Prefix = [][]byte{[]byte(fmt.Sprintf("# service=%v", service)), pktline.Flush}
			err = refs.Encode(writer)
		}
	}

	return
}

func (handler *Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	handler.logger.Info("(%v) %v %v\n", request.Host, request.Method, request.URL.Path)
	writer.Header().Set("Cache-Control", "no-cache")
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")

	if statusCode, err := handleRequest(writer, request, handler.server); err != nil {
		handler.logger.Error("%s", err)
		http.Error(writer, err.Error(), statusCode)
	}
}

// New returns a new Handler
func New(fsys billy.Filesystem) (handler *Handler) {
	loader := go_git_server.NewFilesystemLoader(fsys)

	logger := logging.New(git.Name)
	handler = &Handler{logger, go_git_server.NewServer(loader)}
	handlers.Register(git.Name, handler, logger)

	return
}

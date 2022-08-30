//revive:disable:package-comments

package server

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"git.nathanblair.rocks/server/handlers"
	"git.nathanblair.rocks/server/logging"

	billy "github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5/plumbing/format/pktline"
	"github.com/go-git/go-git/v5/plumbing/protocol/packp"
	"github.com/go-git/go-git/v5/plumbing/transport"
	go_git_server "github.com/go-git/go-git/v5/plumbing/transport/server"
)

const (
	// Name is the name used to identify the service
	Name = "git"

	infoRefsService = "refs"
	receiveService  = "git-receive-pack"
	uploadService   = "git-upload-pack"
)

// Handler handles Git requests
type Handler struct {
	logger *logging.Logger
	server transport.Transport
}

func uploadPack(
	context context.Context,
	session transport.Session,
	body io.ReadCloser,
	writer http.ResponseWriter,
) (err error) {
	uploadPackRequest := packp.NewUploadPackRequest()
	if err = uploadPackRequest.Decode(body); err != nil {
		return
	}

	uploadPackSession, ok := session.(transport.UploadPackSession)
	if !ok {
		err = fmt.Errorf("Could not create upload-pack session")
		return
	}

	uploadResponse, err := uploadPackSession.UploadPack(context, uploadPackRequest)
	if err != nil || uploadResponse == nil {
		return
	}

	return uploadResponse.Encode(writer)
}

func receivePack(
	context context.Context,
	session transport.Session,
	body io.ReadCloser,
	writer http.ResponseWriter,
) (err error) {
	receivePackRequest := packp.NewReferenceUpdateRequest()
	receivePackRequest.Decode(body)

	receivePackSession, ok := session.(transport.ReceivePackSession)
	if !ok {
		err = fmt.Errorf("Could not create receive-pack session")
		return
	}

	reportStatus, err := receivePackSession.ReceivePack(context, receivePackRequest)
	if err != nil || reportStatus == nil {
		return
	}

	return reportStatus.Encode(writer)
}

func handleRequest(
	writer http.ResponseWriter,
	request *http.Request,
	server transport.Transport,
) (statusCode int, err error) {
	requestPath := strings.TrimPrefix(request.URL.Path, fmt.Sprintf("/%v/", Name))
	writer.Header().Set("Cache-Control", "no-cache")
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")

	pathParts := strings.Split(requestPath, "/")
	service := pathParts[len(pathParts)-1]
	isInfoRefsRequest := service == infoRefsService

	if service != receiveService && service != uploadService && !isInfoRefsRequest {
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
			err = receivePack(request.Context(), session, request.Body, writer)
		}
	} else if service == uploadService {
		session, err = server.NewUploadPackSession(transportEndpoint, nil)
		if !isInfoRefsRequest && err == nil {
			err = uploadPack(request.Context(), session, request.Body, writer)
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

	if statusCode, err := handleRequest(writer, request, handler.server); err != nil {
		writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
		handler.logger.Error("%s", err)
		http.Error(writer, err.Error(), statusCode)
	}
}

// New returns a new Handler
func New(fsys billy.Filesystem) (handler *Handler) {
	loader := go_git_server.NewFilesystemLoader(fsys)

	logger := logging.New(Name)
	handler = &Handler{logger, go_git_server.NewServer(loader)}
	handlers.Register(Name, handler, logger)

	return
}

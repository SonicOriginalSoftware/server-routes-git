//revive:disable:package-comments

package git

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"server/env"
	"server/logging"
	"server/net/local"

	"github.com/go-git/go-git/v5/plumbing/format/pktline"
	"github.com/go-git/go-git/v5/plumbing/protocol/packp"
	"github.com/go-git/go-git/v5/plumbing/transport"
	go_git "github.com/go-git/go-git/v5/plumbing/transport/server"
)

const prefix = "git"
const infoRefsService = "refs"
const receiveService = "git-receive-pack"
const uploadService = "git-upload-pack"

// Handler handles Git requests
type Handler struct {
	outlog *log.Logger
	errlog *log.Logger
	server transport.Transport
}

func (handler *Handler) handleError(writer http.ResponseWriter, errCode int, err error) {
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")

	handler.errlog.Printf("%s", err)
	http.Error(writer, err.Error(), errCode)
}

func (handler *Handler) uploadPack(
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

func (handler *Handler) receivePack(
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

// ServeHTTP fulfills the http.Handler contract for Handler
func (handler *Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Add("Cache-Control", "no-cache")

	path := request.URL.Path
	err := fmt.Errorf("Invalid request: %v", path)

	pathParts := strings.Split(path, "/")
	service := pathParts[len(pathParts)-1]
	isInfoRefsRequest := service == infoRefsService

	if service != receiveService && service != uploadService && !isInfoRefsRequest {
		handler.handleError(writer, http.StatusForbidden, err)
		return
	}

	contentTypeSuffix := "result"
	trimParts := 1

	if isInfoRefsRequest {
		trimParts = 2
		contentTypeSuffix = "advertisement"
		service = request.URL.Query().Get("service")
	}

	repoPath := strings.Join(pathParts[0:len(pathParts)-trimParts], "/")
	endpoint := fmt.Sprintf("%v://%v%v", "https", request.Host, repoPath)
	transportEndpoint, err := transport.NewEndpoint(endpoint)
	if err != nil {
		handler.handleError(writer, http.StatusBadRequest, err)
		return
	}

	writer.Header().Set("Content-Type", fmt.Sprintf("application/x-%v-%v", service, contentTypeSuffix))

	var session transport.Session
	if service == receiveService {
		session, err = handler.server.NewReceivePackSession(transportEndpoint, nil)
		if !isInfoRefsRequest && err == nil {
			err = handler.receivePack(request.Context(), session, request.Body, writer)
		}
	} else if service == uploadService {
		session, err = handler.server.NewUploadPackSession(transportEndpoint, nil)
		if !isInfoRefsRequest && err == nil {
			err = handler.uploadPack(request.Context(), session, request.Body, writer)
		}
	}

	if isInfoRefsRequest && err == nil {
		var refs *packp.AdvRefs
		if refs, err = session.AdvertisedReferencesContext(request.Context()); err == nil {
			refs.Prefix = [][]byte{[]byte(fmt.Sprintf("# service=%v", service)), pktline.Flush}
			err = refs.Encode(writer)
		}
	}

	if err != nil {
		handler.handleError(writer, http.StatusBadRequest, err)
	}
}

// Prefix is the subdomain prefix
func (handler *Handler) Prefix() string {
	return prefix
}

// Address returns the address the Handler will service
func (handler *Handler) Address() string {
	return env.Address(prefix, fmt.Sprintf("%v.%v", prefix, local.Path("")))
}

// New returns a new Handler
func New() *Handler {
	return &Handler{
		outlog: logging.NewLog(prefix),
		errlog: logging.NewError(prefix),
		server: go_git.DefaultServer,
	}
}

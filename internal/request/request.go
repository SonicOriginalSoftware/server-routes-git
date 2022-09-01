//revive:disable:package-comments

package request

import (
	"context"
	"fmt"
	"io"
	"strings"

	"git.nathanblair.rocks/routes/git/internal/pack"
	"github.com/go-git/go-git/v5/plumbing/format/pktline"
	"github.com/go-git/go-git/v5/plumbing/transport"
)

const (
	infoRefsService = "refs"
	receiveService  = "git-receive-pack"
	uploadService   = "git-upload-pack"
)

// Parse a request and extract desired operation
func Parse(
	isSecure bool,
	requestHost, requestPath string,
	query map[string][]string,
) (
	isInfoRefsRequest bool,
	service, contentType, endpoint string,
	err error,
) {
	pathParts := strings.Split(requestPath, "/")
	service = pathParts[len(pathParts)-1]
	isInfoRefsRequest = service == infoRefsService

	if service != receiveService && service != uploadService && !isInfoRefsRequest {
		err = fmt.Errorf("Invalid request: %v", requestPath)
		return
	}

	contentTypeSuffix := "result"
	trimParts := 1
	if isInfoRefsRequest {
		trimParts = 2
		contentTypeSuffix = "advertisement"
		service = query["service"][0]
	}

	proto := "http"
	if isSecure {
		proto = "https"
	}
	repoPath := strings.Join(pathParts[0:len(pathParts)-trimParts], "/")
	endpoint = fmt.Sprintf("%v://%v/%v", proto, requestHost, repoPath)
	contentType = fmt.Sprintf("application/x-%v-%v", service, contentTypeSuffix)

	return
}

// Do executes the appropriate operation for the service
func Do(
	context context.Context,
	isInfoRefsRequest bool,
	service, endpoint string,
	server transport.Transport,
	content io.ReadCloser,
	writer io.Writer,
) (err error) {
	isReceiveRequest := service == receiveService
	isUploadRequest := service == uploadService

	transportEndpoint, err := transport.NewEndpoint(endpoint)
	if err != nil {
		return
	}

	var session transport.Session
	if isReceiveRequest {
		session, err = server.NewReceivePackSession(transportEndpoint, nil)
	} else if isUploadRequest {
		session, err = server.NewUploadPackSession(transportEndpoint, nil)
	}
	if err != nil {
		return
	}

	if isReceiveRequest && !isInfoRefsRequest {
		err = pack.Receive(context, session, content, writer)
	} else if isUploadRequest && !isInfoRefsRequest {
		err = pack.Upload(context, session, content, writer)
	}
	if err != nil {
		return
	}

	if !isInfoRefsRequest {
		return
	}

	refs, err := session.AdvertisedReferencesContext(context)
	if err != nil {
		return
	}

	refs.Prefix = [][]byte{[]byte(fmt.Sprintf("# service=%v", service)), pktline.Flush}
	err = refs.Encode(writer)

	return
}

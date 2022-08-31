//revive:disable:package-comments

package request

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/format/pktline"
	"github.com/go-git/go-git/v5/plumbing/protocol/packp"
	"github.com/go-git/go-git/v5/plumbing/transport"
)

const (
	infoRefsService = "refs"
	receiveService  = "git-receive-pack"
	uploadService   = "git-upload-pack"
)

// Initialize an operation from a request
func Initialize(
	context context.Context,
	isSecure bool,
	host, requestPath string,
	server transport.Transport,
	writer io.Writer,
	query map[string][]string,
) (
	contentType string,
	isInfoRefsRequest,
	isReceiveRequest,
	isUploadRequest bool,
	session transport.Session,
	err error,
) {
	pathParts := strings.Split(requestPath, "/")
	service := pathParts[len(pathParts)-1]
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

	isReceiveRequest = service == receiveService
	isUploadRequest = service == uploadService
	contentType = fmt.Sprintf("application/x-%v-%v", service, contentTypeSuffix)

	proto := "http"
	if isSecure {
		proto = "https"
	}

	repoPath := strings.Join(pathParts[0:len(pathParts)-trimParts], "/")
	endpoint := fmt.Sprintf("%v://%v/%v", proto, host, repoPath)
	transportEndpoint, err := transport.NewEndpoint(endpoint)
	if err != nil {
		return
	}

	if isReceiveRequest {
		session, err = server.NewReceivePackSession(transportEndpoint, nil)
	} else if isUploadRequest {
		session, err = server.NewUploadPackSession(transportEndpoint, nil)
	}

	if isInfoRefsRequest {
		var refs *packp.AdvRefs
		if refs, err = session.AdvertisedReferencesContext(context); err == nil {
			refs.Prefix = [][]byte{[]byte(fmt.Sprintf("# service=%v", service)), pktline.Flush}
			err = refs.Encode(writer)
		}
	}

	return
}

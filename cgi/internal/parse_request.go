//revive:disable:package-comments

package internal

import (
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/transport"
)

// RetrieveService parses the service from the request path
func RetrieveService(requestPath string) (service string) {
	pathParts := strings.Split(requestPath, "/")
	pathLength := len(pathParts)
	if pathLength >= 2 && strings.Join(pathParts[pathLength-2:pathLength], "/") == InfoRefsPath {
		service = InfoRefsPath
	} else if pathParts[pathLength] == UploadPackPath {
		service = UploadPackPath
	} else if pathParts[pathLength] == ReceivePackPath {
		service = ReceivePackPath
	}
	return
}

// RetrieveTransportEndpoint generates the transport endpoint from incoming request parameters
func RetrieveTransportEndpoint(requestHost, requestPath, trimString string, isSecure bool) (*transport.Endpoint, error) {
	repoPath := strings.TrimSuffix(requestPath, trimString)
	proto := "http"
	if isSecure {
		proto = "https"
	}
	endpoint := fmt.Sprintf("%v://%v/%v", proto, requestHost, repoPath)

	return transport.NewEndpoint(endpoint)
}

//revive:disable:package-comments

package internal

import (
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/transport"
)

// ParsePath parses the service from the request path
func ParsePath(requestPath string) (service string, err error) {
	pathParts := strings.Split(requestPath, "/")
	pathLength := len(pathParts)
	if pathLength < 2 {
		err = fmt.Errorf("Invalid request: %v", requestPath)
		return
	}
	if pathLength >= 3 && strings.Join(pathParts[pathLength-2:pathLength], "/") == InfoRefsPath {
		service = InfoRefsPath
	} else if pathParts[pathLength-1] == UploadPackPath {
		service = UploadPackPath
	} else if pathParts[pathLength-1] == ReceivePackPath {
		service = ReceivePackPath
	} else {
		err = fmt.Errorf("Invalid request: %v", requestPath)
	}
	return
}

// RetrieveTransportEndpoint generates the transport endpoint from incoming request parameters
func RetrieveTransportEndpoint(requestHost, requestPath, trimString string, isSecure bool) (*transport.Endpoint, error) {
	repoPath := strings.TrimSuffix(requestPath, trimString)
	repoPath = strings.TrimPrefix(repoPath, "/")
	proto := "http"
	if isSecure {
		proto = "https"
	}
	endpoint := fmt.Sprintf("%v://%v/%v", proto, requestHost, repoPath)

	return transport.NewEndpoint(endpoint)
}

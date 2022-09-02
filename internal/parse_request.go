//revive:disable:package-comments

package internal

import (
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/transport"
)

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

//revive:disable:package-comments

package internal

import (
	"strings"
)

const (
	infoRefsService = "refs"
)

// Parse from a request path
func Parse(requestPath string) (service string, infoRefsRequest bool) {
	pathParts := strings.Split(requestPath, "/")
	service = pathParts[len(pathParts)-1]
	infoRefsRequest = service == infoRefsService
	return
}

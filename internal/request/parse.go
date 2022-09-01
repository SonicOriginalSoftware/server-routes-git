//revive:disable:package-comments

package request

import (
	"fmt"
	"strings"
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

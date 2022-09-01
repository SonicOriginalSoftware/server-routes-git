//revive:disable:package-comments

package request

import (
	"context"
	"fmt"
	"io"

	"git.nathanblair.rocks/routes/git/internal/pack"
	"github.com/go-git/go-git/v5/plumbing/format/pktline"
	"github.com/go-git/go-git/v5/plumbing/transport"
)

// Do executes the appropriate operation for the service
func Do(
	context context.Context,
	isInfoRefsRequest bool,
	service, endpoint string,
	server transport.Transport,
	content io.ReadCloser,
	writer io.Writer,
) (err error) {
	transportEndpoint, err := transport.NewEndpoint(endpoint)
	if err != nil {
		return
	}

	var session transport.Session
	if service == receiveService {
		session, err = server.NewReceivePackSession(transportEndpoint, nil)
		if err == nil && !isInfoRefsRequest {
			err = pack.Receive(context, session, content, writer)
		}
	} else if service == uploadService {
		session, err = server.NewUploadPackSession(transportEndpoint, nil)
		if err == nil && !isInfoRefsRequest {
			err = pack.Upload(context, session, content, writer)
		}
	}
	if err != nil || !isInfoRefsRequest {
		return
	}

	refs, err := session.AdvertisedReferencesContext(context)
	if err != nil {
		return
	}

	serviceLine := []byte(fmt.Sprintf("# service=%v", service))
	refs.Prefix = [][]byte{serviceLine, pktline.Flush}
	err = refs.Encode(writer)

	return
}

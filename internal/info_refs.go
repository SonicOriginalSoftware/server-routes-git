//revive:disable:package-comments

package internal

import (
	"context"
	"fmt"
	"io"

	"github.com/go-git/go-git/v5/plumbing/format/pktline"
	"github.com/go-git/go-git/v5/plumbing/transport"
)

// Advertise handles an info/refs request
func Advertise(
	context context.Context,
	service string,
	transportEndpoint *transport.Endpoint,
	server transport.Transport,
	writer io.Writer,
) (err error) {
	var session transport.Session
	if service == ReceivePackPath {
		session, err = server.NewReceivePackSession(transportEndpoint, nil)
	} else if service == UploadPackPath {
		session, err = server.NewUploadPackSession(transportEndpoint, nil)
	}
	if err != nil {
		return
	}

	refs, err := session.AdvertisedReferencesContext(context)
	if err != nil {
		return
	}

	serviceLine := []byte(fmt.Sprintf("# service=%v", service))
	refs.Prefix = [][]byte{serviceLine, pktline.Flush}
	return refs.Encode(writer)
}

//revive:disable:package-comments

package pack

import (
	"context"
	"fmt"
	"io"

	"github.com/go-git/go-git/v5/plumbing/protocol/packp"
	"github.com/go-git/go-git/v5/plumbing/transport"
)

// Receive handles receive-pack requests
func Receive(
	context context.Context,
	session transport.Session,
	body io.ReadCloser,
	writer io.Writer,
) (err error) {
	receivePackRequest := packp.NewReferenceUpdateRequest()
	if err = receivePackRequest.Decode(body); err != nil {
		return
	}

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

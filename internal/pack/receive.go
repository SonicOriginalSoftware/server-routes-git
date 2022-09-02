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
	packRequest := packp.NewReferenceUpdateRequest()
	if err = packRequest.Decode(body); err != nil {
		return
	}

	packSession, ok := session.(transport.ReceivePackSession)
	if !ok {
		err = fmt.Errorf("Could not create receive-pack session")
		return
	}

	response, err := packSession.ReceivePack(context, packRequest)
	if err != nil {
		return
	} else if response == nil {
		err = fmt.Errorf("Could not generate receive pack response from pack session")
		return
	}

	return response.Encode(writer)
}

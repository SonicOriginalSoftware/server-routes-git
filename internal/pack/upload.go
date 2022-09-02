//revive:disable:package-comments

package pack

import (
	"context"
	"fmt"
	"io"

	"github.com/go-git/go-git/v5/plumbing/protocol/packp"
	"github.com/go-git/go-git/v5/plumbing/transport"
)

// Upload handles upload-pack requests
func Upload(
	context context.Context,
	session transport.Session,
	content io.ReadCloser,
	writer io.Writer,
) (err error) {
	packRequest := packp.NewUploadPackRequest()
	if err = packRequest.Decode(content); err != nil {
		return
	}

	packSession, ok := session.(transport.UploadPackSession)
	if !ok {
		err = fmt.Errorf("Could not create upload-pack session")
		return
	}

	response, err := packSession.UploadPack(context, packRequest)
	if err != nil {
		return
	} else if response == nil {
		err = fmt.Errorf("Could not generate upload pack response from pack session")
		return
	}

	return response.Encode(writer)
}

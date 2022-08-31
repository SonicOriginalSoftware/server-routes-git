//revive:disable:package-comments

package pack

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/go-git/go-git/v5/plumbing/protocol/packp"
	"github.com/go-git/go-git/v5/plumbing/transport"
)

// Upload handles upload-pack requests
func Upload(
	context context.Context,
	session transport.Session,
	content io.ReadCloser,
	writer http.ResponseWriter,
) (err error) {
	uploadPackRequest := packp.NewUploadPackRequest()
	if err = uploadPackRequest.Decode(content); err != nil {
		return
	}

	uploadPackSession, ok := session.(transport.UploadPackSession)
	if !ok {
		err = fmt.Errorf("Could not create upload-pack session")
		return
	}

	uploadResponse, err := uploadPackSession.UploadPack(context, uploadPackRequest)
	if err != nil || uploadResponse == nil {
		return
	}

	return uploadResponse.Encode(writer)
}

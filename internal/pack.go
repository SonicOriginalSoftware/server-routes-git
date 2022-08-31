//revive:disable:package-comments

package internal

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/go-git/go-git/v5/plumbing/protocol/packp"
	"github.com/go-git/go-git/v5/plumbing/transport"
)

// UploadPack handles upload-pack requests
func UploadPack(
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

// ReceivePack handles receive-pack requests
func ReceivePack(
	context context.Context,
	session transport.Session,
	body io.ReadCloser,
	writer http.ResponseWriter,
) (err error) {
	receivePackRequest := packp.NewReferenceUpdateRequest()
	receivePackRequest.Decode(body)

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

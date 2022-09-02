//revive:disable:package-comments

package api

import (
	"fmt"
	"net/http"

	"git.sonicoriginal.software/routes/git/internal"
	"git.sonicoriginal.software/server/handlers"
	"git.sonicoriginal.software/server/logging"
)

const apiPathName = "api"

// Handler handles Git requests
type Handler struct {
	logger *logging.Logger
}

func (handler *Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
}

// New returns a new Handler
func New() (handler *Handler) {
	logger := logging.New(internal.Name)
	handler = &Handler{logger}
	apiPath := fmt.Sprintf("%v/%v", internal.Name, apiPathName)
	handlers.Register(internal.Name, "", apiPath, handler, logger)

	return
}

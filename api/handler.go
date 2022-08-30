//revive:disable:package-comments

package api

import (
	"net/http"

	"git.nathanblair.rocks/routes/git"
	"git.nathanblair.rocks/server/handlers"
	"git.nathanblair.rocks/server/logging"
)

// Handler handles Git requests
type Handler struct {
	logger *logging.Logger
}

func (handler *Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
}

// New returns a new Handler
func New() (handler *Handler) {
	logger := logging.New(git.Name)
	handler = &Handler{logger}
	handlers.Register(git.Name, handler, logger)

	return
}

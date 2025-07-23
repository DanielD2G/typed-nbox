package handlers

import (
	"nbox/internal/application"
	"net/http"

	"github.com/norlis/httpgate/pkg/adapter/apidriven/presenters"
)

type StaticHandler struct {
	config *application.Config
	render presenters.Presenters
}

func NewStaticHandler(config *application.Config, render presenters.Presenters) *StaticHandler {
	return &StaticHandler{
		config: config,
		render: render,
	}
}

func (s *StaticHandler) Environments(w http.ResponseWriter, r *http.Request) {
	s.render.JSON(w, r, s.config.AllowedPrefixes)
}

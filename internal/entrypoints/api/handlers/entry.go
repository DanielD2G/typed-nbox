package handlers

import (
	"encoding/json"
	"errors"
	"nbox/internal/domain"
	"nbox/internal/domain/models"
	"net/http"
	"strings"

	"github.com/norlis/httpgate/pkg/adapter/apidriven/presenters"
	_ "github.com/norlis/httpgate/pkg/kit/problem"
)

type EntryHandler struct {
	entryAdapter  domain.EntryAdapter
	entryUseCase  domain.EntryUseCase
	secretAdapter domain.SecretAdapter
	render        presenters.Presenters
}

func NewEntryHandler(entryAdapter domain.EntryAdapter, secretAdapter domain.SecretAdapter, entryUseCase domain.EntryUseCase, render presenters.Presenters) *EntryHandler {
	return &EntryHandler{entryAdapter: entryAdapter, secretAdapter: secretAdapter, entryUseCase: entryUseCase, render: render}
}

// Upsert
// @Summary Upsert entries
// @Description insert / update vars
// @Tags entry
// @Accept json
// @Produce json
// @Param data body []models.Entry true "Upsert template"
// @Param authorization header string true "Bearer | Basic"
// @Success 200 {object} map[string]string ""
// @Failure 400 {object} problem.ProblemDetail "Bad Request"
// @Failure 401 {object} problem.ProblemDetail "Unauthorized"
// @Failure 500 {object} problem.ProblemDetail "Internal error"
// @Router /api/entry [post]
func (h *EntryHandler) Upsert(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var entries []models.Entry

	if err := json.NewDecoder(r.Body).Decode(&entries); err != nil {
		h.render.Error(w, r, err, presenters.WithStatus(http.StatusBadRequest))
		return
	}

	h.render.JSON(w, r, h.entryUseCase.Upsert(ctx, entries))
}

// ListByPrefix
// @Summary Filter by prefix
// @Description list all keys by path
// @Tags entry
// @Produce json
// @Param v query string true "key path"
// @Param authorization header string true "Bearer | Basic"
// @Success 200 {object} []models.Entry ""
// @Failure 401 {object} problem.ProblemDetail "Unauthorized"
// @Failure 500 {object} problem.ProblemDetail "Internal error"
// @Router /api/entry/prefix [get]
func (h *EntryHandler) ListByPrefix(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	prefix := r.URL.Query().Get("v")

	entries, err := h.entryAdapter.List(ctx, prefix)
	if err != nil {
		h.render.Error(w, r, err, presenters.WithStatus(http.StatusBadRequest))
		return
	}

	h.render.JSON(w, r, entries)
}

// GetByKey
// @Summary Retrieve key
// @Description detail
// @Tags entry
// @Produce json
// @Param v query string true "key path"
// @Param authorization header string true "Bearer | Basic"
// @Success 200 {object} models.Entry ""
// @Failure 401 {object} problem.ProblemDetail "Unauthorized"
// @Failure 500 {object} problem.ProblemDetail "Internal error"
// @Router /api/entry/key [get]
func (h *EntryHandler) GetByKey(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	key := r.URL.Query().Get("v")
	entry, err := h.entryAdapter.Retrieve(ctx, key)
	if err != nil {
		h.render.Error(w, r, err, presenters.WithStatus(http.StatusBadRequest))
		return
	}

	h.render.JSON(w, r, entry)
}

// DeleteKey
// @Summary Delete
// @Description delete keys & children
// @Tags entry
// @Produce json
// @Param v query string true "key path"
// @Param authorization header string true "Bearer | Basic"
// @Success 200 {object} object{message=string} ""
// @Failure 401 {object} problem.ProblemDetail "Unauthorized"
// @Failure 500 {object} problem.ProblemDetail "Internal error"
// @Router /api/entry/key [delete]
func (h *EntryHandler) DeleteKey(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	key := r.URL.Query().Get("v")
	err := h.entryAdapter.Delete(ctx, key)
	if err != nil {
		h.render.Error(w, r, err, presenters.WithStatus(http.StatusBadRequest))
		return
	}

	h.render.JSON(w, r, map[string]string{"message": "ok"})
}

// Tracking
// @Summary History
// @Description history changes
// @Tags entry
// @Produce json
// @Param v query string true "key path"
// @Param authorization header string true "Bearer | Basic"
// @Success 200 {object} []models.Tracking ""
// @Failure 401 {object} problem.ProblemDetail "Unauthorized"
// @Failure 500 {object} problem.ProblemDetail "Internal error"
// @Router /api/track/key [get]
func (h *EntryHandler) Tracking(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	key := r.URL.Query().Get("v")

	entries, err := h.entryAdapter.Tracking(ctx, key)
	if err != nil {
		h.render.Error(w, r, err, presenters.WithStatus(http.StatusBadRequest))
		return
	}

	h.render.JSON(w, r, entries)
}

// RetrieveSecretValue
// @Summary Retrieve secret value
// @Description plain value
// @Tags entry
// @Produce json
// @Param v query string true "key path"
// @Param authorization header string true "Bearer | Basic"
// @Success 200 {object} models.Entry ""
// @Failure 401 {object} problem.ProblemDetail "Unauthorized"
// @Failure 404 {object} problem.ProblemDetail "Not found"
// @Failure 500 {object} problem.ProblemDetail "Internal error"
// @Router /api/entry/secret-value [get]
func (h *EntryHandler) RetrieveSecretValue(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	key := r.URL.Query().Get("v")

	if key == "" {
		h.render.Error(w, r, errors.New("empty key"), presenters.WithStatus(http.StatusBadRequest))
		return
	}

	if !strings.HasPrefix(key, "/") {
		key = "/" + key
	}

	entry, err := h.secretAdapter.RetrieveSecretValue(ctx, key)
	if err != nil {
		h.render.Error(w, r, err, presenters.WithStatus(http.StatusBadRequest))
		return
	}

	if entry == nil {
		h.render.Error(w, r, errors.New("not found key"), presenters.WithStatus(http.StatusNotFound))
		return
	}

	h.render.JSON(w, r, entry)
}

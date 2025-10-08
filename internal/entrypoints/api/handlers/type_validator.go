package handlers

import (
	"encoding/json"
	"nbox/internal/domain"
	"nbox/internal/domain/models"
	"net/http"

	"github.com/norlis/httpgate/pkg/adapter/apidriven/presenters"
	_ "github.com/norlis/httpgate/pkg/kit/problem"
)

type TypeValidatorHandler struct {
	typeValidatorAdapter domain.TypeValidatorAdapter
	render               presenters.Presenters
}

func NewTypeValidatorHandler(typeValidatorAdapter domain.TypeValidatorAdapter, render presenters.Presenters) *TypeValidatorHandler {
	return &TypeValidatorHandler{typeValidatorAdapter: typeValidatorAdapter, render: render}
}

// Upsert
// @Summary Create or update type validator
// @Description Create or update a custom type validator
// @Tags type_validator
// @Accept json
// @Produce json
// @Param data body models.TypeValidator true "Type Validator"
// @Param authorization header string true "Bearer | Basic"
// @Success 200 {object} map[string]string ""
// @Failure 400 {object} problem.ProblemDetail "Bad Request"
// @Failure 401 {object} problem.ProblemDetail "Unauthorized"
// @Failure 500 {object} problem.ProblemDetail "Internal error"
// @Router /api/type-validator [post]
func (h *TypeValidatorHandler) Upsert(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var validator models.TypeValidator

	if err := json.NewDecoder(r.Body).Decode(&validator); err != nil {
		h.render.Error(w, r, err, presenters.WithStatus(http.StatusBadRequest))
		return
	}

	if err := h.typeValidatorAdapter.Upsert(ctx, validator); err != nil {
		h.render.Error(w, r, err, presenters.WithStatus(http.StatusBadRequest))
		return
	}

	h.render.JSON(w, r, map[string]string{"message": "ok"})
}

// List
// @Summary List all type validators
// @Description List all type validators (built-in + custom)
// @Tags type_validator
// @Produce json
// @Param authorization header string true "Bearer | Basic"
// @Success 200 {object} []models.TypeValidator ""
// @Failure 401 {object} problem.ProblemDetail "Unauthorized"
// @Failure 500 {object} problem.ProblemDetail "Internal error"
// @Router /api/type-validator [get]
func (h *TypeValidatorHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	validators, err := h.typeValidatorAdapter.List(ctx)
	if err != nil {
		h.render.Error(w, r, err, presenters.WithStatus(http.StatusBadRequest))
		return
	}

	h.render.JSON(w, r, validators)
}

// GetByName
// @Summary Retrieve type validator by name
// @Description Get details of a specific type validator
// @Tags type_validator
// @Produce json
// @Param name query string true "validator name"
// @Param authorization header string true "Bearer | Basic"
// @Success 200 {object} models.TypeValidator ""
// @Failure 401 {object} problem.ProblemDetail "Unauthorized"
// @Failure 404 {object} problem.ProblemDetail "Not found"
// @Failure 500 {object} problem.ProblemDetail "Internal error"
// @Router /api/type-validator/name [get]
func (h *TypeValidatorHandler) GetByName(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	name := r.URL.Query().Get("name")

	if name == "" {
		h.render.Error(w, r, http.ErrMissingFile, presenters.WithStatus(http.StatusBadRequest))
		return
	}

	validator, err := h.typeValidatorAdapter.Retrieve(ctx, name)
	if err != nil {
		h.render.Error(w, r, err, presenters.WithStatus(http.StatusBadRequest))
		return
	}

	if validator == nil {
		h.render.Error(w, r, http.ErrMissingFile, presenters.WithStatus(http.StatusNotFound))
		return
	}

	h.render.JSON(w, r, validator)
}

// Delete
// @Summary Delete type validator
// @Description Delete a custom type validator
// @Tags type_validator
// @Produce json
// @Param name query string true "validator name"
// @Param authorization header string true "Bearer | Basic"
// @Success 200 {object} object{message=string} ""
// @Failure 400 {object} problem.ProblemDetail "Bad Request"
// @Failure 401 {object} problem.ProblemDetail "Unauthorized"
// @Failure 500 {object} problem.ProblemDetail "Internal error"
// @Router /api/type-validator/name [delete]
func (h *TypeValidatorHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	name := r.URL.Query().Get("name")

	if name == "" {
		h.render.Error(w, r, http.ErrMissingFile, presenters.WithStatus(http.StatusBadRequest))
		return
	}

	err := h.typeValidatorAdapter.Delete(ctx, name)
	if err != nil {
		h.render.Error(w, r, err, presenters.WithStatus(http.StatusBadRequest))
		return
	}

	h.render.JSON(w, r, map[string]string{"message": "ok"})
}
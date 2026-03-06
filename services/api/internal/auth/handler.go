package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/mail"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/MobinaToorani/retrosnack/pkg/httputil"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(r chi.Router) {
	r.Post("/auth/register", h.register)
	r.Post("/auth/login", h.login)
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.ErrorMsg(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if msg := validateAuth(req.Email, req.Password); msg != "" {
		httputil.ErrorMsg(w, http.StatusBadRequest, msg)
		return
	}

	resp, err := h.svc.Register(r.Context(), req)
	if err != nil {
		if errors.Is(err, ErrEmailTaken) {
			httputil.Error(w, http.StatusConflict, err)
			return
		}
		httputil.Error(w, http.StatusInternalServerError, err)
		return
	}

	httputil.JSON(w, http.StatusCreated, resp)
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.ErrorMsg(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if msg := validateAuth(req.Email, req.Password); msg != "" {
		httputil.ErrorMsg(w, http.StatusBadRequest, msg)
		return
	}

	resp, err := h.svc.Login(r.Context(), req)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			httputil.Error(w, http.StatusUnauthorized, err)
			return
		}
		httputil.Error(w, http.StatusInternalServerError, err)
		return
	}

	httputil.JSON(w, http.StatusOK, resp)
}

func validateAuth(email, password string) string {
	email = strings.TrimSpace(email)
	if email == "" {
		return "email is required"
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return "invalid email format"
	}
	if len(password) < 8 {
		return "password must be at least 8 characters"
	}
	if len(password) > 72 {
		return "password must be at most 72 characters"
	}
	return ""
}

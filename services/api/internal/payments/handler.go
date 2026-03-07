package payments

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/retrosnack-clothing/retrosnack/pkg/httputil"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(r chi.Router) {
	r.Post("/checkout", h.createCheckout)
	r.Post("/webhooks/square", h.squareWebhook)
}

func (h *Handler) createCheckout(w http.ResponseWriter, r *http.Request) {
	var req CreateCheckoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.ErrorMsg(w, http.StatusBadRequest, "invalid request body")
		return
	}

	origin := r.Header.Get("Origin")
	if origin == "" {
		origin = "https://retrosnack.shop"
	}

	redirectURL := origin + "/orders/" + req.OrderID.String() + "/confirmation"

	sess, err := h.svc.CreateCheckout(r.Context(), req, redirectURL)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, err)
		return
	}
	httputil.JSON(w, http.StatusOK, sess)
}

func (h *Handler) squareWebhook(w http.ResponseWriter, r *http.Request) {
	payload, err := io.ReadAll(io.LimitReader(r.Body, 65536))
	if err != nil {
		httputil.ErrorMsg(w, http.StatusBadRequest, "failed to read body")
		return
	}

	signature := r.Header.Get("x-square-hmacsha256-signature")
	if err := h.svc.HandleWebhook(r.Context(), payload, signature); err != nil {
		// log the error but return 200 so square doesn't retry permanent failures
		slog.Error("webhook processing failed", "error", err)
	}

	httputil.NoContent(w)
}

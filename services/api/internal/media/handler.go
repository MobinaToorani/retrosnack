package media

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/MobinaToorani/retrosnack/pkg/httputil"
	"github.com/MobinaToorani/retrosnack/pkg/middleware"
)

const maxUploadSize = 10 << 20 // 10 MB

type Handler struct {
	svc       Service
	jwtSecret string
}

func NewHandler(svc Service, jwtSecret string) *Handler {
	return &Handler{svc: svc, jwtSecret: jwtSecret}
}

func (h *Handler) Register(r chi.Router) {
	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(h.jwtSecret))
		r.Use(middleware.RequireRole("admin", "seller"))
		r.Post("/products/{productId}/images", h.uploadImage)
	})
}

func (h *Handler) uploadImage(w http.ResponseWriter, r *http.Request) {
	productID, err := uuid.Parse(chi.URLParam(r, "productId"))
	if err != nil {
		httputil.ErrorMsg(w, http.StatusBadRequest, "invalid product id")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		httputil.ErrorMsg(w, http.StatusBadRequest, "file too large or invalid form")
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		httputil.ErrorMsg(w, http.StatusBadRequest, "missing image file")
		return
	}
	defer file.Close()

	sizeStr := strconv.FormatInt(header.Size, 10)
	_ = sizeStr

	upload, err := h.svc.Upload(r.Context(), productID, header.Filename, file, header.Size)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, err)
		return
	}
	httputil.JSON(w, http.StatusCreated, upload)
}

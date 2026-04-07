package handlers

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"github.com/dathan/go-project-template/internal/auth"
	"github.com/dathan/go-project-template/internal/db"
	"github.com/dathan/go-project-template/internal/db/models"
	"github.com/dathan/go-project-template/internal/server/middleware"
)

// AdminHandler provides admin-only endpoints.
type AdminHandler struct {
	store  db.Store
	jwtSvc *auth.JWTService
}

func NewAdminHandler(store db.Store, jwtSvc *auth.JWTService) *AdminHandler {
	return &AdminHandler{store: store, jwtSvc: jwtSvc}
}

// ListUsers returns all users with pagination.
// GET /api/v1/admin/users
func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit == 0 {
		limit = 50
	}

	users, total, err := h.store.Users().List(r.Context(), models.ListOptions{
		Limit:  limit,
		Offset: offset,
		Order:  "created_at DESC",
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list users")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"users": users,
		"total": total,
	})
}

// AssumeUser creates an impersonation JWT for the target user.
// POST /api/v1/admin/users/{id}/assume
func (h *AdminHandler) AssumeUser(w http.ResponseWriter, r *http.Request) {
	adminUser := middleware.UserFromContext(r.Context())

	targetID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	target, err := h.store.Users().GetByID(r.Context(), targetID)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	token, err := h.jwtSvc.SignImpersonation(adminUser.ID, target.ID, target.Role)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to sign impersonation token")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"token":       token,
		"target_user": target,
	})
}

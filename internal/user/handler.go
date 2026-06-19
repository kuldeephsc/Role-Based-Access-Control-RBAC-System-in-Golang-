package user

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"rbac-platform/internal/httpx"
)

// Handler implements the self-or-permission rule from the API contract
// directly (GET/PATCH /users/:id: self OR view_profile/create_user) rather
// than through route-level middleware, since that exception can't be
// expressed as a single static per-route permission.
type Handler struct {
	svc       *Service
	authorize func(ctx context.Context, userID uuid.UUID, permission string) (bool, error)
}

func NewHandler(svc *Service, authorize func(ctx context.Context, userID uuid.UUID, permission string) (bool, error)) *Handler {
	return &Handler{svc: svc, authorize: authorize}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("", h.List)
	rg.GET("/:id", h.Get)
	rg.PATCH("/:id", h.Update)
}

func (h *Handler) List(c *gin.Context) {
	callerID, ok := c.MustGet("user_id").(uuid.UUID)
	if !ok {
		httpx.Error(c, http.StatusUnauthorized, "unauthenticated", "authentication required")
		return
	}
	allowed, err := h.authorize(c.Request.Context(), callerID, "view_users")
	if err != nil {
		httpx.Error(c, http.StatusInternalServerError, "internal_error", "could not verify permission")
		return
	}
	if !allowed {
		httpx.Error(c, http.StatusForbidden, "permission_denied", "you do not have permission to list users")
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	users, total, err := h.svc.List(c.Request.Context(), limit, offset)
	if err != nil {
		httpx.Error(c, http.StatusInternalServerError, "internal_error", "could not list users")
		return
	}
	c.JSON(http.StatusOK, gin.H{"users": users, "total": total})
}

func (h *Handler) Get(c *gin.Context) {
	callerID, ok := c.MustGet("user_id").(uuid.UUID)
	if !ok {
		httpx.Error(c, http.StatusUnauthorized, "unauthenticated", "authentication required")
		return
	}
	targetID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.Error(c, http.StatusBadRequest, "invalid_id", "invalid user id")
		return
	}
	if callerID != targetID {
		if !h.hasPermission(c.Request.Context(), callerID, "view_profile") {
			httpx.Error(c, http.StatusForbidden, "permission_denied", "you do not have permission to view this user")
			return
		}
	}
	u, err := h.svc.Get(c.Request.Context(), targetID)
	if err != nil {
		httpx.Error(c, http.StatusNotFound, "not_found", "user not found")
		return
	}
	c.JSON(http.StatusOK, u)
}

type updateRequest struct {
	FullName string `json:"full_name"`
}

func (h *Handler) Update(c *gin.Context) {
	callerID, ok := c.MustGet("user_id").(uuid.UUID)
	if !ok {
		httpx.Error(c, http.StatusUnauthorized, "unauthenticated", "authentication required")
		return
	}
	targetID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.Error(c, http.StatusBadRequest, "invalid_id", "invalid user id")
		return
	}
	if callerID != targetID {
		if !h.hasPermission(c.Request.Context(), callerID, "create_user") {
			httpx.Error(c, http.StatusForbidden, "permission_denied", "you do not have permission to update this user")
			return
		}
	}
	var req updateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Error(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	u, err := h.svc.UpdateProfile(c.Request.Context(), targetID, req.FullName)
	if err != nil {
		httpx.Error(c, http.StatusInternalServerError, "internal_error", "could not update user")
		return
	}
	c.JSON(http.StatusOK, u)
}

func (h *Handler) hasPermission(ctx context.Context, userID uuid.UUID, permission string) bool {
	allowed, err := h.authorize(ctx, userID, permission)
	return err == nil && allowed
}

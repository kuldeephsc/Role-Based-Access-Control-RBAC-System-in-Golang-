package audit

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"rbac-platform/internal/domain"
	"rbac-platform/internal/httpx"
	"rbac-platform/internal/middleware"
)

type Handler struct {
	repo      domain.AuditRepository
	authorize func(ctx context.Context, userID uuid.UUID, permission string) (bool, error)
}

func NewHandler(repo domain.AuditRepository, authorize func(ctx context.Context, userID uuid.UUID, permission string) (bool, error)) *Handler {
	return &Handler{repo: repo, authorize: authorize}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/audit", middleware.RequirePermission(h.authorize, "view_audit"), h.List)
}

func (h *Handler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	logs, total, err := h.repo.List(c.Request.Context(), limit, offset)
	if err != nil {
		httpx.Error(c, http.StatusInternalServerError, "internal_error", "could not list audit logs")
		return
	}
	c.JSON(http.StatusOK, gin.H{"logs": logs, "total": total})
}

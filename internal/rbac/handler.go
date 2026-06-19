package rbac

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"rbac-platform/internal/httpx"
	"rbac-platform/internal/middleware"
)

type Handler struct {
	svc       *Service
	authorize func(ctx context.Context, userID uuid.UUID, permission string) (bool, error)
}

func NewHandler(svc *Service, authorize func(ctx context.Context, userID uuid.UUID, permission string) (bool, error)) *Handler {
	return &Handler{svc: svc, authorize: authorize}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	// All RBAC mutations (creating/deleting roles & permissions, attaching
	// permissions, assigning roles) are gated behind the same "assign_role"
	// permission, which only the seeded admin role holds. Reads (listing
	// roles/permissions) and the authorize check itself just need a valid
	// session.
	manage := middleware.RequirePermission(h.authorize, "assign_role")

	rg.POST("/roles", manage, h.CreateRole)
	rg.GET("/roles", h.ListRoles)
	rg.DELETE("/roles/:id", manage, h.DeleteRole)

	rg.POST("/permissions", manage, h.CreatePermission)
	rg.GET("/permissions", h.ListPermissions)
	rg.DELETE("/permissions/:id", manage, h.DeletePermission)

	rg.POST("/users/:id/roles", manage, h.AssignRole)
	rg.DELETE("/users/:id/roles/:roleId", manage, h.RemoveRole)
	rg.POST("/roles/:id/permissions", manage, h.AttachPermission)

	rg.POST("/authorize", h.Authorize)
}

type createRoleRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

func (h *Handler) CreateRole(c *gin.Context) {
	var req createRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Error(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	role, err := h.svc.CreateRole(c.Request.Context(), req.Name, req.Description)
	if err != nil {
		httpx.Error(c, http.StatusInternalServerError, "internal_error", "could not create role")
		return
	}
	c.JSON(http.StatusCreated, role)
}

func (h *Handler) ListRoles(c *gin.Context) {
	roles, err := h.svc.ListRoles(c.Request.Context())
	if err != nil {
		httpx.Error(c, http.StatusInternalServerError, "internal_error", "could not list roles")
		return
	}
	c.JSON(http.StatusOK, roles)
}

func (h *Handler) DeleteRole(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.Error(c, http.StatusBadRequest, "invalid_id", "invalid role id")
		return
	}
	if err := h.svc.DeleteRole(c.Request.Context(), id); err != nil {
		httpx.Error(c, http.StatusInternalServerError, "internal_error", "could not delete role")
		return
	}
	c.Status(http.StatusNoContent)
}

type createPermissionRequest struct {
	Name        string `json:"name" binding:"required"`
	Resource    string `json:"resource" binding:"required"`
	Action      string `json:"action" binding:"required"`
	Description string `json:"description"`
}

func (h *Handler) CreatePermission(c *gin.Context) {
	var req createPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Error(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	p, err := h.svc.CreatePermission(c.Request.Context(), req.Name, req.Resource, req.Action, req.Description)
	if err != nil {
		httpx.Error(c, http.StatusInternalServerError, "internal_error", "could not create permission")
		return
	}
	c.JSON(http.StatusCreated, p)
}

func (h *Handler) ListPermissions(c *gin.Context) {
	perms, err := h.svc.ListPermissions(c.Request.Context())
	if err != nil {
		httpx.Error(c, http.StatusInternalServerError, "internal_error", "could not list permissions")
		return
	}
	c.JSON(http.StatusOK, perms)
}

func (h *Handler) DeletePermission(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.Error(c, http.StatusBadRequest, "invalid_id", "invalid permission id")
		return
	}
	if err := h.svc.DeletePermission(c.Request.Context(), id); err != nil {
		httpx.Error(c, http.StatusInternalServerError, "internal_error", "could not delete permission")
		return
	}
	c.Status(http.StatusNoContent)
}

type assignRoleRequest struct {
	RoleID string `json:"role_id" binding:"required"`
}

func (h *Handler) AssignRole(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.Error(c, http.StatusBadRequest, "invalid_id", "invalid user id")
		return
	}
	var req assignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Error(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		httpx.Error(c, http.StatusBadRequest, "invalid_id", "invalid role id")
		return
	}
	actorID, _ := c.Get("user_id")
	actor, _ := actorID.(uuid.UUID)

	if err := h.svc.AssignRoleToUser(c.Request.Context(), userID, roleID, actor); err != nil {
		httpx.Error(c, http.StatusInternalServerError, "internal_error", "could not assign role")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "role assigned"})
}

func (h *Handler) RemoveRole(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.Error(c, http.StatusBadRequest, "invalid_id", "invalid user id")
		return
	}
	roleID, err := uuid.Parse(c.Param("roleId"))
	if err != nil {
		httpx.Error(c, http.StatusBadRequest, "invalid_id", "invalid role id")
		return
	}
	if err := h.svc.RemoveRoleFromUser(c.Request.Context(), userID, roleID); err != nil {
		httpx.Error(c, http.StatusInternalServerError, "internal_error", "could not remove role")
		return
	}
	c.Status(http.StatusNoContent)
}

type attachPermissionRequest struct {
	PermissionID string `json:"permission_id" binding:"required"`
}

func (h *Handler) AttachPermission(c *gin.Context) {
	roleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.Error(c, http.StatusBadRequest, "invalid_id", "invalid role id")
		return
	}
	var req attachPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Error(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	permID, err := uuid.Parse(req.PermissionID)
	if err != nil {
		httpx.Error(c, http.StatusBadRequest, "invalid_id", "invalid permission id")
		return
	}
	if err := h.svc.AttachPermissionToRole(c.Request.Context(), roleID, permID); err != nil {
		httpx.Error(c, http.StatusInternalServerError, "internal_error", "could not attach permission")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "permission attached"})
}

type authorizeRequest struct {
	UserID     string `json:"user_id" binding:"required"`
	Permission string `json:"permission" binding:"required"`
}

func (h *Handler) Authorize(c *gin.Context) {
	var req authorizeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Error(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		httpx.Error(c, http.StatusBadRequest, "invalid_id", "invalid user id")
		return
	}
	allowed, err := h.authorize(c.Request.Context(), userID, req.Permission)
	if err != nil {
		httpx.Error(c, http.StatusInternalServerError, "internal_error", "could not check permission")
		return
	}
	c.JSON(http.StatusOK, gin.H{"allowed": allowed, "permission": req.Permission})
}

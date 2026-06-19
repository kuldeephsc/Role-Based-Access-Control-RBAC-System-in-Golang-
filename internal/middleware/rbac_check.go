package middleware

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"rbac-platform/internal/httpx"
)

// AuthorizeFunc matches rbac.Service.Authorize's signature. Handlers pass
// the service method in directly (e.g. middleware.RequirePermission(rbacSvc.Authorize, "assign_role"))
// without needing to depend on the rbac package's concrete type.
type AuthorizeFunc func(ctx context.Context, userID uuid.UUID, permission string) (bool, error)

// RequirePermission gates a route on a single named permission. The actual
// check always goes through Postgres (or Redis, from Phase 2 on) via the
// authorize function -- this middleware only handles extracting the caller
// and translating the result into the right HTTP status.
func RequirePermission(authorize AuthorizeFunc, permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDVal, exists := c.Get("user_id")
		if !exists {
			httpx.Error(c, http.StatusUnauthorized, "unauthenticated", "authentication required")
			c.Abort()
			return
		}
		userID, ok := userIDVal.(uuid.UUID)
		if !ok {
			httpx.Error(c, http.StatusUnauthorized, "unauthenticated", "authentication required")
			c.Abort()
			return
		}
		allowed, err := authorize(c.Request.Context(), userID, permission)
		if err != nil {
			httpx.Error(c, http.StatusInternalServerError, "internal_error", "could not verify permission")
			c.Abort()
			return
		}
		if !allowed {
			httpx.Error(c, http.StatusForbidden, "permission_denied", "you do not have permission to perform this action")
			c.Abort()
			return
		}
		c.Next()
	}
}

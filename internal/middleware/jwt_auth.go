package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"rbac-platform/internal/httpx"
	"rbac-platform/internal/platform/cache"
	"rbac-platform/internal/platform/jwt"
)

// JWTAuth validates the bearer access token, checks the Redis blacklist
// (failing open if Redis is unreachable -- see architecture spec §3.5),
// and puts the caller's user_id, roles, jti, and token expiry into the
// Gin context for downstream handlers/middleware.
func JWTAuth(mgr *jwt.Manager, blacklist *cache.Blacklist) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			httpx.Error(c, http.StatusUnauthorized, "missing_token", "authorization header is required")
			c.Abort()
			return
		}
		tokenStr := strings.TrimPrefix(header, "Bearer ")
		claims, err := mgr.Parse(tokenStr)
		if err != nil {
			httpx.Error(c, http.StatusUnauthorized, "invalid_token", "token is invalid or expired")
			c.Abort()
			return
		}
		if blacklist.IsBlacklisted(c.Request.Context(), claims.ID) {
			httpx.Error(c, http.StatusUnauthorized, "token_revoked", "this token has been revoked")
			c.Abort()
			return
		}
		c.Set("user_id", claims.UserID)
		c.Set("roles", claims.Roles)
		c.Set("jti", claims.ID)
		if claims.ExpiresAt != nil {
			c.Set("token_exp", claims.ExpiresAt.Time)
		}
		c.Next()
	}
}

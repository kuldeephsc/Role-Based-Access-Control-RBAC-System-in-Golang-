package httpx

import "github.com/gin-gonic/gin"

// Error writes the standard {"error": {"code", "message"}} envelope used
// across every handler package, so callers get one consistent error shape
// regardless of which module produced it.
func Error(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{"error": gin.H{"code": code, "message": message}})
}

package auth

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"rbac-platform/internal/domain"
	"rbac-platform/internal/httpx"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes mounts the routes that don't require an existing session.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/signup", h.Signup)
	rg.POST("/login", h.Login)
	rg.POST("/refresh", h.Refresh)
}

// RegisterProtectedRoutes mounts /logout, which needs to run behind
// middleware.JWTAuth so the caller's jti and expiry are already in the
// Gin context (used to blacklist the access token, not just revoke the
// refresh token).
func (h *Handler) RegisterProtectedRoutes(rg *gin.RouterGroup) {
	rg.POST("/logout", h.Logout)
}

type signupRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	FullName string `json:"full_name"`
}

func (h *Handler) Signup(c *gin.Context) {
	var req signupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Error(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	u, err := h.svc.Signup(c.Request.Context(), req.Email, req.Password, req.FullName)
	if err != nil {
		if errors.Is(err, domain.ErrAlreadyExists) {
			httpx.Error(c, http.StatusConflict, "email_taken", "an account with this email already exists")
			return
		}
		httpx.Error(c, http.StatusInternalServerError, "internal_error", "could not create account")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": u.ID, "email": u.Email})
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *Handler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Error(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	pair, err := h.svc.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		httpx.Error(c, http.StatusUnauthorized, "invalid_credentials", "email or password is incorrect")
		return
	}
	c.JSON(http.StatusOK, gin.H{"access_token": pair.AccessToken, "refresh_token": pair.RefreshToken})
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (h *Handler) Refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Error(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	pair, err := h.svc.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		httpx.Error(c, http.StatusUnauthorized, "invalid_refresh_token", "refresh token is invalid or expired")
		return
	}
	c.JSON(http.StatusOK, gin.H{"access_token": pair.AccessToken, "refresh_token": pair.RefreshToken})
}

func (h *Handler) Logout(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Error(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	jti, _ := c.Get("jti")
	jtiStr, _ := jti.(string)
	expVal, _ := c.Get("token_exp")
	exp, _ := expVal.(time.Time)
	remaining := time.Until(exp)

	_ = h.svc.Logout(c.Request.Context(), req.RefreshToken, jtiStr, remaining)
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

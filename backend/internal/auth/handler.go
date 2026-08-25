package auth

import (
	"PulseWatch/pkg/email"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	service AuthService
}

func NewAuthHandler(service AuthService) AuthHandler {
	return AuthHandler{service: service}
}

func (h AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	message, err := h.service.Register(req)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": message})
}

func (h AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.service.Login(req)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	expiryAccess, err := time.ParseDuration(os.Getenv("JWT_ACCESS_EXPIRY"))

	if err != nil {
		expiryAccess = 15 * time.Minute
	}

	expiryRefresh, err := time.ParseDuration(os.Getenv("JWT_REFRESH_EXPIRY"))

	if err != nil {
		expiryRefresh = 168 * time.Hour
	}

	accessMaxAge := int(expiryAccess.Seconds())
	refreshMaxAge := int(expiryRefresh.Seconds())

	isProduction := strings.EqualFold(os.Getenv("APP_ENV"), "production")

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("access_token", response.AccessToken, accessMaxAge, "/", "", isProduction, true)
	c.SetCookie("refresh_token", response.RefreshToken, refreshMaxAge, "/", "", isProduction, true)

	c.JSON(http.StatusOK, gin.H{"message": "Login Succesfully"})
}

func (h AuthHandler) Logout(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")

	if err == nil && refreshToken != "" {
		if err = h.service.Logout(refreshToken); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.SetCookie("refresh_token", "", -1, "/", "", false, true)
	c.SetCookie("access_token", "", -1, "/", "", false, true)

	c.JSON(200, gin.H{"message": "Logout succesfully"})
}

func (h AuthHandler) RefreshAccessToken(c *gin.Context) {
	token, err := c.Cookie("refresh_token")

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token not found"})
		return
	}

	accessToken, err := h.service.RefreshAccessToken(token)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	expiryAccess, err := time.ParseDuration(os.Getenv("JWT_ACCESS_EXPIRY"))

	if err != nil {
		expiryAccess = 15 * time.Minute
	}
	accessMaxAge := int(expiryAccess.Seconds())

	isProduction := strings.EqualFold(os.Getenv("APP_ENV"), "production")

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("access_token", accessToken, accessMaxAge, "/", "", isProduction, true)
	c.JSON(http.StatusOK, gin.H{"message": "New access Token granted"})
}

func (h *AuthHandler) GetProfile(c *gin.Context) {
	userId, exist := c.Get("userID")

	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	response, err := h.service.GetProfile(userId.(string))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": response})
}

func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	url, err := h.service.ForgotPassword(req.Email)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Only send email if user was found (url is not empty)
	if url != "" {
		if err := email.MailSender(req.Email, url); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	// Always return success to prevent email enumeration
	c.JSON(http.StatusOK, gin.H{"message": "If this email is registered, a reset link has been sent."})
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.ResetPassword(req.Password, req.Token); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}

func (h *AuthHandler) OauthLoginAndRegister(c *gin.Context) {
	var req OAuthRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.service.OauthLoginAndRegister(req)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	expiryAccess, err := time.ParseDuration(os.Getenv("JWT_ACCESS_EXPIRY"))
	if err != nil {
		expiryAccess = 15 * time.Minute
	}

	expiryRefresh, err := time.ParseDuration(os.Getenv("JWT_REFRESH_EXPIRY"))
	if err != nil {
		expiryRefresh = 168 * time.Hour
	}

	isProduction := strings.EqualFold(os.Getenv("APP_ENV"), "production")

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("access_token", response.AccessToken, int(expiryAccess.Seconds()), "/", "", isProduction, true)
	c.SetCookie("refresh_token", response.RefreshToken, int(expiryRefresh.Seconds()), "/", "", isProduction, true)

	c.JSON(http.StatusOK, gin.H{"message": "Login successfully"})
}

func (h AuthHandler) InitPasswordChange(c *gin.Context) {
	userId, exist := c.Get("userID")
	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	var req InitPasswordChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.InitPasswordChange(userId.(string), req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Verification code sent to your email"})
}

func (h AuthHandler) VerifyPasswordChange(c *gin.Context) {
	userId, exist := c.Get("userID")
	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	var req VerifyPasswordChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.VerifyPasswordChange(userId.(string), req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

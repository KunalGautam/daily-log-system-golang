package auth

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup, authSvc *AuthService, mw gin.HandlerFunc) {
	auth := rg.Group("/auth")

	auth.POST("/register", authSvc.HandleRegister)
	auth.POST("/login", authSvc.HandleLogin)
	auth.POST("/logout", authSvc.HandleLogout)
	auth.POST("/refresh", authSvc.HandleRefresh)
	auth.POST("/verify", authSvc.HandleVerifyEmail)
	auth.POST("/forgot-password", authSvc.HandleForgotPassword)
	auth.POST("/reset-password", authSvc.HandleResetPassword)
	auth.POST("/passkey/login", authSvc.HandlePasskeyLogin)

	protected := auth.Group("")
	protected.Use(mw)
	{
		protected.POST("/totp/setup", authSvc.HandleTOTPSetup)
		protected.POST("/totp/verify", authSvc.HandleTOTPVerify)
		protected.POST("/totp/disable", authSvc.HandleTOTPDisable)
		protected.POST("/passkey/register", authSvc.HandlePasskeyRegister)
		protected.GET("/sessions", authSvc.HandleListSessions)
		protected.DELETE("/sessions/:id", authSvc.HandleRevokeSession)
		protected.DELETE("/sessions", authSvc.HandleRevokeAllSessions)
	}
}

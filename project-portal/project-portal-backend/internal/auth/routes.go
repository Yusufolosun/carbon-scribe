package auth

import "github.com/gin-gonic/gin"

// RegisterAuthRoutes registers all auth routes with a router group
func RegisterAuthRoutes(router *gin.RouterGroup, handler *Handler, tokenManager *TokenManager) {
	// Public endpoints
	router.POST("/register", handler.Register)
	router.POST("/login", handler.Login)
	router.POST("/wallet-login", handler.WalletLogin)
	router.POST("/refresh", handler.RefreshToken)
	router.POST("/verify-email", handler.VerifyEmail)
	router.POST("/request-password-reset", handler.RequestPasswordReset)
	router.POST("/reset-password", handler.ResetPassword)
	router.POST("/wallet-challenge", handler.GenerateWalletChallenge)

	// Protected endpoints
	protected := router.Group("")
	protected.Use(AuthMiddleware(tokenManager))
	{
		protected.GET("/me", handler.GetProfile)
		protected.PUT("/me", handler.UpdateProfile)
		protected.POST("/change-password", handler.ChangePassword)
		protected.POST("/logout", handler.Logout)
	}
}

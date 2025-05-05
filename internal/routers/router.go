package routers

import (
	"cooperative-system/internal/config"
	"cooperative-system/internal/handlers"
	"cooperative-system/internal/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type servicehandlers struct {
	UserService    handlers.UserService
	MemberService  handlers.MemberService
	SavingsService handlers.SavingsService
}

func Newhandlers(db *gorm.DB) *servicehandlers {
	return &servicehandlers{
		UserService:    &handlers.UserHandler{DB: db},
		MemberService:  &handlers.MemberHandler{DB: db},
		SavingsService: &handlers.SavingsHandler{DB: db},
	}
}

func SetUpRoute(router *gin.Engine) {
	db := config.DB
	handler := Newhandlers(db)

	router.POST("/signup", handler.UserService.Signup)
	router.POST("/login", handler.UserService.Login)

	memberGroup := router.Group("/api/v1/members")
	memberGroup.Use(middleware.RequireAuth)

	{
		memberGroup.POST("", handler.MemberService.CreateMember)
		memberGroup.GET("/:id", handler.MemberService.GetMemberByID)
		memberGroup.PATCH("/:id", handler.MemberService.UpdateAMember)
		memberGroup.DELETE("/:id", handler.MemberService.DeleteAMember)

	}

	savingsGroup := router.Group("/api/v1/savings")
	savingsGroup.Use(middleware.RequireAuth)
	{
		savingsGroup.POST("", handler.SavingsService.CreateSavings)
		savingsGroup.GET("/:id", handler.SavingsService.GetSavingByID)
		savingsGroup.PUT("/:id", handler.SavingsService.UpdateSavings)
		savingsGroup.DELETE("/:id", handler.SavingsService.DeleteSavings)
	}

	adminGroup := router.Group("/api/v1/admins")
	adminGroup.Use(middleware.RequireAuth, middleware.RequireAdmin)
	{
		adminGroup.POST("", handlers.CreateAdmin)
		adminGroup.GET("/members", handler.MemberService.GetAllMembers)
		adminGroup.GET("/savings/:id", handler.SavingsService.GetTransactionsForMember)
	}

}

package routers

import (
	config "cooperative-system/conf"
	"cooperative-system/middleware"
	v1 "cooperative-system/routers/api/v1"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handlers struct {
	MemberService  v1.MemberService
	UserService    v1.UserService
	SavingsService v1.SavingsService
}

func NewHandlers(db *gorm.DB) *Handlers {
	return &Handlers{
		MemberService:  &v1.MemberHandler{DB: db},
		UserService:    &v1.UserHandler{DB: db},
		SavingsService: &v1.SavingsHandler{DB: db},
	}
}

func SetUpRoute(router *gin.Engine) {
	db := config.DB
	handler := NewHandlers(db)

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
		adminGroup.POST("", v1.CreateAdmin)
		adminGroup.GET("/members", handler.MemberService.GetAllMembers)
		adminGroup.GET("/savings/:id", handler.SavingsService.GetTransactionsForMember)
	}

}

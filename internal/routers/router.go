package routers

import (
	"cooperative-system/internal/config"
	"cooperative-system/internal/handlers"
	"cooperative-system/internal/middleware"
	"cooperative-system/internal/repository"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handlers struct {
	UserService    handlers.UserService
	MemberService  handlers.MemberService
	SavingsService handlers.SavingsService
	LoanService    handlers.LoanService
	AdminService   handlers.AdminService
}

// NewHandlers creates new handler instances
// Renamed from Newhandlers to NewHandlers for Go conventions
// Updated to use NewUserHandler, NewMemberHandler, and NewSavingsHandler constructors
func NewHandlers(db *gorm.DB) *Handlers {

	userRepo := repository.NewUserRepository(db)
	memberRepo := repository.NewMGormemberRepository(db)
	savingsRepo := repository.NewgormSavingsRepository(db)
	loanRepo := repository.NewGormLoanRepository(db)

	adminHandler := handlers.NewAdminHandler(userRepo, memberRepo)

	return &Handlers{
		UserService:    handlers.NewUserHandler(userRepo),
		MemberService:  handlers.NewMemberHandler(memberRepo),
		SavingsService: handlers.NewSavingsHandler(savingsRepo, memberRepo),
		LoanService:    handlers.NewLoanHandler(loanRepo, memberRepo),
		AdminService:   adminHandler,
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
		adminGroup.POST("", handler.AdminService.CreateAdmin)
		adminGroup.DELETE("", handler.AdminService.DeleteMember)
		adminGroup.GET("/members", handler.MemberService.GetAllMembers)
		adminGroup.GET("/savings/:id", handler.SavingsService.GetTransactionsForMember)
	}

	loanGroup := router.Group("/api/v1/loans")
	loanGroup.Use(middleware.RequireAuth)
	{
		loanGroup.POST("", handler.LoanService.ApplyLoan)
		loanGroup.GET("/:loan_id", handler.LoanService.TrackLoanApproval)
	}

}

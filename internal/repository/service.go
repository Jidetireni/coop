package repository

import (
	"cooperative-system/internal/models"

	"gorm.io/gorm"
)

type MemberRepository interface {
	CreateMemberWithSavings(member *models.Member, savings *models.Savings) (*models.Member, *models.Savings, string, error)
	FetchAll(members []models.Member) ([]models.Member, string, error)
	FetchByID(memberID string) (*models.Member, string, error)
	Update(member *models.Member, updateFields interface{}) (*models.Member, string, error)
	Delete(member *models.Member) (*models.Member, string, error)
	FetchMemberByUserID(userID uint) (*models.Member, string, error)
	FetchMemberByID(tx *gorm.DB, memberID string) (*models.Member, string, error)
}

type LoanRepository interface {
	CreateLoanWithInitialHistory(loan *models.Loan, loanHistory *models.LoanHistory) (*models.Loan, *models.LoanHistory, string, error)
	GetLoanByID(loanID string) (*models.Loan, string, error)
	GetLoanHistoryByID(loanID string) ([]models.LoanHistory, string, error)
	BeginTransaction() *gorm.DB
	// RollbackTransaction(tx *gorm.DB)
	// CommitTransaction(tx *gorm.DB) error
	GetLoanByIDForUpdate(tx *gorm.DB, loanID string) (*models.Loan, string, error)
	GetAllLoansByMemberID(tx *gorm.DB, memberID uint) ([]models.Loan, string, error)
	UpdateLoan(tx *gorm.DB, loan *models.Loan) (*models.Loan, string, error)
	CreateLoanHistory(tx *gorm.DB, loanHistory *models.LoanHistory) error
}

type UserRepository interface {
	CreateUser(user *models.User) (*models.User, string, error)
	FindUserByEmail(email string) (*models.User, string, error)
	UpdateUser(user *models.User, role string) (*models.User, string, error)
}

type SavingsRepository interface {
	CreateSavingsEntry(savings *models.Savings) (*models.Savings, string, error)
	FetchMemberByUserID(userID uint) (*models.Member, string, error)
	GetOrCreateSavings(member *models.Member, userID uint, amount int, description string) (*models.Savings, bool, string, error)
	UpdateSavings(savings *models.Savings, updateFields interface{}) (*models.Savings, string, error)
	CreateTransaction(transaction *models.SavingTransaction) (*models.SavingTransaction, string, error)
	GetSavingsByMemberID(memberID uint) (*models.Savings, string, error)
	DeleteSavings(savings *models.Savings) (*models.Savings, string, error)
	GetTransactionsByMemberID(memberID uint) ([]models.SavingTransaction, string, error)
	GetSavingsByMemberIDTx(tx *gorm.DB, memberID uint) (*models.Savings, string, error)
}

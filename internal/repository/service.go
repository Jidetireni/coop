package repository

import "cooperative-system/internal/models"

type MemberRepository interface {
	CreateMemberWithSavings(member *models.Member, savings *models.Savings) (*models.Member, *models.Savings, string, error)
	FetchAll(members []models.Member) ([]models.Member, string, error)
	FetchByID(memberID string) (*models.Member, string, error)
	Update(member *models.Member, updateFields interface{}) (*models.Member, string, error)
	Delete(member *models.Member) (*models.Member, string, error)
	FetchMemberByUserID(userID uint) (*models.Member, string, error)
}

type LoanRepository interface {
	CreateLoanRequestObject(loan *models.Loan) (*models.Loan, error)
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
}

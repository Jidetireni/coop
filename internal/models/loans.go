package models

import (
	"time"

	"gorm.io/gorm"
)

type Loan struct {
	gorm.Model
	MemberID       uint `gorm:"not null"`
	Description    string
	Type           string  `gorm:"not null"` // e.g., "personal", "business", "education"
	Amount         float64 `gorm:"not null"`
	InterestRate   float64 `gorm:"not null"`
	LoanTermMonths uint    `gorm:"not null"` // e.g., 12 for 1 year
	Status         string  `gorm:"not null"` // e.g., "pending", "approved", "rejected"
	// RepaymentSchedule string

	ApprovedBy        *uint
	ApprovalDate      *time.Time
	RejectionReason   string
	InstallmentAmount float64

	TotalRepayableAmount float64
	Member               Member `gorm:"foreignKey:MemberID"`
}

type LoanResponse struct {
	ID          uint      `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	MemberID    uint      `json:"member_id"`
	Description string    `json:"description"`
	Type        string    `json:"type"`

	Amount               float64    `json:"amount"`
	InterestRate         float64    `json:"interest_rate"`
	LoanTermMonths       uint       `json:"loan_term_months"`
	Status               string     `json:"status"`
	ApprovedBy           *uint      `json:"approved_by"`
	ApprovalDate         *time.Time `json:"approval_date"`
	RejectionReason      string     `json:"rejection_reason"`
	InstallmentAmount    float64    `json:"installment_amount"`
	TotalRepayableAmount float64    `json:"total_repayable_amount"`
}

func NewLoanResponse(loan *Loan) LoanResponse {
	return LoanResponse{
		ID:                   loan.ID,
		CreatedAt:            loan.CreatedAt,
		UpdatedAt:            loan.UpdatedAt,
		MemberID:             loan.MemberID,
		Description:          loan.Description,
		Type:                 loan.Type,
		Amount:               loan.Amount,
		InterestRate:         loan.InterestRate,
		LoanTermMonths:       loan.LoanTermMonths,
		Status:               loan.Status,
		ApprovedBy:           loan.ApprovedBy,
		ApprovalDate:         loan.ApprovalDate,
		RejectionReason:      loan.RejectionReason,
		InstallmentAmount:    loan.InstallmentAmount,
		TotalRepayableAmount: loan.TotalRepayableAmount,
	}
}

const (
	LoanStatusPending   = "pending"
	LoanStatusApproved  = "approved"
	LoanStatusActive    = "active"
	LoanStatusRejected  = "rejected"
	LoanStatusPaid      = "paid"
	LoanStatusDefaulted = "defaulted"
)

// You can also define allowed loan types here if you want them centralized
var AllowedLoanTypes = map[string]bool{
	"personal":  true,
	"business":  true,
	"education": true,
	// Add other valid loan types here
}

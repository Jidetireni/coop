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
	Member               Member    `gorm:"foreignKey:MemberID"`
	SubmittedAt          time.Time `gorm:"autoCreateTime"`
	ReviewedAt           *time.Time
	ApprovedAt           *time.Time
	RejectedAt           *time.Time
	LoanHistory          []LoanHistory `gorm:"foreignKey:LoanID"`
	DisbursedAt          *time.Time
	IsActive             bool
}

type LoanHistory struct {
	gorm.Model
	LoanID    uint
	Status    string
	ChangedAt time.Time `gorm:"autoCreateTime"`
	ChangedBy uint
	Remarks   string
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
	IsActive             bool       `json:"is_active"`
	ApprovedBy           *uint      `json:"approved_by,omitempty"`
	ApprovalDate         *time.Time `json:"approval_date,omitempty"`
	RejectionReason      string     `json:"rejection_reason,omitempty"`
	InstallmentAmount    float64    `json:"installment_amount"`
	TotalRepayableAmount float64    `json:"total_repayable_amount"`
	SubmittedAt          time.Time  `json:"submitted_at"`
	ReviewedAt           *time.Time `json:"reviewed_at,omitempty"`
	ApprovedAt           *time.Time `json:"approved_at,omitempty"`
	RejectedAt           *time.Time `json:"rejected_at,omitempty"`
	DisbursedAt          *time.Time `json:"disbursed_at,omitempty"`
	// LoanHistory          []LoanHistoryResponse `json:"loan_history"`
}

type LoanHistoryResponse struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	LoanID    uint      `json:"loan_id"`
	Status    string    `json:"status"`
	ChangedAt time.Time `json:"changed_at"`
	ChangedBy uint      `json:"changed_by"`
	Remarks   string    `json:"remarks"`
}

func NewLoanResponse(loan *Loan) LoanResponse {
	// histories := []LoanHistoryResponse{}
	// for _, h := range loan.LoanHistory {
	// 	histories = append(histories, LoanHistoryResponse{
	// 		Status:    h.Status,
	// 		ChangedAt: h.ChangedAt,
	// 		ChangedBy: h.ChangedBy,
	// 		Remarks:   h.Remarks,
	// 	})
	// }
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
		IsActive:             loan.IsActive, // Map the new field
		ApprovedBy:           loan.ApprovedBy,
		ApprovalDate:         loan.ApprovalDate,
		RejectionReason:      loan.RejectionReason,
		InstallmentAmount:    loan.InstallmentAmount,
		TotalRepayableAmount: loan.TotalRepayableAmount,
		SubmittedAt:          loan.SubmittedAt,
		ReviewedAt:           loan.ReviewedAt,
		ApprovedAt:           loan.ApprovedAt,
		RejectedAt:           loan.RejectedAt,
		DisbursedAt:          loan.DisbursedAt,
		// LoanHistory:          histories,
	}
}

const (
	LoanStatusPending   = "pending"
	LoanStatusApproved  = "approved"
	LoanStatusActive    = "active"
	LoanStatusRejected  = "rejected"
	LoanStatusPaid      = "paid"
	LoanStatusDefaulted = "defaulted"
	LoanStatusDisbursed = "disbursed"
)

// You can also define allowed loan types here if you want them centralized
var AllowedLoanTypes = map[string]bool{
	"personal":  true,
	"business":  true,
	"education": true,
	// Add other valid loan types here
}

func NewLoanHistoryResponse(loanHistory *LoanHistory) LoanHistoryResponse {
	return LoanHistoryResponse{
		ID:        loanHistory.ID,
		CreatedAt: loanHistory.CreatedAt,
		UpdatedAt: loanHistory.UpdatedAt,
		LoanID:    loanHistory.LoanID,

		Status:    loanHistory.Status,
		ChangedAt: loanHistory.ChangedAt,
		ChangedBy: loanHistory.ChangedBy,
		Remarks:   loanHistory.Remarks,
	}
}

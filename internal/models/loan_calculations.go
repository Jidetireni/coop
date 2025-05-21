package models

import (
	"errors"
)

const (
	MaxLoanToSavingsRatio = 2.0 // Maximum loan amount to savings ratio
	MaxActiveLoans        = 1   // Maximum number of active loans allowed

)

func GetInterestRate(loanType string, loanTermMonths uint) float64 {
	var calculatedInterestRate float64

	if loanType == "business" {
		if loanTermMonths > 12 {
			calculatedInterestRate = 0.07 // 7% interest rate for business loans over 12 months
		} else {
			calculatedInterestRate = 0.05 // 5% interest rate for business loans under or equal to 12 months
		}
	} else if loanType == "personal" {
		if loanTermMonths > 24 {
			calculatedInterestRate = 0.055 // 5.5% interest rate for personal loans over 24 months (corrected from 6% in comment)
		} else if loanTermMonths > 12 {
			calculatedInterestRate = 0.045 // 4.5% interest rate for personal loans between 13 and 24 months (corrected from 4% in comment)
		} else {
			calculatedInterestRate = 0.035 // 3.5% interest rate for personal loans under or equal to 12 months (corrected from 3% in comment)
		}
	} else {
		calculatedInterestRate = 0.05 // Default interest rate for other types of loans (e.g., 5%)
	}
	return calculatedInterestRate
}

func CalculateTotalRepayableAmount(principal float64, annualInterestRate float64, loanTermMonths uint) (float64, error) {
	if loanTermMonths == 0 {
		return 0, errors.New("loan term months cannot be zero")
	}

	monthlyInterestRate := annualInterestRate / 12
	totalRepayableAmount := principal * (1 + monthlyInterestRate*float64(loanTermMonths))
	return totalRepayableAmount, nil

}

func CalculateInstallmentAmount(totalRepayableAmount float64, loanTermMonths uint) (float64, error) {
	if loanTermMonths == 0 {
		return 0, errors.New("loan term months cannot be zero")
	}
	installmentAmount := totalRepayableAmount / float64(loanTermMonths)
	return installmentAmount, nil
}

func CheckLoanStatus(loan *Loan) (canProcess bool, message string, err error) {
	if loan == nil {
		return false, "loan data is nil", errors.New("cannot check status of nil loan")
	}

	switch loan.Status {
	case LoanStatusPending: // Assuming LoanStatusPending is "pending"
		return true, "Loan is pending and can be processed.", nil
	case LoanStatusApproved:
		return false, "Loan is already approved.", nil
	case LoanStatusRejected:
		return false, "Loan is already rejected.", nil
	// case LoanStatusDisbursed:
	//     return false, "Loan has already been disbursed.", nil
	case LoanStatusPaid:
		return false, "Loan has already been paid.", nil
	case LoanStatusDefaulted:
		return false, "Loan is defaulted.", nil
	default:
		return false, "Loan has an unknown or unprocessable status: " + loan.Status, errors.New("unprocessable loan status")
	}
}

func CheckLoanEligibility(requestedLoan *Loan, member *Member, savings *Savings, existingLoans []Loan) (bool, []string, error) {

	var reasons []string

	if savings == nil {
		reasons = append(reasons, "savings record not found")
	} else {
		loanLimit := savings.Balance * MaxLoanToSavingsRatio
		if requestedLoan.Amount > float64(loanLimit) {
			reasons = append(reasons, "requested loan exceeds twice the savings balance")
		}
	}

	activeLoanCount := 0
	hasDefaultedLoan := false

	for _, exitstingLoan := range existingLoans {

		if exitstingLoan.ID == requestedLoan.ID {
			continue // Skip the current loan being processed
		}

		if exitstingLoan.Status == LoanStatusDefaulted {
			hasDefaultedLoan = true
		}

		if exitstingLoan.Status == LoanStatusActive || exitstingLoan.Status == LoanStatusDisbursed || exitstingLoan.Status == LoanStatusApproved {
			activeLoanCount++
		}
	}

	if hasDefaultedLoan {
		reasons = append(reasons, "member has a defaulted loan")
	}

	if activeLoanCount >= MaxActiveLoans {
		reasons = append(reasons, "member has reached the maximum number of active loans")
	}

	if requestedLoan.Amount <= 0 {
		reasons = append(reasons, "requested loan amount must be greater than zero")
	}

	if len(reasons) > 0 {
		return false, reasons, errors.New("loan eligibility check failed")
	}

	return true, nil, nil

}

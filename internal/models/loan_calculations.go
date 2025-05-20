package models

import (
	"errors"
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

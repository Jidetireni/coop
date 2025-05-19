package repository

import (
	"cooperative-system/internal/models"
	"errors"

	"gorm.io/gorm"
)

// MemberRepository handles data access and transactions
type MemberRepository struct {
	db *gorm.DB
}

// NewMemberRepository creates a new member repository instance
func NewMemberRepository(db *gorm.DB) *MemberRepository {
	return &MemberRepository{db: db}
}

func (r *MemberRepository) CreateMemberWithSavings(member *models.Member, savings *models.Savings) (*models.Member, *models.Savings, string, error) {
	// Start a transaction
	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Error; err != nil {
		return nil, nil, "failed to start transaction", err
	}

	// Create the member in the database
	if err := tx.Create(member).Error; err != nil {
		tx.Rollback()
		return nil, nil, "failed to create member", err
	}

	// Set the member ID in the savings record
	savings.MemberID = member.ID

	// Create the savings record
	if err := tx.Create(savings).Error; err != nil {
		tx.Rollback()
		return nil, nil, "failed to create initial savings", err
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return nil, nil, "failed to commit transaction", err
	}

	return member, savings, "member created successfully", nil
}

func (r *MemberRepository) FetchAll(members []models.Member) ([]models.Member, string, error) {

	if err := r.db.Preload("User").Find(&members).Error; err != nil {
		return nil, "failed to fetch members", err
	}

	return members, "members fetched successfully", nil

}

func (r *MemberRepository) FetchByID(memberID string, member *models.Member) (*models.Member, string, error) {

	err := r.db.Where("id = ?", memberID).First(&member).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "member not found", err
		} else {
			return nil, "failed to fetch member", err
		}
	}

	return member, "member fetched successfully", nil

}

func (r *MemberRepository) Update(member *models.Member, updateFields interface{}) (*models.Member, string, error) {

	if err := r.db.Model(&member).Updates(updateFields).Error; err != nil {
		return nil, "failed to update member", err
	}

	return member, "member updated successfully", nil
}

func (r *MemberRepository) Delete(member *models.Member) (*models.Member, string, error) {
	if err := r.db.Delete(&member).Error; err != nil {
		return nil, "failed to delete member", err
	}

	return member, "member deleted successfully", nil
}

// FetchMemberByUserID retrieves a member by their UserID
func (r *MemberRepository) FetchMemberByUserID(userID uint) (*models.Member, string, error) {
	var member models.Member
	if err := r.db.Where("user_id = ?", userID).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "member not found for the given user ID", err
		}
		return nil, "failed to fetch member by user ID", err
	}
	return &member, "member fetched successfully", nil
}

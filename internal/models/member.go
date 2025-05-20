package models

import (
	"time"

	"gorm.io/gorm"
)

type Member struct {
	gorm.Model
	UserID      uint   `gorm:"unique"`
	Name        string `gorm:"not null"`
	ContactInfo string `gorm:"not null"`
	User        User   `gorm:"foreignKey:UserID"`
}

type MemberResponse struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	// DeletedAt   string `json:"deleted_at"`
	Name        string `json:"name"`
	ContactInfo string `json:"contact_info"`
	UserID      uint   `json:"user_id"`
}

func NewMemberResponse(member *Member) MemberResponse {
	return MemberResponse{
		ID:          member.ID,
		CreatedAt:   member.CreatedAt,
		UpdatedAt:   member.UpdatedAt,
		Name:        member.Name,
		ContactInfo: member.ContactInfo,
		UserID:      member.UserID,
	}
}

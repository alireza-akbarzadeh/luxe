package dto

import (
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
)

type UserResponse struct {
	ID             uint   `json:"id"`
	Email          string `json:"email"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	Role           string `json:"role"`
	Phone          string `json:"phone"`
	AvatarURL      string `json:"avatar_url,omitempty"`
	MembershipTier string `json:"membership_tier"`
	IsPlusActive   bool   `json:"is_plus_active"`
}

// ToUserResponse maps a user model to the public API shape.
func ToUserResponse(user *models.User) UserResponse {
	if user == nil {
		return UserResponse{MembershipTier: constants.MembershipTierFree}
	}
	active := isPlusActiveUser(user)
	tier := constants.MembershipTierFree
	if active {
		tier = constants.MembershipTierPlus
	}
	return UserResponse{
		ID:             user.ID,
		Email:          user.Email,
		FirstName:      user.FirstName,
		LastName:       user.LastName,
		Role:           user.Role,
		Phone:          user.Phone,
		AvatarURL:      user.AvatarURL,
		MembershipTier: tier,
		IsPlusActive:   active,
	}
}

func isPlusActiveUser(user *models.User) bool {
	if user.MembershipTier != constants.MembershipTierPlus {
		return false
	}
	if user.PlusExpiresAt != nil && user.PlusExpiresAt.Before(time.Now()) {
		return false
	}
	return true
}

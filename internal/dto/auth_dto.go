// Package dto represent the shape of data coming in/out of your API (request/response).
package dto

// RegisterRequest defines input for registration.
type RegisterRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
	FirstName string `json:"first_name" validate:"required,min=1,max=100"`
	LastName  string `json:"last_name" validate:"required,min=1,max=100"`
	Phone     string `json:"phone,omitempty" validate:"omitempty,e164"`
}

// LoginRequest defines input for login.
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// ChangePasswordRequest input
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required,min=6"`
	NewPassword     string `json:"new_password" binding:"required,min=6"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

type SessionResponse struct {
	ID         uint   `json:"id"`
	UserAgent  string `json:"user_agent"`
	IPAddress  string `json:"ip_address"`
	LastUsedAt string `json:"last_used_at"`
	CreatedAt  string `json:"created_at"`
	IsCurrent  bool   `json:"is_current"`
}

type SessionsResponseData struct {
	Sessions []SessionResponse `json:"sessions"`
}

type SessionsResponse struct {
	Success bool                 `json:"success"`
	Message string               `json:"message"`
	Data    SessionsResponseData `json:"data"`
}

type RevokeSessionsRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type LoginResponseData struct {
	User         UserResponse `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
}

type RegisterResponseData struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	User         UserResponse `json:"user"`
}

// RefreshResponseData defines the shape of token refresh response.
type RefreshResponseData struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// MessageResponse is a generic response with only a message.
type MessageResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type RegisterResponse struct {
	Success bool                 `json:"success"`
	Message string               `json:"message"`
	Data    RegisterResponseData `json:"data"`
}

type LoginResponse struct {
	Success bool              `json:"success"`
	Message string            `json:"message"`
	Data    LoginResponseData `json:"data"`
}

type RefreshResponse struct {
	Success bool                `json:"success"`
	Message string              `json:"message"`
	Data    RefreshResponseData `json:"data"`
}

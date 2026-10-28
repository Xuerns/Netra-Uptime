package auth

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email,max=50"`
	Password string `json:"password" binding:"required,min=8,max=50"`
	Name     string `json:"name" binding:"required,min=3,max=30,nohtml"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email,max=50"`
	Password string `json:"password" binding:"required,min=8,max=50"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type ProfileResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar"`
}

type UpdateProfileRequest struct {
	Name string `json:"name" binding:"required,min=3,max=30,nohtml"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email,max=50"`
}

type ResetPasswordRequest struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required,min=8,max=50"`
}

type InitPasswordChangeRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
}

type VerifyPasswordChangeRequest struct {
	Code        string `json:"code" binding:"required,len=6"`
	NewPassword string `json:"new_password" binding:"required,min=8,max=50"`
}

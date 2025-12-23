package dto

type ChangePasswordRequest struct {
	Username    string `json:"username"`
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}

type ResetPasswordRequest struct {
	Username    string `json:"username"`
	NewPassword string `json:"newPassword"`
}

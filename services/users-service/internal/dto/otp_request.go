package dto

type OTPRequest struct {
	Username string `json:"username"`
	OTP      string `json:"otp"`
}

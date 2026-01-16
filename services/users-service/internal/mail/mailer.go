package mail

import "log"

func SendOTP(email, otp string) {
	log.Printf("[MOCK EMAIL] Sending OTP %s to %s\n", otp, email)
}

func SendMagicLink(email, link string) {
	log.Printf("[MOCK EMAIL] Sending magic link to %s: %s\n", email, link)
}

func SendVerificationEmail(email, link string) {
	log.Printf("[MOCK EMAIL] Sending verification email to %s: %s\n", email, link)
}

func SendPasswordResetEmail(email, link string) {
	log.Printf("[MOCK EMAIL] Sending password reset email to %s: %s\n", email, link)
}
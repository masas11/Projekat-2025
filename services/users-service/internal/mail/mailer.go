package mail

import "log"

func SendOTP(email, otp string) {
	log.Printf("[MOCK EMAIL] Sending OTP %s to %s\n", otp, email)
}

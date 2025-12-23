package security

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)

type OTPEntry struct {
	Code      string
	ExpiresAt time.Time
}

func GenerateOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(900000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()+100000), nil
}

func IsExpired(e OTPEntry) bool {
	return time.Now().After(e.ExpiresAt)
}

package totp

import (
	"fmt"
	"time"

	"github.com/pquerna/otp/totp"
)

// GenerateTOTP generates a TOTP from a TOTP secret and time. If the time is nil, then time.Now will be used
func GenerateTOTP(secret string, t time.Time) (string, error) {
	if secret == "" {
		return "", fmt.Errorf("totp secret is required")
	}

	if t.Year() == 0 {
		t = time.Now()
	}

	totpCode, err := totp.GenerateCode(secret, t)

	if err != nil {
		return "", err
	}

	return totpCode, err
}

func RoundTime(now time.Time) time.Time {
	totpTime := now
	totpPeriod := 30
	totpNextTransition := totpPeriod - now.Second()%30
	if totpNextTransition < 5 {
		totpTime = now.Add(30 * time.Second)
	}
	return totpTime
}

package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
	"fantasy-esports-backend/pkg/logger"
)

func GenerateOTP() string {
	// Generate 6-digit OTP
	max := big.NewInt(999999)
	min := big.NewInt(100000)
	
	n, err := rand.Int(rand.Reader, max.Sub(max, min))
	if err != nil {
		// Fallback to time-based OTP if crypto/rand fails
		return fmt.Sprintf("%06d", time.Now().Unix()%1000000)
	}
	
	return fmt.Sprintf("%06d", n.Add(n, min).Int64())
}

func PrintOTPToConsole(mobile, otp string) {
	logger.Info("=== OTP GENERATED ===")
	logger.Info(fmt.Sprintf("Mobile: %s", mobile))
	logger.Info(fmt.Sprintf("OTP: %s", otp))
	logger.Info(fmt.Sprintf("Valid for: 5 minutes"))
	logger.Info("=====================")
}

func IsOTPExpired(createdAt time.Time) bool {
	return time.Since(createdAt) > 5*time.Minute
}

func ValidateOTP(providedOTP, actualOTP string, createdAt time.Time) bool {
	if IsOTPExpired(createdAt) {
		return false
	}
	return providedOTP == actualOTP
}

// For development/testing - always valid OTP
func IsDevelopmentOTP(otp string) bool {
	return otp == "123456"
}
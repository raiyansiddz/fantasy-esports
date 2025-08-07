package utils

import (
	"fmt"
	"time"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gin-gonic/gin"
)

type JWTClaims struct {
	UserID   int64  `json:"user_id"`
	Mobile   string `json:"mobile"`
	Role     string `json:"role"`
	TokenType string `json:"token_type"` // access or refresh
	jwt.RegisteredClaims
}

type AdminJWTClaims struct {
	AdminID  int64  `json:"admin_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

func GenerateTokens(userID int64, mobile string, jwtSecret string) (accessToken, refreshToken string, err error) {
	// Access token (1 hour)
	accessClaims := JWTClaims{
		UserID:   userID,
		Mobile:   mobile,
		Role:     "user",
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "fantasy-esports",
		},
	}
	
	accessTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessToken, err = accessTokenObj.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", "", err
	}
	
	// Refresh token (7 days)
	refreshClaims := JWTClaims{
		UserID:   userID,
		Mobile:   mobile,
		Role:     "user",
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "fantasy-esports",
		},
	}
	
	refreshTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err = refreshTokenObj.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", "", err
	}
	
	return accessToken, refreshToken, nil
}

func GenerateAdminTokens(adminID int64, username, role string, jwtSecret string) (accessToken string, err error) {
	claims := AdminJWTClaims{
		AdminID:  adminID,
		Username: username,
		Role:     role,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(4 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "fantasy-esports-admin",
		},
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

func ValidateToken(tokenString, jwtSecret string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})
	
	if err != nil {
		return nil, err
	}
	
	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}
	
	return nil, fmt.Errorf("invalid token")
}

func ValidateAdminToken(tokenString, jwtSecret string) (*AdminJWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AdminJWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})
	
	if err != nil {
		return nil, err
	}
	
	if claims, ok := token.Claims.(*AdminJWTClaims); ok && token.Valid {
		return claims, nil
	}
	
	return nil, fmt.Errorf("invalid admin token")
}
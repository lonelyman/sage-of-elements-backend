package appauth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Claims คือ "ข้อมูล" ที่เราจะฝังเข้าไปใน Token
type Claims struct {
	UserID uint   `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// Service คือ "บริษัทรักษาความปลอดภัย" ของเรา
type Service interface {
	HashPassword(password string) (string, error)
	ComparePassword(hashedPassword, plainPassword string) error
	GenerateTokens(userID uint, role string) (accessToken string, refreshToken string, err error)
	ValidateAccessToken(tokenString string) (*Claims, error)
	ValidateRefreshToken(tokenString string) (*Claims, error)
}

// authService คือ struct ที่ทำงานจริง
type authService struct {
	accessSecret  []byte
	refreshSecret []byte
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

// NewAuthService คือ "โรงงาน" สำหรับก่อตั้งบริษัทรักษาความปลอดภัย
func NewAuthService(accessSecret, refreshSecret string) Service {
	return &authService{
		accessSecret:  []byte(accessSecret),
		refreshSecret: []byte(refreshSecret),
		//accessExpiry:  15 * time.Minute,   // 👈 อายุ Access Token (15 นาที)
		accessExpiry:  7 * 24 * time.Hour, // 👈 อายุ Access Token (15 นาที)
		refreshExpiry: 7 * 24 * time.Hour, // 👈 อายุ Refresh Token (7 วัน)
	}
}

// --- Implementation of Service interface ---

func (s *authService) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func (s *authService) ComparePassword(hashedPassword, plainPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
}

func (s *authService) GenerateTokens(userID uint, role string) (string, string, error) {
	// 1. สร้าง Access Token (อายุสั้น)
	accessClaims := &Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.accessExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(s.accessSecret)
	if err != nil {
		return "", "", err
	}

	// 2. สร้าง Refresh Token (อายุยาว)
	refreshClaims := &Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.refreshExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString(s.refreshSecret)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *authService) ValidateAccessToken(tokenString string) (*Claims, error) {
	return s.validateToken(tokenString, s.accessSecret)
}

func (s *authService) ValidateRefreshToken(tokenString string) (*Claims, error) {
	return s.validateToken(tokenString, s.refreshSecret)
}

// (Helper function ที่ใช้ร่วมกัน)
func (s *authService) validateToken(tokenString string, secret []byte) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

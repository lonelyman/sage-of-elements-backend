package middleware

import (
	"strings"

	"sage-of-elements-backend/pkg/appauth"
	"sage-of-elements-backend/pkg/apperrors"
	"sage-of-elements-backend/pkg/appresponse"

	"github.com/gofiber/fiber/v2"
)

// AuthMiddleware คือ "โรงงาน" ที่สร้าง Middleware สำหรับยืนยันตัวตน
func AuthMiddleware(authService appauth.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1. ดึง Authorization Header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			appErr := apperrors.UnauthorizedError("Authorization header is required")
			return appresponse.Error(c, appErr)
		}

		// 2. ตรวจสอบรูปแบบ "Bearer [token]"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			appErr := apperrors.New(fiber.StatusUnauthorized, apperrors.ErrInvalidToken, "Invalid token format")
			return appresponse.Error(c, appErr)
		}
		tokenString := parts[1]

		// 3. ส่ง Token ไปให้ Auth Service ตรวจสอบ
		claims, err := authService.ValidateAccessToken(tokenString)
		if err != nil {
			appErr := apperrors.NewWithDetails(fiber.StatusUnauthorized, apperrors.ErrInvalidToken, "Invalid or expired token", err.Error())
			return appresponse.Error(c, appErr)
		}

		// 4. ถ้า Token ถูกต้อง, เก็บข้อมูล Claims ไว้ใน c.Locals()
		// เพื่อให้ Handler ที่อยู่ถัดไปสามารถดึงข้อมูลผู้ใช้ไปใช้ได้
		c.Locals("user_claims", claims) // พี่เปลี่ยนชื่อเป็น user_claims เพื่อความชัดเจนนะ

		// 5. อนุญาตให้ Request วิ่งต่อไปได้
		return c.Next()
	}
}

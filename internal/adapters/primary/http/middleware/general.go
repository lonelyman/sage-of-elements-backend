package middleware

import (
	"time"

	"sage-of-elements-backend/pkg/applogger"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// LoggerMiddleware คือ Middleware ที่จะคอยบันทึก Log ของทุก Request
func LoggerMiddleware(appLogger applogger.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next() // ปล่อยให้ Request วิ่งไปทำงานใน Handler ก่อน
		stop := time.Now()

		// หลังจาก Handler ทำงานเสร็จแล้ว ค่อยมาบันทึก Log
		appLogger.Info("HTTP Request",
			"method", c.Method(),
			"path", c.Path(),
			"status", c.Response().StatusCode(),
			"latency", stop.Sub(start).String(),
			"ip", c.IP(),
		)

		return err
	}
}

// CORSMiddleware คือ Middleware สำหรับจัดการ Cross-Origin Resource Sharing
func CORSMiddleware() fiber.Handler {
	return cors.New(cors.Config{
		// สำหรับ Development เราจะอนุญาตทุก Origin ไปก่อนเพื่อให้ทำงานง่าย
		AllowOrigins: "*",

		// อนุญาตให้ใช้ Method ที่จำเป็น
		AllowMethods: "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",

		// อนุญาตให้มี Header ที่จำเป็น (สำคัญมากสำหรับ Auth)
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	})
}

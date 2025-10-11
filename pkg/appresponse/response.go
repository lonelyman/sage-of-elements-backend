// pkg/response/response.go
package appresponse

import (
	"math"

	"sage-of-elements-backend/pkg/apperrors"

	"github.com/gofiber/fiber/v2"
)

// ====================================================================================
// Response Structs (พิมพ์เขียวของ Response)
// ====================================================================================

// Pagination คือพิมพ์เขียวสำหรับข้อมูลการแบ่งหน้า
// รองรับทั้ง Page-based และ Cursor-based โดยใช้ Pointer และ omitempty
type Pagination struct {
	// --- Page-based fields ---
	TotalRecords *int `json:"total_records,omitempty"`
	Limit        *int `json:"limit,omitempty"`
	Offset       *int `json:"offset,omitempty"`
	TotalPages   *int `json:"total_pages,omitempty"`
	CurrentPage  *int `json:"current_page,omitempty"`

	// --- Cursor-based fields ---
	NextCursor *string `json:"next_cursor,omitempty"`
	HasMore    *bool   `json:"has_more,omitempty"`
}

// ====================================================================================
// Constructors & Helpers (โรงงานสร้าง Response)
// ====================================================================================

// NewPagePagination คือโรงงานสำหรับสร้าง Page-based Pagination object
func NewPagePagination(totalRecords, limit, offset int) *Pagination {
	if limit <= 0 {
		limit = 1
	}
	if offset < 0 {
		offset = 0
	}
	totalPages := int(math.Ceil(float64(totalRecords) / float64(limit)))
	currentPage := int(math.Floor(float64(offset)/float64(limit))) + 1
	return &Pagination{
		TotalRecords: &totalRecords,
		Limit:        &limit,
		Offset:       &offset,
		TotalPages:   &totalPages,
		CurrentPage:  &currentPage,
	}
}

// ⭐️⭐️⭐️ เพิ่ม "โรงงาน" ใหม่สำหรับ Cursor-based! ⭐️⭐️⭐️
// NewCursorPagination คือโรงงานสำหรับสร้าง Cursor-based Pagination object
func NewCursorPagination(nextCursor string, hasMore bool) *Pagination {
	return &Pagination{
		NextCursor: &nextCursor,
		HasMore:    &hasMore,
	}
}

// Success คือ "ผู้ช่วย" หลักสำหรับส่ง Response เมื่อทำงานสำเร็จ
func Success(c *fiber.Ctx, httpStatus int, message string, data interface{}, pagination *Pagination) error {
	body := fiber.Map{
		"success": true,
		"message": message,
		"data":    data,
	}
	if pagination != nil {
		body["pagination"] = pagination
	}
	return c.Status(httpStatus).JSON(body)
}

// Error คือ "ผู้ช่วย" หลักสำหรับส่ง Error Response
func Error(c *fiber.Ctx, err *apperrors.AppError) error {
	return c.Status(err.HTTPStatus).JSON(fiber.Map{
		"success": false,
		"message": err.Message,
		"error": fiber.Map{
			"code":    err.Code,
			"details": err.Details,
		},
	})
}

// Message คือ "ผู้ช่วย" สำหรับส่งแค่ข้อความกลับไป
func Message(c *fiber.Ctx, httpStatus int, message string) error {
	return c.Status(httpStatus).JSON(fiber.Map{
		"success": true,
		"message": message,
	})
}

// NoContent คือ "ผู้ช่วย" สำหรับส่ง Response ที่ไม่มี Body กลับไป
func NoContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

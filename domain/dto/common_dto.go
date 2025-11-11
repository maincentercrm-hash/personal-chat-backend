// domain/dto/common_dto.go

package dto

// ErrorResponse สำหรับข้อผิดพลาดทั่วไป (ใช้ร่วมกันในหลาย endpoint)
type ErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// PaginationData ข้อมูลการแบ่งหน้า
type PaginationData struct {
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalPages int64 `json:"total_pages"`
}

// GenericResponse เป็น DTO พื้นฐานสำหรับทุก response
type GenericResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

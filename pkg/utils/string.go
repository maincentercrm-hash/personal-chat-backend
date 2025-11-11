// utils/string.go
package utils

import (
	"regexp"
	"strings"
)

// SplitCommaString แยกสตริงด้วยเครื่องหมายจุลภาค และตัดช่องว่าง
func SplitCommaString(s string) []string {
	if s == "" {
		return []string{}
	}

	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

// IsValidEmail ตรวจสอบรูปแบบอีเมล
func IsValidEmail(email string) bool {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	match, _ := regexp.MatchString(pattern, email)
	return match
}

// ... ฟังก์ชันอื่นๆ ...

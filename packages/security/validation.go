package security

import (
	"errors"
	"mime/multipart"
	"net/http"
	"regexp"
	"strings"
	"unicode"
)

var (
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	phoneRegex = regexp.MustCompile(`^\+[1-9]\d{1,14}$`)

	sqlInjectionPatterns = []string{
		"'",
		"--",
		"/*",
		"*/",
		"xp_",
		"sp_",
		"0x",
		"union",
		"select",
		"insert",
		"update",
		"delete",
		"drop",
		"create",
		"alter",
		"exec",
		"execute",
		"script",
		"javascript:",
		"onload=",
		"onerror=",
		"onclick=",
	}
)

func ValidateEmail(email string) error {
	email = strings.TrimSpace(strings.ToLower(email))
	if !emailRegex.MatchString(email) {
		return errors.New("invalid email format")
	}
	if len(email) > 255 {
		return errors.New("email too long")
	}
	return nil
}

func ValidatePhone(phone string) error {
	phone = strings.TrimSpace(phone)
	if !phoneRegex.MatchString(phone) {
		return errors.New("invalid phone format (use E.164 format)")
	}
	return nil
}

func ValidatePassword(password string) error {
	if len(password) < 12 {
		return errors.New("password must be at least 12 characters")
	}

	var hasUpper, hasLower, hasNumber, hasSpecial bool
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return errors.New("password must contain at least one lowercase letter")
	}
	if !hasNumber {
		return errors.New("password must contain at least one number")
	}
	if !hasSpecial {
		return errors.New("password must contain at least one special character")
	}

	return nil
}

func SanitizeInput(input string) string {
	input = strings.ReplaceAll(input, "\x00", "")
	input = strings.TrimSpace(input)

	if len(input) > 10000 {
		input = input[:10000]
	}

	for _, pattern := range sqlInjectionPatterns {
		if strings.Contains(strings.ToLower(input), pattern) {
			input = strings.ReplaceAll(input, pattern, "")
		}
	}

	return input
}

func ValidateFileUpload(file multipart.File, header *multipart.FileHeader, maxSize int64, allowedTypes []string) error {
	if header.Size > maxSize {
		return errors.New("file size exceeds limit")
	}

	buffer := make([]byte, 512)
	_, err := file.Read(buffer)
	if err != nil {
		return err
	}
	file.Seek(0, 0)

	contentType := http.DetectContentType(buffer)

	allowed := false
	for _, t := range allowedTypes {
		if contentType == t {
			allowed = true
			break
		}
	}

	if !allowed {
		return errors.New("file type not allowed")
	}

	return nil
}

func ContainsSQLInjectionPattern(input string) bool {
	lowered := strings.ToLower(input)
	for _, pattern := range sqlInjectionPatterns {
		if strings.Contains(lowered, pattern) {
			return true
		}
	}
	return false
}

func ContainsXSSPattern(input string) bool {
	xssPatterns := []string{
		"<script",
		"javascript:",
		"onerror=",
		"onload=",
		"onclick=",
		"<iframe",
		"<embed",
		"<object",
	}

	lowered := strings.ToLower(input)
	for _, pattern := range xssPatterns {
		if strings.Contains(lowered, pattern) {
			return true
		}
	}
	return false
}

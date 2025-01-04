package validec

import (
	"strings"

	"github.com/go-playground/validator/v10"
)

func validateStingNotBlankWhiteSpace(fl validator.FieldLevel) bool {
	val := fl.Field().String()
	trimmed := strings.TrimSpace(val)
	return trimmed != ""
}

// func ValidatePgTypeText(ft validator.FieldLevel)

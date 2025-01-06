package validec

import (
	"strings"

	"github.com/go-playground/validator/v10"
)

// func datetimeValidation(fl validator.FieldLevel) bool {
// 	format := "2006-01-02"
// 	_, err := time.Parse(format, fl.Field().String())
// 	return err == nil
// }

func datetimeValidation(fl validator.FieldLevel) bool {
	// Get the field value as a pointer to *time.Time

	t, ok := fl.Field().Interface().(Time)
	if !ok {
		panic("Type must be of type validec.Time")
	}
	if t.string == "" || !t.Time.IsZero() {
		return true
	}

	return false

}

func validateStingNotBlankWhiteSpace(fl validator.FieldLevel) bool {
	val := fl.Field().String()
	trimmed := strings.TrimSpace(val)
	return trimmed != ""
}

func validatePGTypeText(fl validator.FieldLevel) bool {
	val := fl.Field().String()
	trimmed := strings.TrimSpace(val)
	return trimmed != ""
}

// func ValidatePgTypeText(ft validator.FieldLevel)

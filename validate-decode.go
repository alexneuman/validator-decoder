package validec

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/schema"
	"github.com/jackc/pgx/v5/pgtype"
)

func init() {
	setValidTypes[validTypes]()
}

func setValidTypes[T any]() {
	var t T
	typ := reflect.TypeOf(t)
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldType := field.Type
		okStructFields[fieldType] = true
	}

}

var validStructs = make(map[reflect.Type]bool)
var okStructFields = make(map[reflect.Type]bool)

var decoder = new(DecoderParams{IgnoreUnknownKeys: true, ZeroEmpty: true})
var v = newValidator()

type ValidationError struct {
	Field   string
	Message string
	Val     string
}

// maps struct fields to error messages
// var fieldErrsKeys = make(map[string]map[string]map[string]string)
var fieldErrsKeys = make(map[string]map[string]map[string]string)
var defaultErrMap = make(map[string]string)

// Whether the struct be validated safely
func structIsValid[T any]() bool {
	var t T
	typ := reflect.TypeOf(t)
	_, ok := validStructs[typ]
	if ok {
		return true
	}
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldTyp := field.Type
		_, ok := okStructFields[fieldTyp]
		if ok {
			continue
		}
		return false

	}

	validStructs[typ] = true
	return true
}

// Registers a struct instance{} with a map of errors so that the Validator maps errors to a given validator
// t is an instance of the struct that is going to be validated
// fieldErrMap key follows the pattern:
//
//	fieldErrMap := map[string]string{
//		"FirstName.validator": "First Name is Required",
//		"_default.required":   "Field is Required",
//		"_default":           "This field is required",
//	}
func RegisterValidation[T any](t T, fieldErrMap map[string]string) {

	var errFieldsMap = make(map[string]map[string]string)
	errFieldsMap["_default"] = make(map[string]string)
	errMsg, ok := fieldErrMap["_default"]
	if ok {
		errFieldsMap["_default"][""] = errMsg
	}

	typ := reflect.TypeOf(t)
	structName := typ.Name()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldName := field.Name
		validatorNames := strings.Split(field.Tag.Get("validate"), ",")
		if validatorNames[0] == "" {
			continue
		}
		if errFieldsMap[fieldName] == nil {
			errFieldsMap[fieldName] = make(map[string]string)
		}
	validatorNamesLoop:
		for _, vn := range validatorNames {
			key := fmt.Sprintf("%s.%s", fieldName, vn)
			errMsg, ok := fieldErrMap[key]
			if ok {
				errFieldsMap[fieldName][vn] = errMsg
				continue validatorNamesLoop
			}

			defaultKey := fmt.Sprintf("_default.%s", vn)
			errMsg, ok = fieldErrMap[defaultKey]
			if ok {
				errFieldsMap["_default"][vn] = errMsg
				continue validatorNamesLoop
			}
			// fmt.Printf("[Struct: %s]: Error msg for %s.%s not found", structName, fieldName, vn)

		}

	}
	fieldErrsKeys[structName] = errFieldsMap
}

// Sets default error messages for a given validator across all structs
// ex:
//
//	errMap := map[string]{
//		"required": "This is field is required",
//		"email": "This must be na email",
//	}
func RegisterDefaultValidatorErrMsg(errMap map[string]string) error {
	for k, v := range errMap {
		defaultErrMap[k] = v
	}
	return nil

}

func getErrors(d any, errs []validator.FieldError) (map[string]ValidationError, error) {
	var errMsgs = make(map[string]ValidationError)
	var tag string
	var ok bool
	typ := reflect.TypeOf(d)
	structName := typ.Name()
	fieldErrKeys, fieldErrKeysSet := fieldErrsKeys[structName]
	defaultErrMsg, defaultErrMsgOK := fieldErrKeys["_default"][""]

	if fieldErrKeys == nil {
		fieldErrKeys = map[string]map[string]string{}

	}

	for _, err := range errs {

		errParam := err.Param()
		failureVal := err.Value()
		failureValStr := fmt.Sprintf("%v", failureVal)

		if errParam != "" {
			tag = err.Tag() + "=" + errParam
		} else {
			tag = err.Tag()

		}

		errField := err.Field()
		defaultUniversalErrMsg, defaultUniversalErrMsgFound := defaultErrMap[tag]

		var errMsg string

		// No error messages are set, use default, if any
		if !fieldErrKeysSet {
			// check for universal validator err msg first
			errMsg, ok = defaultErrMap[tag]

			// errMsgs[errField] = ve
		} else if errMsg, ok = fieldErrKeys[errField][tag]; ok {

		} else if errMsg, ok = fieldErrKeys["_default"][tag]; ok {

		} else if defaultErrMsgOK {
			errMsg = defaultErrMsg
		} else if defaultUniversalErrMsgFound {
			errMsg = defaultUniversalErrMsg

		}

		// set default message if exists

		// No error message
		ve := ValidationError{
			Field:   errField,
			Message: errMsg,
			Val:     failureValStr,
			// SuppliedValue: err.Value() ,

		}

		errMsgs[errField] = ve

	}

	return errMsgs, nil

}

func newValidator() *validator.Validate {
	v := validator.New()
	v.RegisterValidation("datetime", datetimeValidation)
	v.RegisterValidation("notblank", validateStingNotBlankWhiteSpace)
	return v
}

type DecoderParams struct {
	IgnoreUnknownKeys bool
	ZeroEmpty         bool
}

func new(p DecoderParams) *schema.Decoder {
	d := schema.NewDecoder()
	d.IgnoreUnknownKeys(p.IgnoreUnknownKeys)
	d.ZeroEmpty(p.ZeroEmpty)
	d.RegisterConverter(Time{}, timeConverter)
	// d.RegisterConverter(CustomTime{}, timeConverter)
	d.RegisterConverter(pgtype.Text{}, PgTypeTextConverter)
	d.RegisterConverter(pgtype.Int2{}, PgTypeInt2Converter)
	d.RegisterConverter(pgtype.Int4{}, PgTypeInt4Converter)
	d.RegisterConverter(pgtype.Date{}, PgTypeDateTimeConverter)

	return d
}

func Decode[T any](data map[string][]string) T {
	var t T
	decoder.Decode(&t, data)
	return t

}

func Validate(t any) error {
	err := v.Struct(t)

	return err

}

func DecodeValidate[T any](data url.Values) (T, map[string]ValidationError) {
	if !structIsValid[T]() {
		var t T
		panic(fmt.Sprintf("Type %T is not valid", t))
	}
	d := Decode[T](data)
	err := Validate(d)
	var errMap = map[string]ValidationError{}
	if err != nil {
		fieldErrs := err.(validator.ValidationErrors)
		errMap, err := getErrors(d, fieldErrs)
		if err != nil {
			panic(fmt.Sprintf("Failed to get errors from: %s", err))
		}
		return d, errMap
	}
	return d, errMap
}

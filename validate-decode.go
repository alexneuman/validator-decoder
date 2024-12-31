package validec

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/schema"
)

var decoder = new(DecoderParams{IgnoreUnknownKeys: true, ZeroEmpty: true})
var v = newValidator()

// maps struct fields to error messages
// var fieldErrsKeys = make(map[string]map[string]map[string]string)
var fieldErrsKeys = make(map[string]map[string]map[string]string)
var defaultErrMap = make(map[string]string)

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
			fmt.Printf("[Struct: %s]: Error msg for %s.%s not found", structName, fieldName, vn)

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

func getErrors(d any, errs []validator.FieldError) (map[string]string, error) {
	var errMsgs = make(map[string]string)
	var tag string
	typ := reflect.TypeOf(d)
	structName := typ.Name()
	fieldErrKeys, fieldErrKeysSet := fieldErrsKeys[structName]
	defaultErrMsg, defaultErrMsgOK := fieldErrKeys["_default"][""]
	if fieldErrKeys == nil {
		fieldErrKeys = map[string]map[string]string{}
		// return nil, nil
	}

	for _, err := range errs {
		errParam := err.Param()

		if errParam != "" {
			tag = err.Tag() + "=" + errParam
		} else {
			tag = err.Tag()

		}

		errField := err.Field()

		// No error messages are set, use default, if any
		if !fieldErrKeysSet {
			// check for universal validator err msg first
			errMsg, ok := defaultErrMap[tag]
			if !ok {
				errMsg = ""
			}
			errMsgs[errField] = errMsg
			continue
		}

		errMsg, ok := fieldErrKeys[errField][tag]
		if ok {
			errMsgs[errField] = errMsg
			continue
		}

		errMsg, ok = fieldErrKeys["_default"][tag]
		if ok {
			errMsgs[errField] = errMsg
			continue
		}

		if defaultErrMsgOK {
			errMsgs[errField] = defaultErrMsg
			continue
		}

	}

	return errMsgs, nil

}

func newValidator() *validator.Validate {
	v := validator.New()
	v.RegisterValidation("datetime", datetimeValidation)
	v.RegisterValidation("notblank", validateStingNotBlankWhiteSpace)
	return v
}

func datetimeValidation(fl validator.FieldLevel) bool {
	format := "2006-01-02"
	_, err := time.Parse(format, fl.Field().String())
	return err == nil
}

type DecoderParams struct {
	IgnoreUnknownKeys bool
	ZeroEmpty         bool
}

func new(p DecoderParams) *schema.Decoder {
	d := schema.NewDecoder()
	d.IgnoreUnknownKeys(p.IgnoreUnknownKeys)
	d.ZeroEmpty(p.ZeroEmpty)
	d.RegisterConverter(time.Time{}, timeConverter)
	// d.RegisterConverter(CustomTime{}, timeConverter)

	return d
}

// func Decode[T any](data url.Values) T {
// 	var t T
// 	decoder.Decode(&t, data)
// 	return t

// }

func Decode[T any](data map[string][]string) T {
	var t T
	decoder.Decode(&t, data)
	return t

}

func Validate(t any) error {
	err := v.Struct(t)

	return err

}

func DecodeValidate[T any](data url.Values) (T, map[string]string) {
	d := Decode[T](data)
	err := Validate(d)
	var errMap = map[string]string{}
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

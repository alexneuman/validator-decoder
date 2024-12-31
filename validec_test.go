package validec

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type TestData struct {
	FirstName           string    `decoder:""`
	Age                 int       `validate:"required"`
	FavNum              int       `validate:""`
	Bob                 string    `validate:"required,notblank"`
	CreatedAt           time.Time `validate:"required"`
	NotRequiredNotBlank string    `validate:"notblank"`
	MinTenChars         string    `validate:"required,min=10"`
}

func createTestDecoder(_ *testing.T) (map[string][]string, TestData) {
	var fixtureVals = map[string][]string{
		"FirstName":    {"Steve"},
		"UnknownField": {"UnknownPerson"},
		"Age":          {"55"},
		"FavNum":       {"Bob"},
		"CreatedAt":    {"2014-11-12"},
		"MinTenChars":  {"ABCDEFGHIJLMNOPQRSTUVWXYZ"},
	}

	fixtureResult := Decode[TestData](fixtureVals)
	// data :=
	return fixtureVals, fixtureResult
}

func TestDecoderResult(t *testing.T) {
	_, result := createTestDecoder(t)
	require.Equal(t, result.FirstName, "Steve")
	require.Equal(t, result.FavNum, 0)
	// require.NotContains(t, result, "UnknownPerson")

}

func TestValidator(t *testing.T) {
	_, result := createTestDecoder(t)
	require.Equal(t, result.Age, 55)
	err := Validate(result)
	require.NotNil(t, err)

}

func TestDecodeValidate(t *testing.T) {
	testMap, _ := createTestDecoder(t)
	data, errMap := DecodeValidate[TestData](testMap)
	require.IsType(t, TestData{}, data)
	require.Equal(t, errMap["Bob"].Message, "")

	testMap["Bob"] = []string{}
	data, errMap = DecodeValidate[TestData](testMap)
	require.Equal(t, errMap["Bob"].Message, "")

	testMap["Bob"] = []string{""}
	data, errMap = DecodeValidate[TestData](testMap)
	require.Equal(t, errMap["Bob"].Message, "")

	testMap["Bob"] = []string{"X"}
	data, errMap = DecodeValidate[TestData](testMap)
	require.Empty(t, errMap["Bob"].Message)

	testMap["FavNum"] = []string{"55"}
	data, errMap = DecodeValidate[TestData](testMap)
	require.Empty(t, errMap["FavNum"].Message)

}

func TestValidateWhitespaceRequired(t *testing.T) {
	errMsgs := map[string]string{
		"_default.notblank": "This field cannot be blank",
	}
	RegisterValidation(TestData{}, errMsgs)
	testMap, _ := createTestDecoder(t)
	testMap["Bob"] = []string{" "}
	_, errMap := DecodeValidate[TestData](testMap)
	require.Contains(t, errMap, "Bob")
	require.Equal(t, errMap["Bob"].Message, "This field cannot be blank")

}

func TestDecodeValidateDate(t *testing.T) {
	testMap, _ := createTestDecoder(t)
	data, errMap := DecodeValidate[TestData](testMap)

	require.NotContains(t, errMap, "CreatedAt")
	require.Equal(t, data.CreatedAt.IsZero(), false)

	testMap["CreatedAt"] = []string{}
	data, errMap = DecodeValidate[TestData](testMap)
	require.Contains(t, errMap, "CreatedAt")

	// invalid date
	testMap["CreatedAt"] = []string{"xxxxxxx"}
	data, errMap = DecodeValidate[TestData](testMap)
	require.Contains(t, errMap, "CreatedAt")
}

func TestDecoderResultWithErrorMsgs(t *testing.T) {
	errMsgs := map[string]string{
		"Age.required":       "Age is required",
		"CreatedAt.required": "CreatedAt is required",
	}
	RegisterValidation(TestData{}, errMsgs)

	var testMap = map[string][]string{
		"FirstName":    {"Steve"},
		"UnknownField": {"UnknownPerson"},
		"FavNum":       {"Bob"},
	}
	// testMap["Age"] = []string{}
	// testMap["CreatedAt"] = []string{}
	_, errMap := DecodeValidate[TestData](testMap)
	require.Equal(t, errMap["Age"].Message, "Age is required")
	require.Equal(t, errMap["CreatedAt"].Message, "CreatedAt is required")

}

func TestRegisterDefaultValidatorErrMsg(t *testing.T) {
	defaultErrMap := map[string]string{
		"notblank": "This field cannot be blank",
	}
	RegisterDefaultValidatorErrMsg(defaultErrMap)
	testMap, _ := createTestDecoder(t)
	testMap["NotRequiredNotBlank"] = []string{" "}
	_, errMap := DecodeValidate[TestData](testMap)
	require.Equal(t, errMap["NotRequiredNotBlank"].Message, "This field cannot be blank")

}

func TestRegisterDefaultValidatorErrMsgDefaultandSpecific(t *testing.T) {
	defaultErrMap := map[string]string{
		"notblank": "This field cannot be blank",
	}
	structErrMap := map[string]string{
		"required": "This field is required",
	}
	RegisterDefaultValidatorErrMsg(defaultErrMap)
	RegisterValidation(TestData{}, structErrMap)

	testMap, _ := createTestDecoder(t)
	testMap["NotRequiredNotBlank"] = []string{" "}
	_, errMap := DecodeValidate[TestData](testMap)
	require.Equal(t, errMap["NotRequiredNotBlank"].Message, "This field cannot be blank")

}

func TestErrorOnValidatorWithEqualSign(t *testing.T) {
	testMap, _ := createTestDecoder(t)
	testMap["MinTenChars"] = []string{"Not10"}
	_, errMap := DecodeValidate[TestData](testMap)
	require.Contains(t, errMap, "MinTenChars")

}

func TestErrorOnValidatorWithEqualSignAndErrMsgs(t *testing.T) {
	testMap, _ := createTestDecoder(t)
	structErrMap := map[string]string{
		"_default.required": "This field is required",
		"_default":          "This field is invalid",
	}
	RegisterValidation(TestData{}, structErrMap)
	testMap["MinTenChars"] = []string{"Not10"}
	_, errMap := DecodeValidate[TestData](testMap)
	require.Contains(t, errMap, "MinTenChars")

}

package validec

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type TestData struct {
	FirstName string    `decoder:""`
	Age       int       `validate:"required"`
	FavNum    int       `validate:""`
	Bob       string    `validate:"required"`
	TestDate  time.Time `validate:"required"`
}

func createTestDecoder(_ *testing.T) (map[string][]string, TestData) {
	var fixtureVals = map[string][]string{
		"FirstName":    {"Steve"},
		"UnknownField": {"UnknownPerson"},
		"Age":          {"55"},
		"FavNum":       {"Bob"},
		"TestDate":     {"2014-11-12"},
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
	require.Equal(t, errMap["Bob"], "")

	testMap["Bob"] = []string{}
	data, errMap = DecodeValidate[TestData](testMap)
	require.Equal(t, errMap["Bob"], "")

	testMap["Bob"] = []string{""}
	data, errMap = DecodeValidate[TestData](testMap)
	require.Equal(t, errMap["Bob"], "")

	testMap["Bob"] = []string{"X"}
	data, errMap = DecodeValidate[TestData](testMap)
	require.Empty(t, errMap["Bob"])

	testMap["FavNum"] = []string{"55"}
	data, errMap = DecodeValidate[TestData](testMap)
	require.Empty(t, errMap["FavNum"])

}

func TestDecodeValidateDate(t *testing.T) {
	testMap, _ := createTestDecoder(t)
	data, errMap := DecodeValidate[TestData](testMap)

	require.NotContains(t, errMap, "TestDate")
	require.Equal(t, data.TestDate.IsZero(), false)

	testMap["TestDate"] = []string{}
	data, errMap = DecodeValidate[TestData](testMap)
	require.Contains(t, errMap, "TestDate")

	// invalid date
	testMap["TestDate"] = []string{"xxxxxxx"}
	data, errMap = DecodeValidate[TestData](testMap)
	require.Contains(t, errMap, "TestDate")
}

func TestDecoderResultWithErrorMsgs(t *testing.T) {
	errMsgs := map[string]string{
		"Age.required":      "Age is required",
		"TestDate.required": "TestDate is required",
	}
	RegisterValidation(TestData{}, errMsgs)

	var testMap = map[string][]string{
		"FirstName":    {"Steve"},
		"UnknownField": {"UnknownPerson"},
		"FavNum":       {"Bob"},
	}
	// testMap["Age"] = []string{}
	// testMap["TestDate"] = []string{}
	_, errMap := DecodeValidate[TestData](testMap)
	require.Equal(t, errMap["Age"], "Age is required")
	require.Equal(t, errMap["TestDate"], "TestDate is required")

}

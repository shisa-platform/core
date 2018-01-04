package service

import (
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFieldMatchName(t *testing.T) {
	cut := Field{
		Name: "zalgo",
	}

	assert.True(t, cut.Match("zalgo"))
	assert.False(t, cut.Match("xxx"))
}

func TestFieldMatchRegex(t *testing.T) {
	cut := Field{
		Regex: regexp.MustCompile("zal"),
	}

	assert.True(t, cut.Match("zalgo"))
	assert.False(t, cut.Match("xxx"))
}

func TestFieldMatchPreferRegex(t *testing.T) {
	cut := Field{
		Name:  "xxx",
		Regex: regexp.MustCompile("zal"),
	}

	assert.True(t, cut.Match("zalgo"))
	assert.False(t, cut.Match("xxx"))
}

func TestFieldValidateNoValidator(t *testing.T) {
	cut := Field{
		Name: "zalgo",
	}

	assert.NoError(t, cut.Validate([]string(nil)))
	assert.NoError(t, cut.Validate([]string{}))
	assert.NoError(t, cut.Validate([]string{"foo"}))
}

func TestFieldValidate(t *testing.T) {
	cut := Field{
		Name:      "zalgo",
		Validator: FixedStringValidator{"he comes"}.Validate,
	}

	assert.NoError(t, cut.Validate([]string{"he comes"}))
	assert.NoError(t, cut.Validate([]string(nil)))
	assert.NoError(t, cut.Validate([]string{}))
	assert.Error(t, cut.Validate([]string{"foo"}))
}

func TestFieldValidateMultiplicity(t *testing.T) {
	cut := Field{
		Name:         "zalgo",
		Validator:    FixedStringValidator{"he comes"}.Validate,
		Multiplicity: 1,
	}

	assert.NoError(t, cut.Validate([]string{"he comes"}))
	assert.NoError(t, cut.Validate([]string(nil)))
	assert.NoError(t, cut.Validate([]string{}))
	assert.Error(t, cut.Validate([]string{"he comes", "foo"}))
	assert.Error(t, cut.Validate([]string{"foo"}))
}

func TestFixedStringValidator(t *testing.T) {
	cut := FixedStringValidator{
		Target: "zalgo",
	}

	assert.NoError(t, cut.Validate([]string{"zalgo"}))
	assert.NoError(t, cut.Validate([]string(nil)))
	assert.NoError(t, cut.Validate([]string{}))
	assert.Error(t, cut.Validate([]string{"he comes"}))
	assert.Error(t, cut.Validate([]string{"zalgo", "he comes"}))
	assert.Error(t, cut.Validate([]string{"foo"}))
}

func TestStringSliceValidator(t *testing.T) {
	cut := StringSliceValidator{
		Target: []string{"slithy", "he comes"},
	}

	assert.NoError(t, cut.Validate([]string{"he comes"}))
	assert.NoError(t, cut.Validate([]string(nil)))
	assert.NoError(t, cut.Validate([]string{}))
	assert.Error(t, cut.Validate([]string{"zalgo"}))
	assert.Error(t, cut.Validate([]string{"zalgo", "he comes"}))
	assert.Error(t, cut.Validate([]string{"foo"}))
}

func TestStringValidatorEmtpy(t *testing.T) {
	cut := StringValidator{}

	assert.NoError(t, cut.Validate([]string{"zalgo"}))
	assert.NoError(t, cut.Validate([]string{"he comes"}))
	assert.NoError(t, cut.Validate([]string{"zalgo", "he comes"}))
	assert.NoError(t, cut.Validate([]string{"foo"}))
	assert.NoError(t, cut.Validate([]string(nil)))
	assert.NoError(t, cut.Validate([]string{}))
}

func TestStringValidatorMin(t *testing.T) {
	cut := StringValidator{
		MinLen: 5,
	}

	assert.NoError(t, cut.Validate([]string{"zalgo"}))
	assert.NoError(t, cut.Validate([]string{"he comes"}))
	assert.NoError(t, cut.Validate([]string{"zalgo", "he comes"}))
	assert.NoError(t, cut.Validate([]string(nil)))
	assert.NoError(t, cut.Validate([]string{}))
	assert.Error(t, cut.Validate([]string{"foo", "zalgo"}))
}

func TestStringValidatorMax(t *testing.T) {
	cut := StringValidator{
		MaxLen: 3,
	}

	assert.NoError(t, cut.Validate([]string{"foo"}))
	assert.NoError(t, cut.Validate([]string(nil)))
	assert.NoError(t, cut.Validate([]string{}))
	assert.Error(t, cut.Validate([]string{"zalgo", "foo"}))
	assert.Error(t, cut.Validate([]string{"he comes"}))
	assert.Error(t, cut.Validate([]string{"zalgo", "he comes"}))
}

func TestStringValidatorMinMax(t *testing.T) {
	cut := StringValidator{
		MinLen: 3,
		MaxLen: 5,
	}

	assert.NoError(t, cut.Validate([]string{"zalgo"}))
	assert.NoError(t, cut.Validate([]string{"foo"}))
	assert.NoError(t, cut.Validate([]string(nil)))
	assert.NoError(t, cut.Validate([]string{}))
	assert.Error(t, cut.Validate([]string{"he comes"}))
	assert.Error(t, cut.Validate([]string{"zalgo", "he comes"}))
}

func TestIntValidatorEmpty(t *testing.T) {
	cut := IntValidator{}

	assert.NoError(t, cut.Validate([]string{"1"}))
	assert.NoError(t, cut.Validate([]string(nil)))
	assert.NoError(t, cut.Validate([]string{}))
	assert.Error(t, cut.Validate([]string{"one"}))
	assert.Error(t, cut.Validate([]string{"1", "forty two"}))
}

func TestIntValidatorMin(t *testing.T) {
	min := 5
	cut := IntValidator{
		Min: &min,
	}

	assert.NoError(t, cut.Validate([]string{"5", "666"}))
	assert.NoError(t, cut.Validate([]string(nil)))
	assert.NoError(t, cut.Validate([]string{}))
	assert.Error(t, cut.Validate([]string{"1"}))
	assert.Error(t, cut.Validate([]string{"one"}))
	assert.Error(t, cut.Validate([]string{"1", "forty two"}))
}

func TestIntValidatorMax(t *testing.T) {
	max := 5
	cut := IntValidator{
		Max: &max,
	}

	assert.NoError(t, cut.Validate([]string{"2", "3"}))
	assert.NoError(t, cut.Validate([]string(nil)))
	assert.NoError(t, cut.Validate([]string{}))
	assert.Error(t, cut.Validate([]string{"666"}))
	assert.Error(t, cut.Validate([]string{"one"}))
	assert.Error(t, cut.Validate([]string{"1", "forty two"}))
	assert.Error(t, cut.Validate([]string{"5", "666"}))
}

func TestIntValidatorMinMax(t *testing.T) {
	min := 2
	max := 5
	cut := IntValidator{
		Min: &min,
		Max: &max,
	}

	assert.NoError(t, cut.Validate([]string{"2", "3"}))
	assert.NoError(t, cut.Validate([]string(nil)))
	assert.NoError(t, cut.Validate([]string{}))
	assert.Error(t, cut.Validate([]string{"666"}))
	assert.Error(t, cut.Validate([]string{"one"}))
	assert.Error(t, cut.Validate([]string{"1", "forty two"}))
	assert.Error(t, cut.Validate([]string{"5", "666"}))
}

func TestBoolValidator(t *testing.T) {
	assert.NoError(t, BoolValidator([]string{}))
	assert.NoError(t, BoolValidator([]string(nil)))
	assert.NoError(t, BoolValidator([]string{"true", "false", "0", "1"}))
	assert.Error(t, BoolValidator([]string{"true", "false", "foo", "bar"}))
}

func TestTimestampValidator(t *testing.T) {
	cut := TimestampValidator{
		Format: time.RFC3339,
	}

	assert.NoError(t, cut.Validate([]string{"2020-07-24T20:00:00+09:00", "2020-08-09T20:00:00+09:00"}))
	assert.NoError(t, cut.Validate([]string(nil)))
	assert.NoError(t, cut.Validate([]string{}))
	assert.Error(t, cut.Validate([]string{"2020-07-24T20:00+09:00", "2020-08-09T20:00:00+09:00"}))
	assert.Error(t, cut.Validate([]string{"foo", "bar", "2020-07-24T20:00:00+09:00"}))
}

func TestTimestampValidatorMin(t *testing.T) {
	min := time.Date(2020, 7, 24, 20, 0, 0, 0, tokyo)
	cut := TimestampValidator{
		Format: time.RFC3339,
		Min:    &min,
	}

	assert.NoError(t, cut.Validate([]string{"2020-07-24T20:00:00+09:00", "2020-08-09T20:00:00+09:00"}))
	assert.NoError(t, cut.Validate([]string(nil)))
	assert.NoError(t, cut.Validate([]string{}))
	assert.Error(t, cut.Validate([]string{"2020-01-01T00:00:00+09:00"}))
	assert.Error(t, cut.Validate([]string{"2020-07-24T20:00+09:00", "2020-08-09T20:00:00+09:00"}))
	assert.Error(t, cut.Validate([]string{"foo", "bar", "2020-07-24T20:00:00+09:00"}))
}

func TestTimestampValidatorMax(t *testing.T) {
	max := time.Date(2020, 8, 9, 20, 0, 0, 0, tokyo)
	cut := TimestampValidator{
		Format: time.RFC3339,
		Max:    &max,
	}

	assert.NoError(t, cut.Validate([]string{"2020-07-24T20:00:00+09:00", "2020-08-09T20:00:00+09:00"}))
	assert.NoError(t, cut.Validate([]string(nil)))
	assert.NoError(t, cut.Validate([]string{}))
	assert.Error(t, cut.Validate([]string{"2020-09-10T10:15:00+09:00"}))
	assert.Error(t, cut.Validate([]string{"2020-07-24T20:00+09:00", "2020-08-09T20:00:00+09:00"}))
	assert.Error(t, cut.Validate([]string{"foo", "bar", "2020-07-24T20:00:00+09:00"}))
}

func TestTimestampValidatorMinMax(t *testing.T) {
	min := time.Date(2020, 7, 24, 20, 0, 0, 0, tokyo)
	max := time.Date(2020, 8, 9, 20, 0, 0, 0, tokyo)
	cut := TimestampValidator{
		Format: time.RFC3339,
		Min:    &min,
		Max:    &max,
	}

	assert.NoError(t, cut.Validate([]string{"2020-07-24T20:00:00+09:00", "2020-08-09T20:00:00+09:00"}))
	assert.NoError(t, cut.Validate([]string(nil)))
	assert.NoError(t, cut.Validate([]string{}))
	assert.Error(t, cut.Validate([]string{"2020-01-01T00:00:00+09:00"}))
	assert.Error(t, cut.Validate([]string{"2020-09-10T10:15:00+09:00"}))
	assert.Error(t, cut.Validate([]string{"2020-07-24T20:00+09:00", "2020-08-09T20:00:00+09:00"}))
	assert.Error(t, cut.Validate([]string{"foo", "bar", "2020-07-24T20:00:00+09:00"}))
}

func TestRegexValidator(t *testing.T) {
	cut := RegexValidator{
		Regex: regexp.MustCompile("zal"),
	}

	assert.NoError(t, cut.Validate([]string{"zalgo", "zalimba"}))
	assert.NoError(t, cut.Validate([]string(nil)))
	assert.NoError(t, cut.Validate([]string{}))
	assert.Error(t, cut.Validate([]string{"zalgo", "xxx", "zamboni"}))
}

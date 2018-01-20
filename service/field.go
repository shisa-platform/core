package service

import (
	"regexp"
	"strconv"
	"time"

	"github.com/ansel1/merry"
	"go.uber.org/multierr"
)

var (
	validationErr = merry.New("unexpected value").WithUserMessage("Value is not allowed")
)

// Validator ensures that the given strings all meet certain
// criteria.
type Validator func([]string) merry.Error

// Field is the schema to validate a query parameter or request
// body.
// Either `Name` or `Regex` _should_ be provided.
// Providing `Default` *requires* `Name`.
// If `Default` is provided and a matching input is not presented
// then a syntheitic value will be created.
type Field struct {
	Name         string         // Match keys by exact value
	Regex        *regexp.Regexp // Match keys by pattern
	Default      string         `json:",omitempty"` // Default for Name when no input
	Validator    Validator      `json:"-"`          // Optional validator of value(s)
	Multiplicity uint           `json:",omitempty"` // Value count, 0 is unlimited
	Required     bool           `json:",omitempty"` // Is this input mandatory?
}

// Match returns true if the given key name is for this Field
func (f Field) Match(name string) bool {
	if f.Regex != nil {
		return f.Regex.MatchString(name)
	}

	return f.Name == name
}

// Validate returns an error if all input values don't meet the
// criteria of `Field.Validator`.
func (f Field) Validate(values []string) merry.Error {
	if f.Multiplicity != 0 && uint(len(values)) > f.Multiplicity {
		return validationErr.Here()
	}
	if f.Validator != nil {
		return f.Validator(values)
	}

	return nil
}

// FixedStringValidator enforces that all input values match a
// fixed string
type FixedStringValidator struct {
	Target string
}

func (v FixedStringValidator) Validate(values []string) merry.Error {
	for _, value := range values {
		if v.Target != value {
			return validationErr.Here().WithValue("value", value)
		}
	}

	return nil
}

// StringSliceValidator enforces that all input values match at
// least one of the given strings
type StringSliceValidator struct {
	Target []string
}

func (v StringSliceValidator) Validate(values []string) merry.Error {
	var err error
	for _, value := range values {
		var found bool
		for _, t := range v.Target {
			if value == t {
				found = true
				break
			}
		}
		if !found {
			err1 := validationErr.Here().WithValue("value", value)
			err = multierr.Combine(err, err1)
		}
	}

	return merry.Wrap(err)
}

// StringValidator enforces that all input values have a certain
// length
type StringValidator struct {
	MinLen uint
	MaxLen uint
}

func (v StringValidator) Validate(values []string) merry.Error {
	var err error
	for _, value := range values {
		if v.MinLen != 0 && uint(len(value)) < v.MinLen {
			err1 := validationErr.Here().WithValue("value", value)
			err = multierr.Combine(err, err1)
			continue
		}
		if v.MaxLen != 0 && uint(len(value)) > v.MaxLen {
			err1 := validationErr.Here().WithValue("value", value)
			err = multierr.Combine(err, err1)
		}
	}

	return merry.Wrap(err)
}

// IntValidator enforces that all input values are parsable as
// integers.  It also optionally enforces that values are within
// a range.
type IntValidator struct {
	Min *int
	Max *int
}

func (v IntValidator) Validate(values []string) merry.Error {
	var err error
	for _, value := range values {
		i, parseErr := strconv.Atoi(value)
		if parseErr != nil {
			err1 := merry.Wrap(parseErr).WithUserMessage("Value is not allowed").WithValue("value", value)
			err = multierr.Combine(err, err1)
			continue
		}
		if v.Min != nil && i < *v.Min {
			err1 := validationErr.Here().WithValue("value", value)
			err = multierr.Combine(err, err1)
			continue
		}
		if v.Max != nil && i > *v.Max {
			err1 := validationErr.Here().WithValue("value", value)
			err = multierr.Combine(err, err1)
		}
	}

	return merry.Wrap(err)
}

// BoolValidator enforces that all input values are parsable as a
// boolean
func BoolValidator(values []string) merry.Error {
	var err error
	for _, value := range values {
		if _, parseErr := strconv.ParseBool(value); parseErr != nil {
			err1 := merry.Wrap(parseErr).WithUserMessage("Value is not allowed").WithValue("value", value)
			err = multierr.Combine(err, err1)
		}
	}

	return merry.Wrap(err)
}

// TimestampValidator enforces that all input values are parsable
// as a timestamp with a certain format.  It also optionally
// enforces the time value falls within a range.
type TimestampValidator struct {
	Format string
	Min    *time.Time
	Max    *time.Time
}

func (v TimestampValidator) Validate(values []string) merry.Error {
	var err error
	for _, value := range values {
		t, parseErr := time.Parse(v.Format, value)
		if parseErr != nil {
			err1 := merry.Wrap(parseErr).WithUserMessage("Value is not allowed").WithValue("value", value)
			err = multierr.Combine(err, err1)
			continue
		}
		if v.Min != nil && t.Before(*v.Min) {
			err1 := validationErr.Here().WithValue("value", value)
			err = multierr.Combine(err, err1)
			continue
		}
		if v.Max != nil && t.After(*v.Max) {
			err1 := validationErr.Here().WithValue("value", value)
			err = multierr.Combine(err, err1)
		}
	}

	return merry.Wrap(err)
}

// RegexValidator enforces that all input values match the given
// regular expression.
type RegexValidator struct {
	Regex *regexp.Regexp
}

func (v RegexValidator) Validate(values []string) merry.Error {
	var err error
	for _, value := range values {
		if !v.Regex.MatchString(value) {
			err1 := validationErr.Here().WithValue("value", value)
			err = multierr.Combine(err, err1)
		}
	}

	return merry.Wrap(err)
}

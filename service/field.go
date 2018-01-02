package service

import (
	"regexp"
	"time"

	"github.com/ansel1/merry"
	"go.uber.org/multierr"
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
	Name      string         // Match keys by exact value
	Regex     *regexp.Regexp // Match keys by pattern
	Default   string         // Default for Name when no input
	Validator Validator      // Optional validator of value(s)
	Required  bool           // Is this input mandatory?
}

// Match returns true if the given key name is for this Field
func (f *Field) Match(name string) bool {
	if f.Regex != nil {
		return f.Regex.MatchString(name)
	}

	return f.Name == name
}

// Validate returns an error if all input values don't meet the
// criteria of `Field.Validator`.
func (f *Field) Validate(value []string) merry.Error {
	if f.Validator != nil {
		return f.Validator(value)
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
			return merry.New("unexpected value").WithUserMessage("Value is not allowed").WithValue("value", value)
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
			err1 := merry.New("unexpected value").WithUserMessage("Value is not allowed").WithValue("value", value)
			multierr.Combine(err, err1)
		}
	}

	return merry.Wrap(err)
}

type StringValidator struct {
	MaxLen uint
	MinLen uint
}

type IntValidator struct {
	Min int
	Max int
}

type UintValidator struct {
	Min uint
	Max uint
}

func (v UintValidator) Validate(values []string) merry.Error {
	return nil
}

type BoolValidator struct {
	Expected bool
}

type TimestampValidator struct {
	Format string
	Min    time.Time
	Max    time.Time
}

type RegexValidator struct {
	Regex *regexp.Regexp
}

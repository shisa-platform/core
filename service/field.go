package service

import (
	"regexp"
	"time"

	"github.com/ansel1/merry"
)

type Validator func([]string) merry.Error

type Field struct {
	Name      string
	Regex     *regexp.Regexp
	Default   string
	Validator Validator
	Required  bool
}

func (f *Field) Match(name string) bool {
	if f.Regex != nil {
		return f.Regex.MatchString(name)
	}

	return f.Name == name
}

func (f *Field) Validate(value []string) merry.Error {
	if f.Validator != nil {
		return f.Validator(value)
	}

	return nil
}

type StringMatchValidator struct {
	Target string
}

func (v StringMatchValidator) Validate(values []string) merry.Error {
	if len(values) != 1 {
		return merry.New("expected 1 value").WithUserMessage("too many values")
	}
	if v.Target != values[0] {
		return merry.New("unexpected value").WithUserMessage("unsupported value").WithValue("value", values[0])
	}

	return nil
}

type StringSliceValidator struct {
	Target []string
}

func (v StringSliceValidator) Validate(values []string) merry.Error {
	for _, value := range values {
		var found bool
		for _, t := range v.Target {
			if value == t {
				found = true
				break
			}
		}
		if !found {

		}
	}

	return nil
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

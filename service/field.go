package service

import (
	"regexp"

	"github.com/ansel1/merry"
)

type Validator interface {
	Validate(string) merry.Error
}

type Field struct {
	Name      string
	Regex     *regexp.Regexp
	Default   string
	Validator Validator
	Required  bool
	Requires  []*Field
	Forbids   []*Field
}

func (f *Field) Match(name string) bool {
	if f.Regex != nil {
		return f.Regex.MatchString(name)
	}

	return f.Name == name
}

func (f *Field) Validate(value string) merry.Error {
	if f.Validator != nil {
		return f.Validator.Validate(value)
	}

	return nil
}

type UintValidator struct {
	Min uint
	Max uint
}

package httpx

import (
	"regexp"
	"strconv"
	"time"

	"github.com/ansel1/merry"
	"go.uber.org/multierr"

	"github.com/shisa-platform/core/errorx"
)

var (
	UnexpectedValue = merry.New("unexpected value")
)

// Validator ensures that the given strings all meet certain
// criteria.
type Validator func(QueryParameter) merry.Error

func (v Validator) InvokeSafely(p QueryParameter) (_ merry.Error, exception merry.Error) {
	defer errorx.CapturePanic(&exception, "panic in validator")

	return v(p), nil
}

// ParameterSchema is the schema to validate a query parameter or request.
// Either `Name` or `Regex` _should_ be provided.
// Providing `Default` *requires* `Name`.
// If `Default` is provided and a matching input is not presented
// then a syntheitic value will be created.
type ParameterSchema struct {
	Name         string         // Match keys by exact value
	Regex        *regexp.Regexp // Match keys by pattern
	Default      string         `json:",omitempty"` // Default value for `Name`
	Validator    Validator      `json:"-"`          // Optional validator of value(s)
	Multiplicity uint           `json:",omitempty"` // Value count, 0 is unlimited
	Required     bool           // Is this input mandatory?
}

// Match returns true if the given key name is for this ParameterSchema
func (f ParameterSchema) Match(name string) bool {
	if f.Regex != nil {
		return f.Regex.MatchString(name)
	}

	return f.Name == name
}

// Validate returns an error if all input values don't meet the
// criteria of `ParameterSchema.Validator`.  If the validator panics an
// error will be returned in the second result parameter.
func (f ParameterSchema) Validate(p QueryParameter) (merry.Error, merry.Error) {
	if f.Multiplicity != 0 && uint(len(p.Values)) > f.Multiplicity {
		return merry.New("too many values"), nil
	}
	if f.Validator != nil {
		return f.Validator.InvokeSafely(p)
	}

	return nil, nil
}

// FixedStringValidator enforces that all input values match a
// fixed string
type FixedStringValidator struct {
	Target string
}

func (v FixedStringValidator) Validate(p QueryParameter) merry.Error {
	for _, value := range p.Values {
		if v.Target != value {
			return UnexpectedValue.Append(value)
		}
	}

	return nil
}

// StringSliceValidator enforces that all input values match at
// least one of the given strings
type StringSliceValidator struct {
	Target []string
}

func (v StringSliceValidator) Validate(p QueryParameter) merry.Error {
	var err error
	for _, value := range p.Values {
		var found bool
		for _, t := range v.Target {
			if value == t {
				found = true
				break
			}
		}
		if !found {
			err1 := UnexpectedValue.Append(value)
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

func (v StringValidator) Validate(p QueryParameter) merry.Error {
	var err error
	for _, value := range p.Values {
		if v.MinLen != 0 && uint(len(value)) < v.MinLen {
			err1 := UnexpectedValue.Appendf("%q is shorter than minimum: %d", value, v.MinLen)
			err = multierr.Combine(err, err1)
			continue
		}
		if v.MaxLen != 0 && uint(len(value)) > v.MaxLen {
			err1 := UnexpectedValue.Appendf("%q is longer than maximum: %d", value, v.MaxLen)
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

func (v IntValidator) Validate(p QueryParameter) merry.Error {
	var err error
	for _, value := range p.Values {
		i, parseErr := strconv.Atoi(value)
		if parseErr != nil {
			err1 := UnexpectedValue.Appendf("%q is not an integer", value)
			err = multierr.Combine(err, err1)
			continue
		}
		if v.Min != nil && i < *v.Min {
			err1 := UnexpectedValue.Appendf("%q is less than minimum: %d", value, *v.Min)
			err = multierr.Combine(err, err1)
			continue
		}
		if v.Max != nil && i > *v.Max {
			err1 := UnexpectedValue.Appendf("%q is greater than maximum: %d", value, *v.Max)
			err = multierr.Combine(err, err1)
		}
	}

	return merry.Wrap(err)
}

// UIntValidator enforces that all input values are parsable as
// uints.  It also optionally enforces that values are within
// a range.
type UIntValidator struct {
	Min *uint
	Max *uint
}

func (v UIntValidator) Validate(p QueryParameter) merry.Error {
	var err error
	for _, value := range p.Values {
		i, parseErr := strconv.ParseUint(value, 10, 0)
		if parseErr != nil {
			err1 := UnexpectedValue.Appendf("%q is not a uint", value)
			err = multierr.Combine(err, err1)
			continue
		}
		if v.Min != nil && uint(i) < *v.Min {
			err1 := UnexpectedValue.Appendf("%q is less than minimum: %d", value, *v.Min)
			err = multierr.Combine(err, err1)
			continue
		}
		if v.Max != nil && uint(i) > *v.Max {
			err1 := UnexpectedValue.Appendf("%q is greater than maximum: %d", value, *v.Max)
			err = multierr.Combine(err, err1)
		}
	}

	return merry.Wrap(err)
}

// BoolValidator enforces that all input values are parsable as a
// boolean
func BoolValidator(p QueryParameter) merry.Error {
	var err error
	for _, value := range p.Values {
		if _, parseErr := strconv.ParseBool(value); parseErr != nil {
			err1 := UnexpectedValue.Appendf("%q is not a boolean", value)
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

func (v TimestampValidator) Validate(p QueryParameter) merry.Error {
	var err error
	for _, value := range p.Values {
		t, parseErr := time.Parse(v.Format, value)
		if parseErr != nil {
			err1 := UnexpectedValue.Appendf("%q is not a timestamp", value)
			err = multierr.Combine(err, err1)
			continue
		}
		if v.Min != nil && t.Before(*v.Min) {
			err1 := UnexpectedValue.Appendf("%q is before minimum: %s", value, v.Min)
			err = multierr.Combine(err, err1)
			continue
		}
		if v.Max != nil && t.After(*v.Max) {
			err1 := UnexpectedValue.Appendf("%q is after maximum: %s", value, v.Min)
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

func (v RegexValidator) Validate(p QueryParameter) merry.Error {
	var err error
	for _, value := range p.Values {
		if !v.Regex.MatchString(value) {
			err1 := UnexpectedValue.Append(value)
			err = multierr.Combine(err, err1)
		}
	}

	return merry.Wrap(err)
}

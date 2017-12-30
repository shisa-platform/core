package service

import (
	"github.com/ansel1/merry"
	"time"
)

// QueryParameter is a single query parameter passed to an
// endpoint.
type QueryParameter struct {
	Name    string   // the key of the query parameter pair
	Values  []string // the concatenated values
	Ordinal int      // the position in the query or -1 for a default
	Invalid bool     // the pair was not parsable or validated
	Unknown bool     // the pair was not matched by a field
}

func (p *QueryParameter) CSV(t *[]string) merry.Error {
	return nil
}

func (p *QueryParameter) Bool(t *bool) merry.Error {
	return nil
}

func (p *QueryParameter) BoolSlice(t *[]bool) merry.Error {
	return nil
}

func (p *QueryParameter) Time(t *time.Time, format string) merry.Error {
	return nil
}

func (p *QueryParameter) TimeSlice(t *[]time.Time, format string) merry.Error {
	return nil
}

func (p *QueryParameter) Uint8(t *uint8) merry.Error {
	return nil
}

func (p *QueryParameter) Uint8Slice(t *[]uint8) merry.Error {
	return nil
}

func (p *QueryParameter) Uint16(t *uint16) merry.Error {
	return nil
}

func (p *QueryParameter) Uint16Slice(t *[]uint16) merry.Error {
	return nil
}

func (p *QueryParameter) Uint32(t *uint32) merry.Error {
	return nil
}

func (p *QueryParameter) Uint32Slice(t *[]uint32) merry.Error {
	return nil
}

func (p *QueryParameter) Uint64(t *uint64) merry.Error {
	return nil
}

func (p *QueryParameter) Uint64Slice(t *[]uint64) merry.Error {
	return nil
}

func (p *QueryParameter) Uint(t *uint) merry.Error {
	return nil
}

func (p *QueryParameter) UintSlice(t *[]uint) merry.Error {
	return nil
}

func (p *QueryParameter) Int8(t *int8) merry.Error {
	return nil
}

func (p *QueryParameter) Int8Slice(t *[]int8) merry.Error {
	return nil
}

func (p *QueryParameter) Int16(t *int16) merry.Error {
	return nil
}

func (p *QueryParameter) Int16Slice(t *[]int16) merry.Error {
	return nil
}

func (p *QueryParameter) Int32(t *int32) merry.Error {
	return nil
}

func (p *QueryParameter) Int32Slice(t *[]int32) merry.Error {
	return nil
}

func (p *QueryParameter) Int64(t *int64) merry.Error {
	return nil
}

func (p *QueryParameter) Int64Slice(t *[]int64) merry.Error {
	return nil
}

func (p *QueryParameter) Int(t *int) merry.Error {
	return nil
}

func (p *QueryParameter) IntSlice(t *[]int) merry.Error {
	return nil
}

package service

import (
	"github.com/ansel1/merry"
	"time"
)

type QueryParameter struct {
	Name    string
	Values  []string
	Invalid bool
	Unknown bool
}

func (p *QueryParameter) String(t *string) merry.Error {
	return nil
}

func (p *QueryParameter) StringSlice(t *[]string) merry.Error {
	return nil
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

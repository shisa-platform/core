package service

import (
	"testing"
	"time"

	"github.com/ansel1/merry"
	"github.com/stretchr/testify/assert"
)

type queryParamConversionTestFixture struct {
	values   []string
	expected interface{}
}

var (
	// tokyo = mustLoadLocation("Asia/Tokyo")
	tokyo    = time.FixedZone("", 9*60*60)
	op       = time.Date(2020, 7, 24, 20, 0, 0, 0, tokyo)
	cl       = time.Date(2020, 8, 9, 20, 0, 0, 0, tokyo)
	fixtures = []queryParamConversionTestFixture{
		{[]string{"one,\"foo\",bar", "\"two\",baz,quux"}, []string{"one", "foo", "bar", "two", "baz", "quux"}},
		{[]string{"true"}, true},
		{[]string{"1"}, true},
		{[]string{"false"}, false},
		{[]string{"0"}, false},
		{[]string{"true", "false"}, []bool{true, false}},
		{[]string{"1", "0"}, []bool{true, false}},
		{[]string{"2020-07-24T20:00:00+09:00"}, op},
		{[]string{"2020-07-24T20:00:00+09:00", "2020-08-09T20:00:00+09:00"}, []time.Time{op, cl}},
		{[]string{"-127"}, int8(-127)},
		{[]string{"127", "-42"}, []int8{127, -42}},
		{[]string{"-127"}, int16(-127)},
		{[]string{"127", "-42"}, []int16{127, -42}},
		{[]string{"-127"}, int32(-127)},
		{[]string{"127", "-42"}, []int32{127, -42}},
		{[]string{"-127"}, int64(-127)},
		{[]string{"127", "-42"}, []int64{127, -42}},
		{[]string{"-127"}, int(-127)},
		{[]string{"127", "-42"}, []int{127, -42}},
		{[]string{"127"}, uint8(127)},
		{[]string{"127", "42"}, []uint8{127, 42}},
		{[]string{"127"}, uint16(127)},
		{[]string{"127", "42"}, []uint16{127, 42}},
		{[]string{"127"}, uint32(127)},
		{[]string{"127", "42"}, []uint32{127, 42}},
		{[]string{"127"}, uint64(127)},
		{[]string{"127", "42"}, []uint64{127, 42}},
		{[]string{"127"}, uint(127)},
		{[]string{"127", "42"}, []uint{127, 42}},
	}
	failures = []queryParamConversionTestFixture{
		{[]string{"foo,\",bar"}, []string(nil)},
		{[]string{"foo,bar", "bar,\",baz"}, []string(nil)},
		{[]string{"bliggle"}, false},
		{[]string{"bliggle"}, []bool(nil)},
		{[]string{"xxx"}, time.Time{}},
		{[]string{"xxx", "yyy"}, []time.Time(nil)},
		{[]string{"xxx"}, int(0)},
		{[]string{"xxx"}, []int(nil)},
		{[]string{"xxx"}, int8(0)},
		{[]string{"xxx"}, []int8(nil)},
		{[]string{"xxx"}, int16(0)},
		{[]string{"xxx"}, []int16(nil)},
		{[]string{"xxx"}, int32(0)},
		{[]string{"xxx"}, []int32(nil)},
		{[]string{"xxx"}, int64(0)},
		{[]string{"xxx"}, []int64(nil)},
		{[]string{"xxx"}, uint(0)},
		{[]string{"xxx"}, []uint(nil)},
		{[]string{"xxx"}, uint8(0)},
		{[]string{"xxx"}, []uint8(nil)},
		{[]string{"xxx"}, uint16(0)},
		{[]string{"xxx"}, []uint16(nil)},
		{[]string{"xxx"}, uint32(0)},
		{[]string{"xxx"}, []uint32(nil)},
		{[]string{"xxx"}, uint64(0)},
		{[]string{"xxx"}, []uint64(nil)},
	}
)

func assertParameterConversion(t *testing.T, fixture queryParamConversionTestFixture, fail bool) {
	t.Helper()

	cut := QueryParameter{
		Name:   "test",
		Values: fixture.values,
	}

	var err merry.Error
	var actual interface{}

	switch fixture.expected.(type) {
	case []string:
		var result []string
		err = cut.CSV(&result)
		actual = result
	case bool:
		var result bool
		err = cut.Bool(&result)
		actual = result
	case []bool:
		var result []bool
		err = cut.BoolSlice(&result)
		actual = result
	case time.Time:
		var result time.Time
		err = cut.Time(&result, time.RFC3339)
		actual = result
	case []time.Time:
		var result []time.Time
		err = cut.TimeSlice(&result, time.RFC3339)
		actual = result
	case uint8:
		var result uint8
		err = cut.Uint8(&result)
		actual = result
	case []uint8:
		var result []uint8
		err = cut.Uint8Slice(&result)
		actual = result
	case uint16:
		var result uint16
		err = cut.Uint16(&result)
		actual = result
	case []uint16:
		var result []uint16
		err = cut.Uint16Slice(&result)
		actual = result
	case uint32:
		var result uint32
		err = cut.Uint32(&result)
		actual = result
	case []uint32:
		var result []uint32
		err = cut.Uint32Slice(&result)
		actual = result
	case uint64:
		var result uint64
		err = cut.Uint64(&result)
		actual = result
	case []uint64:
		var result []uint64
		err = cut.Uint64Slice(&result)
		actual = result
	case uint:
		var result uint
		err = cut.Uint(&result)
		actual = result
	case []uint:
		var result []uint
		err = cut.UintSlice(&result)
		actual = result
	case int8:
		var result int8
		err = cut.Int8(&result)
		actual = result
	case []int8:
		var result []int8
		err = cut.Int8Slice(&result)
		actual = result
	case int16:
		var result int16
		err = cut.Int16(&result)
		actual = result
	case []int16:
		var result []int16
		err = cut.Int16Slice(&result)
		actual = result
	case int32:
		var result int32
		err = cut.Int32(&result)
		actual = result
	case []int32:
		var result []int32
		err = cut.Int32Slice(&result)
		actual = result
	case int64:
		var result int64
		err = cut.Int64(&result)
		actual = result
	case []int64:
		var result []int64
		err = cut.Int64Slice(&result)
		actual = result
	case int:
		var result int
		err = cut.Int(&result)
		actual = result
	case []int:
		var result []int
		err = cut.IntSlice(&result)
		actual = result
	}

	if fail {
		assert.Error(t, err)
	} else {
		assert.NoError(t, err)
	}
	assert.Equal(t, fixture.expected, actual)
}

func TestQueryParameterConversion(t *testing.T) {
	for _, fixture := range fixtures {
		assertParameterConversion(t, fixture, false)
	}
}

func TestQueryParameterConversionFailure(t *testing.T) {
	for _, fixture := range failures {
		assertParameterConversion(t, fixture, true)
	}
}

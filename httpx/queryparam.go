package httpx

import (
	"encoding/csv"
	"io"
	"strconv"
	"time"

	"github.com/ansel1/merry"
)

// QueryParameter is a single URL query parameter passed to an
// endpoint.
type QueryParameter struct {
	Name   string      // the key of the query parameter pair
	Values []string    // the concatenated values
	Err    merry.Error // the param is unparsable or invalid
}

type sliceReader struct {
	i, j   int64
	values []string
}

func (r *sliceReader) Read(p []byte) (n int, err error) {
	if r.i == int64(len(r.values)-1) && r.j >= int64(len(r.values[r.i])) {
		return 0, io.EOF
	}

	if r.j >= int64(len(r.values[r.i])) {
		p[n] = 0x0A
		n++
		r.i++
		r.j = 0
	}

	n += copy(p[n:], r.values[r.i][r.j:])
	r.j += int64(n)

	return
}

func (p *QueryParameter) CSV(v *[]string) merry.Error {
	reader := csv.NewReader(&sliceReader{values: p.Values})
	records, err := reader.ReadAll()
	if err != nil {
		return merry.Prepend(err, "query parameter: parse csv")
	}
	for _, fields := range records {
		*v = append(*v, fields...)
	}

	return nil
}

func (p *QueryParameter) Bool(v *bool) merry.Error {
	b, err := strconv.ParseBool(p.Values[0])
	*v = b
	return merry.Prepend(err, "query parameter: parse bool")
}

func (p *QueryParameter) BoolSlice(v *[]bool) merry.Error {
	for _, value := range p.Values {
		b, err := strconv.ParseBool(value)
		if err != nil {
			return merry.Prepend(err, "query parameter: parse bool slice")
		}
		*v = append(*v, b)
	}

	return nil
}

func (p *QueryParameter) Time(v *time.Time, format string) merry.Error {
	t, err := time.Parse(format, p.Values[0])
	*v = t
	return merry.Prepend(err, "query parameter: parse timestamp")
}

func (p *QueryParameter) TimeSlice(v *[]time.Time, format string) merry.Error {
	for _, value := range p.Values {
		t, err := time.Parse(format, value)
		if err != nil {
			return merry.Prepend(err, "query parameter: parse timestamp slice")
		}
		*v = append(*v, t)
	}

	return nil
}

func (p *QueryParameter) Int8(v *int8) merry.Error {
	i, err := strconv.ParseInt(p.Values[0], 10, 8)
	*v = int8(i)
	return merry.Prepend(err, "query parameter: parse int8")
}

func (p *QueryParameter) Int8Slice(v *[]int8) merry.Error {
	for _, value := range p.Values {
		i, err := strconv.ParseInt(value, 10, 8)
		if err != nil {
			return merry.Prepend(err, "query parameter: parse int8 slice")
		}
		*v = append(*v, int8(i))
	}

	return nil
}

func (p *QueryParameter) Int16(v *int16) merry.Error {
	i, err := strconv.ParseInt(p.Values[0], 10, 16)
	*v = int16(i)
	return merry.Prepend(err, "query parameter: parse int16")
}

func (p *QueryParameter) Int16Slice(v *[]int16) merry.Error {
	for _, value := range p.Values {
		i, err := strconv.ParseInt(value, 10, 16)
		if err != nil {
			return merry.Prepend(err, "query parameter: parse int16 slice")
		}
		*v = append(*v, int16(i))
	}

	return nil
}

func (p *QueryParameter) Int32(v *int32) merry.Error {
	i, err := strconv.ParseInt(p.Values[0], 10, 32)
	*v = int32(i)
	return merry.Prepend(err, "query parameter: parse int32")
}

func (p *QueryParameter) Int32Slice(v *[]int32) merry.Error {
	for _, value := range p.Values {
		i, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return merry.Prepend(err, "query parameter: parse int32 slice")
		}
		*v = append(*v, int32(i))
	}

	return nil
}

func (p *QueryParameter) Int64(v *int64) merry.Error {
	i, err := strconv.ParseInt(p.Values[0], 10, 64)
	*v = int64(i)
	return merry.Prepend(err, "query parameter: parse int64")
}

func (p *QueryParameter) Int64Slice(v *[]int64) merry.Error {
	for _, value := range p.Values {
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return merry.Prepend(err, "query parameter: parse int64 slice")
		}
		*v = append(*v, int64(i))
	}

	return nil
}

func (p *QueryParameter) Int(v *int) merry.Error {
	i, err := strconv.Atoi(p.Values[0])
	*v = i
	return merry.Prepend(err, "query parameter: parse int")
}

func (p *QueryParameter) IntSlice(v *[]int) merry.Error {
	for _, value := range p.Values {
		i, err := strconv.Atoi(value)
		if err != nil {
			return merry.Prepend(err, "query parameter: parse int slice")
		}
		*v = append(*v, i)
	}

	return nil
}

func (p *QueryParameter) Uint8(v *uint8) merry.Error {
	i, err := strconv.ParseUint(p.Values[0], 10, 8)
	*v = uint8(i)
	return merry.Prepend(err, "query parameter: parse uint8")
}

func (p *QueryParameter) Uint8Slice(v *[]uint8) merry.Error {
	for _, value := range p.Values {
		i, err := strconv.ParseUint(value, 10, 8)
		if err != nil {
			return merry.Prepend(err, "query parameter: parse uint8 slice")
		}
		*v = append(*v, uint8(i))
	}

	return nil
}

func (p *QueryParameter) Uint16(v *uint16) merry.Error {
	i, err := strconv.ParseUint(p.Values[0], 10, 16)
	*v = uint16(i)
	return merry.Prepend(err, "query parameter: parse uint16")
}

func (p *QueryParameter) Uint16Slice(v *[]uint16) merry.Error {
	for _, value := range p.Values {
		i, err := strconv.ParseUint(value, 10, 16)
		if err != nil {
			return merry.Prepend(err, "query parameter: parse uint16 slice")
		}
		*v = append(*v, uint16(i))
	}

	return nil
}

func (p *QueryParameter) Uint32(v *uint32) merry.Error {
	i, err := strconv.ParseUint(p.Values[0], 10, 32)
	*v = uint32(i)
	return merry.Prepend(err, "query parameter: parse uint32")
}

func (p *QueryParameter) Uint32Slice(v *[]uint32) merry.Error {
	for _, value := range p.Values {
		i, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			return merry.Prepend(err, "query parameter: parse uint32 slice")
		}
		*v = append(*v, uint32(i))
	}

	return nil
}

func (p *QueryParameter) Uint64(v *uint64) merry.Error {
	i, err := strconv.ParseUint(p.Values[0], 10, 64)
	*v = uint64(i)
	return merry.Prepend(err, "query parameter: parse uint64")
}

func (p *QueryParameter) Uint64Slice(v *[]uint64) merry.Error {
	for _, value := range p.Values {
		i, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return merry.Prepend(err, "query parameter: parse uint64 slice")
		}
		*v = append(*v, uint64(i))
	}

	return nil
}

func (p *QueryParameter) Uint(v *uint) merry.Error {
	i, err := strconv.ParseUint(p.Values[0], 10, 0)
	*v = uint(i)
	return merry.Prepend(err, "query parameter: parse uint")
}

func (p *QueryParameter) UintSlice(v *[]uint) merry.Error {
	for _, value := range p.Values {
		i, err := strconv.ParseUint(value, 10, 0)
		if err != nil {
			return merry.Prepend(err, "query parameter: parse uint slice")
		}
		*v = append(*v, uint(i))
	}

	return nil
}

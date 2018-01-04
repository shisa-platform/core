package ratelimit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const day = time.Hour * 24

type rateLimitStringCase struct {
	Expected string
	Instance RateLimit
}

func TestRateLimit_String(t *testing.T) {
	cases := []rateLimitStringCase{
		{"10/d", RateLimit{10, day}},
		{"50/h", RateLimit{50, time.Hour}},
		{"100/m", RateLimit{100, time.Minute}},
		{"1000/s", RateLimit{1000, time.Second}},
	}

	for _, c := range cases {
		s := c.Instance.String()
		assert.Equal(t, c.Expected, s)
	}
}

type fromStringCase struct {
	Val         string
	Limit       int
	Period      time.Duration
	ExpectError bool
}

func TestFromString(t *testing.T) {
	cases := []fromStringCase{
		{"nomatch", 0, 0, true},
		{"gibberish100/dgibberish", 0, 0, true},
		{"100/d100/d100/d100/d100/d", 0, 0, true},
		{"-9000/h", 0, 0, true},
		{"0/dz", 0, 0, true},
		{"0/d", 0, day, false},
		{"10/d", 10, day, false},
		{"10/h", 10, time.Hour, false},
		{"20/m", 20, time.Minute, false},
		{"500/s", 500, time.Second, false},
	}

	for _, c := range cases {
		r, err := FromString(c.Val)
		if c.ExpectError {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}
		assert.Equal(t, c.Limit, r.Limit)
		assert.Equal(t, c.Period, r.Period)
	}
}

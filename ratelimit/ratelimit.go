package ratelimit

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/ansel1/merry"
)

var limitRegex = regexp.MustCompile(`^(\d+)/([dhms])$`)

// RateLimit encodes a maximum number of repititions allowed over a time interval.
type RateLimit struct {
	Limit  int
	Period time.Duration
}

// FromString parses a rate limit string in the form of "<limit>/<duration>".
func FromString(value string) (r RateLimit, merr merry.Error) {
	matches := limitRegex.FindAllStringSubmatch(value, 2)
	if matches == nil || len(matches) == 0 || len(matches[0]) != 3 {
		merr = merry.Errorf("%q is not a valid rate limit expression", value)
		return
	}

	limit, err := strconv.Atoi(matches[0][1])
	if err != nil {
		merr = merry.Errorf("%q is not a valid rate limit expression", value)
		return
	}
	r.Limit = limit

	switch matches[0][2] {
	case "d":
		r.Period = time.Hour * 24
	case "h":
		r.Period = time.Hour
	case "m":
		r.Period = time.Minute
	case "s":
		r.Period = time.Second
	}

	return
}

func (r RateLimit) String() string {
	var period string
	switch {
	case r.Period.Hours() > 1:
		period = "d"
	case r.Period.Hours() == 1:
		period = "h"
	case r.Period.Minutes() == 1:
		period = "m"
	default:
		period = "s"
	}

	return fmt.Sprintf("%d/%s", r.Limit, period)
}

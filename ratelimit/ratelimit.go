package ratelimit

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

var limitRegex = regexp.MustCompile(`(\d+)/(d|h|m|s)`)

type RateLimit struct {
	Limit  int
	Period time.Duration
}

func FromString(value string) (RateLimit, error) {
	result := RateLimit{}

	matches := limitRegex.FindAllStringSubmatch(value, 2)
	if matches == nil || len(matches) == 0 || len(matches[0]) != 3 {
		return result, fmt.Errorf("%q is not a valid rate limit expression", value)
	}

	limit, err := strconv.Atoi(matches[0][1])
	if err != nil || 0 > limit {
		return result, fmt.Errorf("%q is not a valid rate limit expression", value)
	}
	result.Limit = limit

	switch matches[0][2] {
	case "d":
		result.Period = time.Hour * 24
	case "h":
		result.Period = time.Hour
	case "m":
		result.Period = time.Minute
	case "s":
		result.Period = time.Second
	default:
		return result, fmt.Errorf("%q is not a valid rate limit expression", value)
	}

	return result, nil
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

package env

import (
	"fmt"
	"os"
	"strconv"
)

const (
	notFoundFormat = "Environment variable %q not found"
)

func Get(name string) (string, error) {
	if value, ok := os.LookupEnv(name); ok {
		if value == "" {
			return "", fmt.Errorf("Empty value for environment variable: %s", name)
		}
		return value, nil
	}

	return "", fmt.Errorf(notFoundFormat, name)
}

func GetInt(name string) (int, error) {
	if value, ok := os.LookupEnv(name); ok {
		if r, err := strconv.Atoi(value); err == nil {
			return r, nil
		} else {
			return 0, fmt.Errorf("Environement variable %q is not an integer", name)
		}
	}

	return 0, fmt.Errorf(notFoundFormat, name)
}

func GetBool(name string) (bool, error) {
	if value, ok := os.LookupEnv(name); ok {
		return strconv.ParseBool(value)
	}

	return false, fmt.Errorf(notFoundFormat, name)
}

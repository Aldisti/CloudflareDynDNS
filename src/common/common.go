package common

import (
	"fmt"
	"strconv"
	"strings"
)

func IsBlank(s string) bool {
	return strings.TrimSpace(s) == ""
}

func IsNotBlank(s string) bool {
	return !IsBlank(s)
}

func Filter[T any](slice []T, predicate func(T)bool) []T {
	filtered := make([]T, 0)
	for _, e := range slice {
		if predicate(e) {
			filtered = append(filtered, e)
		}
	}
	return filtered
}

// getIntUnsafe parses a string to integer and panics if parsing fails.
//
// Parameters:
//   - value: the string to parse into an integer
//   - name: used only in the error message for context
func GetIntUnsafe(value, name string) int {
	if n, err := strconv.Atoi(value); err != nil {
		panic(fmt.Errorf("Value of %s must be an integer: %v", name, err))
	} else {
		return n
	}
}

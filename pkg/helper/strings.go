package helper

import (
	"strconv"
)

func StringInSlice(s string, slice []string) bool {
	for _, v := range slice {
		if s == v {
			return true
		}
	}
	return false
}

// CheckIfStructValueISEmpty returns value as a string, or "N/A" if empty.
func CheckIfStructValueISEmpty(value interface{}) string {
	if value == nil {
		return "N/A"
	}

	switch v := value.(type) {
	case int:
		return strconv.Itoa(v)

	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)

	case string:
		if v != "" && v != " " {
			return v
		}
		return "N/A"

	default:
		return "N/A"
	}
}

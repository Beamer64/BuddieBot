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

// CheckIfStructValueISEmpty checks if the value is empty and returns it as
// string if not. Otherwise returns "N/A".
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

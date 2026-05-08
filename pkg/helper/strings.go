package helper

import (
	"encoding/json"
	"io"
	"os"
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

func ToConvertedText(text string, convertGroup string) (string, error) {
	letters, err := getLetters()
	if err != nil {
		return "", err
	}

	convertedText := ""
	for _, char := range text {
		randSubs := ""
		subSet := letters[convertGroup][0][string(char)]
		if subSet != nil {
			randSubs = GetRandomStringFromSet(subSet)
		} else {
			randSubs = string(char)
		}
		convertedText += randSubs
	}

	return convertedText, nil
}

func getLetters() (map[string][]map[string][]string, error) {
	fontsDir := "datasets/text_fonts.json"
	if IsLaunchedByDebugger() {
		fontsDir = "../../datasets/text_fonts.json"
	}

	jsonFile, err := os.Open(fontsDir)
	if err != nil {
		return nil, err
	}

	defer func(jsonFile *os.File) {
		_ = jsonFile.Close()
	}(jsonFile)

	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}

	letters := make(map[string][]map[string][]string)
	err = json.Unmarshal(byteValue, &letters)
	if err != nil {
		return nil, err
	}

	return letters, nil
}

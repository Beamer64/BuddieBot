package helper

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestToLeetCode(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("skipping due to INTEGRATION env var not being set to 'true'")
	}

	letters := make(map[string][]map[string][]string)

	jsonFile, err := os.Open("../../config_files/text_fonts.json")
	if err != nil {
		t.Fatal(err)
	}

	defer func(jsonFile *os.File) {
		err = jsonFile.Close()
	}(jsonFile)
	if err != nil {
		t.Fatal(err)
	}

	byteValue, _ := ioutil.ReadAll(jsonFile)
	err = json.Unmarshal(byteValue, &letters)
	if err != nil {
		t.Fatal(err)
	}

	test := "This is a test."

	leet := ""
	randLeet := ""
	for _, char := range strings.ToLower(test) {
		subs := letters["1337"][0][string(char)]
		if subs != nil {
			randLeet = GetRandomStringFromSet(subs)
		} else {
			randLeet = string(char)
		}
		leet += randLeet
	}

	fmt.Println(leet)
}

func TestToBubbleCode(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("skipping due to INTEGRATION env var not being set to 'true'")
	}

	letters := make(map[string][]map[string][]string)

	jsonFile, err := os.Open("../../config_files/text_fonts.json")
	if err != nil {
		t.Fatal(err)
	}

	defer func(jsonFile *os.File) {
		err = jsonFile.Close()
	}(jsonFile)
	if err != nil {
		t.Fatal(err)
	}

	byteValue, _ := ioutil.ReadAll(jsonFile)
	err = json.Unmarshal(byteValue, &letters)
	if err != nil {
		t.Fatal(err)
	}

	test := "This is a test. baby"

	bubble := ""
	randBubble := ""
	for _, char := range strings.ToLower(test) {
		subs := letters["bubble"][0][string(char)]
		if subs != nil {
			randBubble = GetRandomStringFromSet(subs)
		} else {
			randBubble = string(char)
		}
		bubble += randBubble
	}

	fmt.Println(bubble)
}

func TestCheckIfStructValueISEmpty(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("skipping due to INTEGRATION env var not being set to 'true'")
	}

	var floatVal interface{} = 12.5
	var stringVal interface{} = "Test String"
	var intVal interface{} = 7
	var nilVal interface{}

	var x []interface{}

	x = append(x, floatVal)
	x = append(x, stringVal)
	x = append(x, intVal)
	x = append(x, nilVal)

	for _, face := range x {
		retVal := ""

		if face != nil {
			switch face.(type) {
			case int:
				retVal = fmt.Sprintf("%v", reflect.ValueOf(face).Int())

			case float64:
				retVal = fmt.Sprintf("%v", reflect.ValueOf(face).Float())

			case string:
				if face != "" && face != " " {
					retVal = reflect.ValueOf(face).String()
				}

			default:
				retVal = "N/A"
			}
		} else {
			retVal = "N/A"
		}

		fmt.Println(fmt.Sprintf("Type: %v Value: %s", reflect.TypeOf(face), retVal))
	}
}

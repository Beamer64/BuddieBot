package helper

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestToLeetCode(t *testing.T) {
	/*if os.Getenv("INTEGRATION") != "true" {
		t.Skip("skipping due to INTEGRATION env var not being set to 'true'")
	}*/

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
	for _, v := range strings.ToLower(test) {
		subs := letters["leet"][0][string(v)]
		randLeet := GetRandomStringFromSet(subs)
		leet += randLeet
	}

	fmt.Println(leet)
}

func TestToBubbleCode(t *testing.T) {
	/*if os.Getenv("INTEGRATION") != "true" {
		t.Skip("skipping due to INTEGRATION env var not being set to 'true'")
	}*/

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
	for _, v := range strings.ToLower(test) {
		subs := letters["bubble"][0][string(v)]
		randBubble := GetRandomStringFromSet(subs)
		bubble += randBubble
	}

	fmt.Println(bubble)
}

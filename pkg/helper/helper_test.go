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
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("skipping due to INTEGRATION env var not being set to 'true'")
	}

	codes := make(map[string][]string)

	jsonFile, err := os.Open("../helper/leetCodes.json")
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

	err = json.Unmarshal(byteValue, &codes)
	if err != nil {
		t.Fatal(err)
	}

	test := "This is a test."

	leet := ""
	for _, v := range strings.ToLower(test) {
		subs := codes[string(v)]
		randLeet := getRandLeet(subs)
		leet += randLeet
	}

	fmt.Println(leet)
}

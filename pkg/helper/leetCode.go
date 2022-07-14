package helper

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
)

func ToLeetCode(text string) (string, error) {
	codes, err := getCodes()
	if err != nil {
		return "", err
	}

	leet := ""
	for _, char := range strings.ToLower(text) {
		subs := codes[string(char)]
		randLeet := getRandLeet(subs)
		leet += randLeet
	}

	return leet, nil

}

func getCodes() (map[string][]string, error) {
	codes := make(map[string][]string)

	jsonFile, err := os.Open("config_files/leetCodes.json")
	if err != nil {
		return nil, err
	}

	defer func(jsonFile *os.File) {
		err = jsonFile.Close()
	}(jsonFile)
	if err != nil {
		return nil, err
	}

	byteValue, _ := ioutil.ReadAll(jsonFile)

	err = json.Unmarshal(byteValue, &codes)
	if err != nil {
		return nil, err
	}

	return codes, nil
}

func getRandLeet(subs []string) string {
	randLeet := ""
	if subs != nil {
		randNum := rand.Intn(len(subs))
		randLeet = subs[randNum]
	}
	return randLeet
}

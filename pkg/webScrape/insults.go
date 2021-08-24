package webScrape

import (
	"encoding/json"
	"io"
	"net/http"
)

type insult struct {
	Insult string `json:"insult"`
}

func GetInsult(insultURL string) (string, error) {
	res, err := http.Get(insultURL)
	if err != nil {
		return "", err
	}

	var insultObj insult

	err = json.NewDecoder(res.Body).Decode(&insultObj)
	if err != nil {
		return "", err
	}

	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			return
		}
	}(res.Body)

	return insultObj.Insult, nil
}

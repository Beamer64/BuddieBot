package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
)

type gif struct {
	Results []struct {
		Media []struct {
			Gif struct {
				Size    int    `json:"size"`
				Preview string `json:"preview"`
				Dims    []int  `json:"dims"`
				URL     string `json:"url"`
			} `json:"gif"`
		} `json:"media"`
	} `json:"results"`
}

func RequestGifURL(searchStr, tenorAPIkey string) (string, error) {
	var gifObj gif

	searchStr = url.QueryEscape(searchStr)
	URL := "https://g.tenor.com/v1/search?q=" + searchStr + "&key=" + tenorAPIkey + "&limit=1"

	res, err := http.Get(URL)
	if err != nil {
		return "", err
	}

	err = json.NewDecoder(res.Body).Decode(&gifObj)
	if err != nil {
		return "", err
	}

	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			return
		}
	}(res.Body)

	gifURL := gifObj.Results[0].Media[0].Gif.URL

	return gifURL, nil
}

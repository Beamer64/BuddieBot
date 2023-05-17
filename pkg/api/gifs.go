package api

import (
	"encoding/json"
	"fmt"
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
	searchStr = url.QueryEscape(searchStr)
	URL := "https://g.tenor.com/v1/search?q=" + searchStr + "&key=" + tenorAPIkey + "&limit=1"

	resp, err := http.Get(URL)
	if err != nil {
		return "", fmt.Errorf("failed to request gif URL: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API call failed with status code %d", resp.StatusCode)
	}

	var gifObj gif
	err = json.NewDecoder(resp.Body).Decode(&gifObj)
	if err != nil {
		return "", fmt.Errorf("failed to decode response body: %v", err)
	}

	if len(gifObj.Results) == 0 || len(gifObj.Results[0].Media) == 0 {
		return "", fmt.Errorf("no gif found for the search string %s", searchStr)
	}

	gifURL := gifObj.Results[0].Media[0].Gif.URL

	return gifURL, nil
}

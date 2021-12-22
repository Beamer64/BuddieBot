package webScrape

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Jeffail/gabs"
	"github.com/parnurzeal/gorequest"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type stream map[string]string
type youtube struct {
	streamList []stream
	videoID    string
	videoInfo  *http.Response
}

// Client offers methods to download video metadata and video streams.
type Client struct {
	// Debug enables debugging output through log package
	Debug bool

	// HTTPClient can be used to set a custom HTTP client.
	// If not set, http.DefaultClient will be used
	HTTPClient *http.Client

	// playerCache caches the JavaScript code of a player response
	playerCache playerCache
}

type playerCache struct {
	key       string
	expiredAt time.Time
	config    playerConfig
}

type playerConfig []byte

// GetYoutubeURL converts a standard Youtube URL or ID to an mp4 download link,
// or searches Youtube and picks the first result.
func GetYoutubeURL(query, apiKey string) (string, string, error) {
	y := new(youtube)

	if len(query) < 4 || query[:4] != "http" {
		link, err := searchYoutube(query, apiKey)
		if err != nil {
			return "", "", err
		}
		query = link
	}

	err := y.findVideoID(query)
	if err != nil {
		return "", "", fmt.Errorf("findVideoID error=%s", err)
	}

	err = y.getVideoInfo(apiKey)
	if err != nil {
		return "", "", fmt.Errorf("getVideoInfo error=%s", err)
	}

	err = y.parseVideoInfo()
	if err != nil {
		return "", "", fmt.Errorf("parse video info failed, err=%s", err)
	}

	targetStream := y.streamList[0]
	downloadURL := targetStream["url"] + "&signature=" + targetStream["sig"]
	return downloadURL, targetStream["title"], nil
}

func searchYoutube(text, apiKey string) (string, error) {
	formattedURL := fmt.Sprintf("https://www.googleapis.com/youtube/v3/search?part=snippet&q=%s&key=%s", url.QueryEscape(text), apiKey)

	_, body, err := gorequest.New().Get(formattedURL).EndBytes()
	if len(err) > 0 { // ??? never seen an []err before
		return "", err[0]
	}

	jsonParsed, _ := gabs.ParseJSON(body)
	children, _ := jsonParsed.S("items").Children()
	if len(children) == 0 {
		return "", fmt.Errorf("No video found")
	}

	videoID, _ := children[0].Path("id.videoId").Data().(string)
	return videoID, nil
}

func (y *youtube) findVideoID(videoID string) error {
	if strings.Contains(videoID, "youtu") || strings.ContainsAny(videoID, "\"?&/<%=") {
		reList := []*regexp.Regexp{
			regexp.MustCompile(`(?:v|embed|watch\?v)(?:=|/)([^"&?/=%]{11})`),
			regexp.MustCompile(`(?:=|/)([^"&?/=%]{11})`),
			regexp.MustCompile(`([^"&?/=%]{11})`),
		}
		for _, re := range reList {
			if isMatch := re.MatchString(videoID); isMatch {
				subs := re.FindStringSubmatch(videoID)
				videoID = subs[1]
				break
			}
		}
	}

	y.videoID = videoID
	if strings.ContainsAny(videoID, "?&/<%=") {
		return errors.New("invalid characters in video id")
	}
	if len(videoID) < 10 {
		return errors.New("the video id must be at least 10 characters long")
	}
	return nil
}

func (y *youtube) getVideoInfo(apiKey string) error {
	URL := "https://www.googleapis.com/youtube/v3/videos?id=" + y.videoID + "&key=" + apiKey + "&part=snippet,status,player"

	resp, err := http.Get(URL)
	if err != nil {
		return err
	}
	y.videoInfo = resp
	return nil
}

func (y *youtube) parseVideoInfo() error {
	var responseResults map[string]interface{}
	videoTitle := ""
	videoURL := ""

	// answer, err := url.ParseQuery(y.videoInfo)
	err := json.NewDecoder(y.videoInfo.Body).Decode(&responseResults)
	if err != nil {
		return err
	}

	for key, responseValue := range responseResults {
		if key == "items" {
			for key, itemsValue := range responseValue.([]interface{})[0].(map[string]interface{}) {
				if key == "status" {
					for key, statusValue := range itemsValue.(map[string]interface{}) {
						if key == "uploadStatus" {
							if statusValue != "processed" {
								err = fmt.Errorf("no response status found in the server's answer")
								return err
							}
							break
						}
					}
				}
			}
		}
	}
	for key, responseValue := range responseResults {
		if key == "items" {
			for key, itemsValue := range responseValue.([]interface{})[0].(map[string]interface{}) {
				if key == "player" {
					for key, playerValue := range itemsValue.(map[string]interface{}) {
						if key == "embedHtml" {
							rawURLValue := fmt.Sprintf("%v", playerValue)
							rawURLString := strings.SplitAfterN(rawURLValue, "\"", 7)[5]
							// URL := strings.Replace(rawURLString, "/", "", -1)
							// URL = strings.Replace(URL, "\"", "", -1)
							URL := strings.Trim(rawURLString, "/")
							videoURL = URL
						}
						break
					}
					break
				}
			}
		}
	}

	stream := stream{
		"quality": "",
		"type":    "",
		"url":     videoURL,
		"sig":     "",
		"title":   videoTitle,
		"author":  "",
	}

	y.streamList = append(y.streamList, stream)
	return nil
}

func (y *youtube) downloadAudio(apiKey string) error {
	return nil
}

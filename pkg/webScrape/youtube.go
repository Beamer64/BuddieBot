package webScrape

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strings"

	"github.com/Jeffail/gabs"
	"github.com/parnurzeal/gorequest"
)

type stream map[string]string
type youtube struct {
	streamList []stream
	videoID    string
	videoInfo  string
}

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

	err = y.getVideoInfo()
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

func (y *youtube) getVideoInfo() error {
	url := "http://youtube.com/get_video_info?video_id=" + y.videoID
	_, body, err := gorequest.New().Get(url).End()
	if err != nil {
		return err[0]
	}
	y.videoInfo = body
	return nil
}

func (y *youtube) parseVideoInfo() error {
	answer, err := url.ParseQuery(y.videoInfo)
	if err != nil {
		return err
	}

	status, ok := answer["status"]
	if !ok {
		err = fmt.Errorf("no response status found in the server's answer")
		return err
	}
	if status[0] == "fail" {
		reason, ok := answer["reason"]
		if ok {
			err = fmt.Errorf("'fail' response status found in the server's answer, reason: '%s'", reason[0])
		} else {
			err = errors.New(fmt.Sprint("'fail' response status found in the server's answer, no reason given"))
		}
		return err
	}
	if status[0] != "ok" {
		err = fmt.Errorf("non-success response status found in the server's answer (status: '%s')", status)
		return err
	}

	// read the streams map
	streamMap, ok := answer["url_encoded_fmt_stream_map"]
	if !ok {
		err = errors.New(fmt.Sprint("no stream map found in the server's answer"))
		return err
	}

	// read each stream
	for streamPos, streamRaw := range strings.Split(streamMap[0], ",") {
		streamQry, err := url.ParseQuery(streamRaw)
		if err != nil {
			log.Println(fmt.Sprintf("An error occured while decoding one of the video's stream's information: stream %d: %s\n", streamPos, err))
			continue
		}

		stream := stream{
			"quality": streamQry["quality"][0],
			"type":    streamQry["type"][0],
			"url":     streamQry["url"][0],
			"sig":     "",
			"title":   answer["title"][0],
			"author":  answer["author"][0],
		}
		if _, exist := streamQry["sig"]; exist {
			stream["sig"] = streamQry["sig"][0]
		}

		y.streamList = append(y.streamList, stream)
	}
	return nil
}

package commands

var ones = map[string]struct {
	x1 int
	y1 int
	x2 int
	y2 int
}{
	"1": {100, 20, 140, 20},
	"2": {100, 60, 140, 60},
	"3": {100, 20, 140, 60},
	"4": {100, 60, 140, 20},
	"5": {100, 20, 140, 20},
	"6": {140, 20, 140, 60},
	"7": {100, 20, 140, 20},
	"8": {100, 60, 140, 60},
	"9": {100, 20, 140, 20},
}

var tens = map[string]struct {
	x1 int
	y1 int
	x2 int
	y2 int
}{
	"1": {100, 20, 60, 20},
	"2": {100, 60, 60, 60},
	"3": {100, 20, 60, 60},
	"4": {100, 60, 60, 20},
	"5": {100, 20, 60, 20},
	"6": {60, 20, 60, 60},
	"7": {100, 20, 60, 20},
	"8": {100, 60, 60, 60},
	"9": {100, 20, 60, 20},
}
var hunds = map[string]struct {
	x1 int
	y1 int
	x2 int
	y2 int
}{
	"1": {100, 180, 140, 180},
	"2": {100, 140, 140, 140},
	"3": {100, 180, 140, 140},
	"4": {100, 140, 140, 180},
	"5": {100, 180, 140, 180},
	"6": {140, 180, 140, 140},
	"7": {100, 180, 140, 180},
	"8": {100, 140, 140, 140},
	"9": {100, 180, 140, 180},
}
var thous = map[string]struct {
	x1 int
	y1 int
	x2 int
	y2 int
}{
	"1": {100, 180, 60, 180},
	"2": {100, 140, 60, 140},
	"3": {100, 180, 60, 140},
	"4": {100, 140, 60, 180},
	"5": {100, 180, 60, 180},
	"6": {60, 180, 60, 140},
	"7": {100, 180, 60, 180},
	"8": {100, 140, 60, 140},
	"9": {100, 180, 60, 180},
}

type imgBBData struct {
	Data    imgData `json:"data"`
	Success bool    `json:"success"`
	Status  int     `json:"status"`
}
type imageInfo struct {
	Filename  string `json:"filename"`
	Name      string `json:"name"`
	Mime      string `json:"mime"`
	Extension string `json:"extension"`
	URL       string `json:"url"`
}
type thumb struct {
	Filename  string `json:"filename"`
	Name      string `json:"name"`
	Mime      string `json:"mime"`
	Extension string `json:"extension"`
	URL       string `json:"url"`
}
type medium struct {
	Filename  string `json:"filename"`
	Name      string `json:"name"`
	Mime      string `json:"mime"`
	Extension string `json:"extension"`
	URL       string `json:"url"`
}
type imgData struct {
	ID         string    `json:"id"`
	Title      string    `json:"title"`
	URLViewer  string    `json:"url_viewer"`
	URL        string    `json:"url"`
	DisplayURL string    `json:"display_url"`
	Width      int       `json:"width"`
	Height     int       `json:"height"`
	Size       int       `json:"size"`
	Time       int       `json:"time"`
	Expiration int       `json:"expiration"`
	Image      imageInfo `json:"image"`
	Thumb      thumb     `json:"thumb"`
	Medium     medium    `json:"medium"`
	DeleteURL  string    `json:"delete_url"`
}

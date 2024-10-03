package commands

var ones = map[string]struct {
	x1 int
	y1 int
	x2 int
	y2 int
}{
	"1": {50, 10, 70, 10},
	"2": {50, 30, 70, 30},
	"3": {50, 10, 70, 30},
	"4": {50, 30, 70, 10},
	"5": {50, 10, 70, 10},
	"6": {70, 10, 70, 30},
	"7": {50, 10, 70, 10},
	"8": {50, 30, 70, 30},
	"9": {50, 10, 70, 10},
}

var tens = map[string]struct {
	x1 int
	y1 int
	x2 int
	y2 int
}{
	"1": {50, 10, 30, 10},
	"2": {50, 30, 30, 30},
	"3": {50, 10, 30, 30},
	"4": {50, 30, 30, 10},
	"5": {50, 10, 30, 10},
	"6": {30, 10, 30, 30},
	"7": {50, 10, 30, 10},
	"8": {50, 30, 30, 30},
	"9": {50, 10, 30, 10},
}
var hunds = map[string]struct {
	x1 int
	y1 int
	x2 int
	y2 int
}{
	"1": {50, 90, 70, 90},
	"2": {50, 70, 70, 70},
	"3": {50, 90, 70, 70},
	"4": {50, 70, 70, 90},
	"5": {50, 90, 70, 90},
	"6": {70, 90, 70, 70},
	"7": {50, 90, 70, 90},
	"8": {50, 70, 70, 70},
	"9": {50, 90, 70, 90},
}
var thous = map[string]struct {
	x1 int
	y1 int
	x2 int
	y2 int
}{
	"1": {50, 90, 30, 90},
	"2": {50, 70, 30, 70},
	"3": {50, 90, 30, 70},
	"4": {50, 70, 30, 90},
	"5": {50, 90, 30, 90},
	"6": {30, 90, 30, 70},
	"7": {50, 90, 30, 90},
	"8": {50, 70, 30, 70},
	"9": {50, 90, 30, 90},
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

package helper

var ApprovalWords = []string{
	"enabled",
	"on",
	"true",
	"yes",
	"sure",
}

var DisapprovalWords = []string{
	"disabled",
	"off",
	"false",
	"no",
	"nope",
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

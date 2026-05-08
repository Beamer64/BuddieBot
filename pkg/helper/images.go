package helper

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"

	"github.com/Beamer64/BuddieBot/pkg/config"
)

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

func CreateImgFile(imgPath string, img image.Image) error {
	// save imageData
	toimg, err := os.Create(imgPath)
	if err != nil {
		return err
	}
	log.Println("Created image location...")

	defer func(toimg *os.File) {
		err = toimg.Close()
	}(toimg)
	if err != nil {
		return err
	}

	err = png.Encode(toimg, img)
	if err != nil {
		return err
	}
	log.Println("Image Encoded...")

	return nil
}

func GetImgbbUploadURL(cfg *config.Configs, imgPath string, expireSecs ...int) (string, error) {
	apiUrl := fmt.Sprintf("%s&key=%s", cfg.Configs.ApiURLs.ImgbbAPI, cfg.Configs.Keys.ImgbbAPIkey)
	if expireSecs != nil {
		apiUrl = fmt.Sprintf("%sexpiration=%s&key=%s", cfg.Configs.ApiURLs.ImgbbAPI, strconv.Itoa(expireSecs[0]), cfg.Configs.Keys.ImgbbAPIkey)
	}

	// Read the entire file into a byte slice
	imgBytes, err := os.ReadFile(imgPath)
	if err != nil {
		return "", err
	}

	// Create a buffer to write out multipart form data to
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add the imageData field
	base64img := base64.StdEncoding.EncodeToString(imgBytes)
	err = writer.WriteField("image", base64img)
	if err != nil {
		return "", fmt.Errorf("failed to write field: %w", err)
	}

	// Close the writer to finalize the form data
	err = writer.Close()
	if err != nil {
		return "", fmt.Errorf("failed to close writer: %w", err)
	}

	req, err := http.NewRequest("POST", apiUrl, body)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var imgbbDataObject imgBBData
	err = json.Unmarshal(respBody, &imgbbDataObject)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return imgbbDataObject.Data.URL, nil
}

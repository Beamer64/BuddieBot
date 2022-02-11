package commands

import (
	"bytes"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var ResponseTimer chan bool

type steamGames struct {
	Applist struct {
		Apps []struct {
			Appid int    `json:"appid"`
			Name  string `json:"name"`
		} `json:"apps"`
	} `json:"applist"`
}

type affirmation struct {
	Affirmation string `json:"affirmation"`
}

type kanye struct {
	Quote string `json:"quote"`
}

type advice struct {
	Slip struct {
		ID     int    `json:"id"`
		Advice string `json:"advice"`
	} `json:"slip"`
}

type doggo []struct {
	Breeds []struct {
		Weight struct {
			Imperial string `json:"imperial"`
			Metric   string `json:"metric"`
		} `json:"weight"`
		Height struct {
			Imperial string `json:"imperial"`
			Metric   string `json:"metric"`
		} `json:"height"`
		ID               int    `json:"id"`
		Name             string `json:"name"`
		BredFor          string `json:"bred_for"`
		BreedGroup       string `json:"breed_group"`
		LifeSpan         string `json:"life_span"`
		Temperament      string `json:"temperament"`
		Origin           string `json:"origin"`
		ReferenceImageID string `json:"reference_image_id"`
	} `json:"breeds"`
	ID     string `json:"id"`
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type insult struct {
	Insult string `json:"insult"`
}

type joke struct {
	ID   string `json:"id"`
	Joke string `json:"joke"`
}

type pickupLine struct {
	Category string `json:"category"`
	Joke     string `json:"joke"`
}

type fact struct {
	Fact string `json:"fact"`
}

type wtp struct {
	Data struct {
		Type      []string `json:"Type"`
		Abilities []string `json:"abilities"`
		ASCII     string   `json:"ascii"`
		Height    float64  `json:"height"`
		ID        int      `json:"id"`
		Link      string   `json:"link"`
		Name      string   `json:"name"`
		Weight    int      `json:"weight"`
	} `json:"Data"`
	Answer   string `json:"answer"`
	Question string `json:"question"`
}

func getErrorEmbed(err error) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Title:       "ERROR",
		Description: "(ノಠ益ಠ)ノ彡┻━┻",
		Color:       16726843,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Stack",
				Value:  fmt.Sprintf("%+v", errors.WithStack(err)),
				Inline: true,
			},
		},
	}

	return embed
}

// Returns pseudo rand num between low and high.
// For random embed color: rangeIn(1, 16777215)
func rangeIn(low, hi int) int {
	return low + rand.Intn(hi-low)
}

// Checks if the value is empty and returns it if not.
// Otherwise, return 'N/A'
func checkIfEmpty(value string) string {
	if value != "" {
		return value
	}
	return "N/A"
}

// ImageCleanUp Clear out Image directory
func ImageCleanUp(dir string) error {
	fmt.Println("Running Cleanup")

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, f := range files {
		if strings.Contains(filepath.Join(dir, filepath.Base(f.Name())), ".png") {
			err = os.Remove(filepath.Join(dir, filepath.Base(f.Name())))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// saveResponseImage saves []byte data to .png file
func saveResponseImage(byte []byte, guildID string) (string, error) {
	// setting image name to written time
	wrTime := time.Now().Format("2006-01-02 15:04:05")
	wrTimeName := strings.Replace(wrTime, " ", "-", -1)
	wrTimeName = strings.Replace(wrTimeName, ":", "-", -1)

	// Create the dir
	dir := fmt.Sprintf("%s/Image", guildID)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		// does not exist
		err = os.MkdirAll(dir, 0777)
		if err != nil {
			return "", err
		}

		fmt.Println(fmt.Sprintf("Dir created: %s", dir))
	}

	imagePath := filepath.Join(dir, filepath.Base(wrTimeName+".png"))

	//open a file for writing
	file, err := os.Create(imagePath)
	if err != nil {
		return "", err
	}

	defer func(file *os.File) {
		err = file.Close()
	}(file)
	if err != nil {
		return "", err
	}

	// Use io.Copy to just dump the response body to the file.
	_, err = io.Copy(file, bytes.NewReader(byte))
	if err != nil {
		return "", err
	}

	return imagePath, nil
}

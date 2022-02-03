package commands

import (
	"math/rand"
)

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

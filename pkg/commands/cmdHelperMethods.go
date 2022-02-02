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

// rangeIn Returns pseudo rand num between low and high.
// For random embed color: rangeIn(1, 16777215)
func rangeIn(low, hi int) int {
	return low + rand.Intn(hi-low)
}

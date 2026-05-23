package helper

import (
	"math/rand"
)

// RangeIn returns a pseudo-random int in [low, hi).
func RangeIn(low, hi int) int {
	return low + rand.Intn(hi-low)
}

// DiscordMaxColor is the maximum 24-bit color value Discord embeds accept.
const DiscordMaxColor = 0xFFFFFF

// RandomDiscordColor returns a random non-black embed color.
func RandomDiscordColor() int {
	return RangeIn(1, DiscordMaxColor)
}

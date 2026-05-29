package helper

import "math/rand"

// RatingNames is the canonical list of /rate-this rating types. The slash-command
// spec choices and database seeding both source from here so they can't drift.
// Schmeat is the odd one out: it has no numeric /100 score, just an ASCII
// "C===…=8" of size 1–14 (see RangeIn semantics).
var RatingNames = []string{
	"dank",
	"epic-gamer",
	"gay",
	"geek",
	"looks",
	"neckbeard",
	"nerd",
	"npc",
	"pickme",
	"schmeat",
	"simp",
	"smarts",
	"stinky",
	"thot",
}

// RandomRatingValue picks a fresh value in the right range for the given
// rating type. Standard ratings are 0–99 (matches the live /rate-this output of
// "X/100"); schmeat is 1–14 (matches RangeIn(1, 15)'s half-open semantics so
// the ASCII strip is at least one '=' wide).
func RandomRatingValue(name string) int {
	if name == "schmeat" {
		return RangeIn(1, 8)
	}
	return rand.Intn(100)
}

// SchmeatString renders a stored schmeat size as the ASCII strip the live
// command produces — e.g. size 5 → "C=====8".
func SchmeatString(size int) string {
	if size < 1 {
		size = 1
	}
	bar := make([]byte, size+2)
	bar[0] = 'C'
	for i := 1; i <= size; i++ {
		bar[i] = '='
	}
	bar[size+1] = '8'
	return string(bar)
}

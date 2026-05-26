package helper

const (
	// PollMessageContent — the reaction handler matches messages by this body.
	PollMessageContent = "Poll Time!"

	// LmgtfyEmojiName triggers the "Let Me Google That For You" reaction handler.
	LmgtfyEmojiName = "grey_question"

	// ErrorReaction — fallback when LogAndReact can't DM the user.
	ErrorReaction = "⚠️"

	CistercianMin = -9999
	CistercianMax = 9999
)

var CistOnes = map[string]struct {
	X1 int
	Y1 int
	X2 int
	Y2 int
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

var CistTens = map[string]struct {
	X1 int
	Y1 int
	X2 int
	Y2 int
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
var CistHunds = map[string]struct {
	X1 int
	Y1 int
	X2 int
	Y2 int
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
var CistThous = map[string]struct {
	X1 int
	Y1 int
	X2 int
	Y2 int
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

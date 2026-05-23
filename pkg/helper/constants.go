package helper

const (
	// PollMessageContent is the message body posted by /pick poll. The reaction
	// handler matches on this string to manage poll reactions.
	PollMessageContent = "Poll Time!"

	// LmgtfyEmojiName is the Discord emoji name (without colons) that triggers
	// the "Let Me Google That For You" reaction handler.
	LmgtfyEmojiName = "grey_question"

	// ErrorReaction used helper.LogAndReact to react to messages that caused an error when the bot tries to DM the user about the error.
	// This is a fallback to at least notify the user that their command failed.
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

package slash

// Public control-flow values shared between slash command handlers and event handlers.
// Keeping them in one spot prevents the message body and the reaction-handler matcher
// from drifting apart silently.

const (
	// PollMessageContent is the message body posted by /pick poll. The reaction
	// handler matches on this string to manage poll reactions.
	PollMessageContent = "Poll Time!"

	// LmgtfyEmojiName is the Discord emoji name (without colons) that triggers
	// the "Let Me Google That For You" reaction handler.
	LmgtfyEmojiName = "grey_question"
)

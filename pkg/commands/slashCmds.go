package commands

// ApplicationCommandType represents the type of application command.
type ApplicationCommandType uint8

// ApplicationCommandOptionType indicates the type of a slash command's option.
type ApplicationCommandOptionType uint8

type ApplicationCommand struct {
	ID            string                 `json:"id,omitempty"`
	Type          ApplicationCommandType `json:"type,omitempty"`
	ApplicationID string                 `json:"application_id,omitempty"`
	Name          string                 `json:"name"`
	// NOTE: Chat commands only. Otherwise it mustn't be set.
	Description string `json:"description,omitempty"`
	Version     string `json:"version,omitempty"`
	// NOTE: Chat commands only. Otherwise it mustn't be set.
	Options []*ApplicationCommandOption `json:"options"`
}

type ApplicationCommandOption struct {
	Type        ApplicationCommandOptionType `json:"type"`
	Name        string                       `json:"name"`
	Description string                       `json:"description,omitempty"`
	// NOTE: This feature was on the API, but at some point developers decided to remove it.
	// So I commented it, until it will be officially on the docs.
	// Default     bool                        Default`json:"default"`
	Required bool                              `json:"required"`
	Choices  []*ApplicationCommandOptionChoice `json:"choices"`
	Options  []*ApplicationCommandOption       `json:"options"`
}

// ApplicationCommandOptionChoice represents a slash command option choice.
type ApplicationCommandOptionChoice struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

// Application command option types.
const (
	ApplicationCommandOptionSubCommand      ApplicationCommandOptionType = 1
	ApplicationCommandOptionSubCommandGroup ApplicationCommandOptionType = 2
	ApplicationCommandOptionString          ApplicationCommandOptionType = 3
	ApplicationCommandOptionInteger         ApplicationCommandOptionType = 4
	ApplicationCommandOptionBoolean         ApplicationCommandOptionType = 5
	ApplicationCommandOptionUser            ApplicationCommandOptionType = 6
	ApplicationCommandOptionChannel         ApplicationCommandOptionType = 7
	ApplicationCommandOptionRole            ApplicationCommandOptionType = 8
	ApplicationCommandOptionMentionable     ApplicationCommandOptionType = 9
	ApplicationCommandOptionNumber          ApplicationCommandOptionType = 10
)

// Application Command Types
const (
	CHAT_INPUT ApplicationCommandType = 1
	USER       ApplicationCommandType = 2
	MESSAGE    ApplicationCommandType = 3
)

var (
	commands = []*ApplicationCommand{
		{
			Name:        "Flip-Coin",
			Type:        CHAT_INPUT,
			Description: "Flips a coin. Pretty simple.",
		},

		/*{
				Name: "basic-command",
				// All commands and options must have a description
				// Commands/options without description will fail the registration
				// of the command.
				Description: "Basic command",
		},
		{
			Name:        "basic-command-with-files",
			Description: "Basic command with files",
		},
		{
			Name:        "options",
			Description: "Command for demonstrating options",
			Options: []*ApplicationCommandOption{

				{
					Type:        ApplicationCommandOptionString,
					Name:        "string-option",
					Description: "String option",
					Required:    true,
				},
				{
					Type:        ApplicationCommandOptionInteger,
					Name:        "integer-option",
					Description: "Integer option",
					Required:    true,
				},
				{
					Type:        ApplicationCommandOptionBoolean,
					Name:        "bool-option",
					Description: "Boolean option",
					Required:    true,
				},

				// Required options must be listed first since optional parameters
				// always come after when they're used.
				// The same concept applies to Discord's Slash-commands API

				{
					Type:        ApplicationCommandOptionChannel,
					Name:        "channel-option",
					Description: "Channel option",
					Required:    false,
				},
				{
					Type:        ApplicationCommandOptionUser,
					Name:        "user-option",
					Description: "User option",
					Required:    false,
				},
				{
					Type:        ApplicationCommandOptionRole,
					Name:        "role-option",
					Description: "Role option",
					Required:    false,
				},
			},
		},
		{
			Name:        "subcommands",
			Description: "Subcommands and command groups example",
			Options: []*ApplicationCommandOption{
				// When a command has subcommands/subcommand groups
				// It must not have top-level options, they aren't accesible in the UI
				// in this case (at least not yet), so if a command has
				// subcommands/subcommand any groups registering top-level options
				// will cause the registration of the command to fail

				{
					Name:        "scmd-grp",
					Description: "Subcommands group",
					Options: []*ApplicationCommandOption{
						// Also, subcommand groups aren't capable of
						// containing options, by the name of them, you can see
						// they can only contain subcommands
						{
							Name:        "nst-subcmd",
							Description: "Nested subcommand",
							Type:        ApplicationCommandOptionSubCommand,
						},
					},
					Type: ApplicationCommandOptionSubCommandGroup,
				},
				// Also, you can create both subcommand groups and subcommands
				// in the command at the same time. But, there's some limits to
				// nesting, count of subcommands (top level and nested) and options.
				// Read the intro of slash-commands docs on Discord dev portal
				// to get more information
				{
					Name:        "subcmd",
					Description: "Top-level subcommand",
					Type:        ApplicationCommandOptionSubCommand,
				},
			},
		},
		{
			Name:        "responses",
			Description: "Interaction responses testing initiative",
			Options: []*ApplicationCommandOption{
				{
					Name:        "resp-type",
					Description: "Response type",
					Type:        ApplicationCommandOptionInteger,
					Choices: []*ApplicationCommandOptionChoice{
						{
							Name:  "Channel message with source",
							Value: 4,
						},
						{
							Name:  "Deferred response With Source",
							Value: 5,
						},
					},
					Required: true,
				},
			},
		},
		{
			Name:        "followups",
			Description: "Followup messages",
		},*/

	}
)

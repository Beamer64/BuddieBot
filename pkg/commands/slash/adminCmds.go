package slash

// Skeleton for the future /admin command group (per-guild settings, future
// configuration toggles, etc.). Intentionally NOT registered in handlers.go
// or slashCmds.go until at least one real subcommand exists — registering
// a command with no subcommands would make `/admin` invokable as a no-op.
//
// When adding the first subcommand:
//   1. Add a SubCommand to adminSpec.Options
//   2. Add a case in sendAdminResponse's switch
//   3. Wire `"admin": wrap(sendAdminResponse)` into handlers.CommandHandlers
//   4. Add adminSpec() to slashCmds.Commands

import (
	"fmt"

	"github.com/Beamer64/BuddieBot/pkg/config"
	"github.com/Beamer64/BuddieBot/pkg/helper"
	"github.com/bwmarrin/discordgo"
)

func sendAdminResponse(s *discordgo.Session, i *discordgo.InteractionCreate, _ *config.Configs) error {
	if err := s.InteractionRespond(
		i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		},
	); err != nil {
		return fmt.Errorf("failed to defer interaction for /admin: %w", err)
	}

	sub := i.ApplicationCommandData().Options[0]
	return helper.ReturnUserErrorDeferred(s, i, "Unknown admin subcommand.", fmt.Errorf("unknown admin subcommand: %s", sub.Name))
}

func adminSpec() *discordgo.ApplicationCommand {
	// ManageServer gates the whole command at the Discord layer.
	perm := int64(discordgo.PermissionManageServer)
	return &discordgo.ApplicationCommand{
		Name:                     "admin",
		Description:              "Server-admin configuration commands",
		DefaultMemberPermissions: &perm,
		Options:                  []*discordgo.ApplicationCommandOption{},
	}
}

package slash

import (
	"errors"
	"strings"
	"testing"

	"github.com/Beamer64/BuddieBot/pkg/config"
	"github.com/bwmarrin/discordgo"
)

// minInteraction returns an InteractionCreate whose GuildID can be read
// without panicking.
func minInteraction() *discordgo.InteractionCreate {
	return &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{GuildID: "test-guild"}}
}

func TestWrap_ForwardsHandlerError(t *testing.T) {
	wantErr := errors.New("boom from handler")
	var loggedErr error
	var loggedGuildID string

	logErr := func(s *discordgo.Session, cfg *config.Configs, err error, guildID string) {
		loggedErr = err
		loggedGuildID = guildID
	}
	notifier := func(s *discordgo.Session, i *discordgo.InteractionCreate, msg string) error {
		t.Errorf("notifier should not be called when handler returns error (no panic): msg=%q", msg)
		return nil
	}

	h := func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) error {
		return wantErr
	}

	wrapped := wrapWithDeps(h, logErr, notifier)
	wrapped(nil, minInteraction(), &config.Configs{})

	if !errors.Is(loggedErr, wantErr) {
		t.Errorf("logger received wrong error: got %v, want %v", loggedErr, wantErr)
	}
	if loggedGuildID != "test-guild" {
		t.Errorf("logger received wrong guildID: got %q, want %q", loggedGuildID, "test-guild")
	}
}

func TestWrap_RecoversPanic(t *testing.T) {
	var loggedErr error
	var notifiedMsg string

	logErr := func(s *discordgo.Session, cfg *config.Configs, err error, guildID string) {
		loggedErr = err
	}
	notifier := func(s *discordgo.Session, i *discordgo.InteractionCreate, msg string) error {
		notifiedMsg = msg
		return nil
	}

	h := func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) error {
		panic("oh no")
	}

	wrapped := wrapWithDeps(h, logErr, notifier)

	// If wrap fails to recover, this defer will catch the propagated panic and
	// fail the test instead of crashing the test runner.
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("panic propagated past wrap: %v", r)
		}
	}()

	wrapped(nil, minInteraction(), &config.Configs{})

	if loggedErr == nil {
		t.Fatal("expected logger to receive a panic error, got nil")
	}
	if !strings.Contains(loggedErr.Error(), "oh no") {
		t.Errorf("logged error doesn't mention panic value: %v", loggedErr)
	}
	if notifiedMsg == "" {
		t.Error("expected user notifier to be called after panic recovery")
	}
}

func TestCommandHandlers_AllNonNil(t *testing.T) {
	if len(CommandHandlers) == 0 {
		t.Fatal("CommandHandlers is empty — dispatch table missing")
	}
	for name, h := range CommandHandlers {
		if h == nil {
			t.Errorf("CommandHandlers[%q] is nil", name)
		}
	}
}

func TestComponentHandlers_AllNonNil(t *testing.T) {
	if len(ComponentHandlers) == 0 {
		t.Fatal("ComponentHandlers is empty — dispatch table missing")
	}
	for name, h := range ComponentHandlers {
		if h == nil {
			t.Errorf("ComponentHandlers[%q] is nil", name)
		}
	}
}

func TestCommands_AllHaveDescriptions(t *testing.T) {
	if len(Commands) == 0 {
		t.Fatal("Commands slice is empty — registration list missing")
	}
	for _, cmd := range Commands {
		if cmd == nil {
			t.Error("Commands contains a nil entry")
			continue
		}
		if cmd.Name == "" {
			t.Errorf("command %+v has empty Name", cmd)
		}
		if cmd.Description == "" {
			t.Errorf("command %q has empty Description (Discord rejects this)", cmd.Name)
		}
		for _, opt := range cmd.Options {
			if opt.Description == "" {
				t.Errorf("command %q option %q has empty Description (Discord rejects this)", cmd.Name, opt.Name)
			}
		}
	}
}

func TestWrap_HappyPathDoesNothing(t *testing.T) {
	logCalled := false
	notifyCalled := false

	logErr := func(s *discordgo.Session, cfg *config.Configs, err error, guildID string) {
		logCalled = true
	}
	notifier := func(s *discordgo.Session, i *discordgo.InteractionCreate, msg string) error {
		notifyCalled = true
		return nil
	}

	h := func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) error {
		return nil
	}

	wrapped := wrapWithDeps(h, logErr, notifier)
	wrapped(nil, minInteraction(), &config.Configs{})

	if logCalled {
		t.Error("logger should not be called on success")
	}
	if notifyCalled {
		t.Error("notifier should not be called on success")
	}
}

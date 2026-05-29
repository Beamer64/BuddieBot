package helper

import (
	"fmt"
	"strings"
	"sync"
)

// Registered top-level command IDs (name → Discord command ID), populated once
// after command registration. Clickable command mentions like
// "</tuuck cmd-list:123>" need the live ID Discord assigns, and it's the
// top-level command's ID even for subcommands — so one entry per command.
var (
	commandIDMu sync.RWMutex
	commandIDs  = map[string]string{}
)

// SetCommandIDs records the registered command IDs so CommandMention can render
// clickable links. Called after (re)registration; replaces any prior set.
func SetCommandIDs(ids map[string]string) {
	commandIDMu.Lock()
	defer commandIDMu.Unlock()
	commandIDs = make(map[string]string, len(ids))
	for name, id := range ids {
		commandIDs[name] = id
	}
}

// CommandMention renders a clickable Discord command link, e.g.
// CommandMention("tuuck", "cmd-list") → "</tuuck cmd-list:123>". sub is the
// optional subcommand path. When the ID isn't known yet (registry unpopulated,
// e.g. in tests), it degrades to plain "/tuuck cmd-list" text rather than
// emitting a broken mention.
func CommandMention(name string, sub ...string) string {
	full := name
	if len(sub) > 0 {
		full = name + " " + strings.Join(sub, " ")
	}

	commandIDMu.RLock()
	id := commandIDs[name]
	commandIDMu.RUnlock()

	if id == "" {
		return "/" + full
	}
	return fmt.Sprintf("</%s:%s>", full, id)
}

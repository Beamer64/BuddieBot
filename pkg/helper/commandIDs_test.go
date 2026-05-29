package helper

import "testing"

func TestCommandMention(t *testing.T) {
	// Snapshot and restore the package registry so this test is independent.
	commandIDMu.Lock()
	saved := commandIDs
	commandIDMu.Unlock()
	t.Cleanup(func() { SetCommandIDs(saved) })

	// Unknown ID → plain-text fallback, never a broken "</x:>".
	if got := CommandMention("tuuck", "cmd-list"); got != "/tuuck cmd-list" {
		t.Errorf("fallback = %q, want %q", got, "/tuuck cmd-list")
	}

	SetCommandIDs(map[string]string{"tuuck": "123", "user": "456"})

	cases := []struct {
		name string
		sub  []string
		want string
	}{
		{"tuuck", []string{"cmd-list"}, "</tuuck cmd-list:123>"},
		{"user", []string{"forget-me"}, "</user forget-me:456>"},
		{"user", nil, "</user:456>"},
		{"missing", []string{"x"}, "/missing x"},
	}
	for _, c := range cases {
		if got := CommandMention(c.name, c.sub...); got != c.want {
			t.Errorf("CommandMention(%q, %v) = %q, want %q", c.name, c.sub, got, c.want)
		}
	}
}

package prefix

import (
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// TestPrefCmdsListMatchesSwitch ensures the PrefCmdsList map (the source of
// truth for the prefix-command catalogue, including /tuuck rendering) stays
// in sync with the case labels in ParsePrefixCmds. If a case is added or
// removed without updating PrefCmdsList, the dispatcher and the help list
// drift — this test catches it.
func TestPrefCmdsListMatchesSwitch(t *testing.T) {
	src, err := os.ReadFile("prefixCmds.go")
	if err != nil {
		t.Fatalf("read prefixCmds.go: %v", err)
	}
	re := regexp.MustCompile(`case "([a-z0-9_-]+)":`)
	var fromSwitch []string
	for _, m := range re.FindAllStringSubmatch(string(src), -1) {
		fromSwitch = append(fromSwitch, m[1])
	}
	fromList := make([]string, 0, len(PrefCmdsList))
	for name := range PrefCmdsList {
		fromList = append(fromList, name)
	}
	sort.Strings(fromSwitch)
	sort.Strings(fromList)
	if strings.Join(fromSwitch, ",") != strings.Join(fromList, ",") {
		t.Fatalf("PrefCmdsList drifted from switch cases:\n  switch: %v\n  list:   %v", fromSwitch, fromList)
	}
}

// TestNoShowCmdsAreKnown guards against typos in NoShowCmds — every hidden
// entry must be a real prefix command (i.e. a key in PrefCmdsList). A typo
// would silently hide nothing.
func TestNoShowCmdsAreKnown(t *testing.T) {
	for _, hidden := range NoShowCmds {
		if _, ok := PrefCmdsList[hidden]; !ok {
			t.Errorf("NoShowCmds entry %q has no matching PrefCmdsList key", hidden)
		}
	}
}

// TestSplitPrefixCommand covers the parser's STRICT contract: a tight prefix
// ("$") matches only when there's no space between it and the command word;
// a trailing-space prefix ("plz ") matches only when exactly one space
// follows the literal prefix text. Whitespace inside param must survive
// (palindrome relies on this).
func TestSplitPrefixCommand(t *testing.T) {
	cases := []struct {
		name        string
		content     string
		prefix      string
		wantOK      bool
		wantCommand string
		wantParam   string
	}{
		{
			name:        "default prefix tight",
			content:     "$roman 5",
			prefix:      "$",
			wantOK:      true,
			wantCommand: "roman",
			wantParam:   "5",
		},
		{
			// Strict matching: extra space after a no-trailing-space prefix means
			// the user didn't intend a command, bot stays silent.
			name:    "default prefix rejects stray space after",
			content: "$ roman 5",
			prefix:  "$",
			wantOK:  false,
		},
		{
			name:        "trailing-space prefix natural usage",
			content:     "plz roman 5",
			prefix:      "plz ",
			wantOK:      true,
			wantCommand: "roman",
			wantParam:   "5",
		},
		{
			name:    "trailing-space prefix requires the space",
			content: "plzroman 5",
			prefix:  "plz ",
			wantOK:  false,
		},
		{
			// Double space — first space is consumed by the prefix, the second
			// counts as "stray" under strict matching. Bot ignores.
			name:    "trailing-space prefix rejects double space",
			content: "plz  roman 5",
			prefix:  "plz ",
			wantOK:  false,
		},
		{
			// No-trailing-space prefix is also strict — "plz" requires "plzcmd",
			// not "plz cmd". Type the prefix as configured or not at all.
			name:    "no-trailing-space prefix rejects a stray space",
			content: "plz roman 5",
			prefix:  "plz",
			wantOK:  false,
		},
		{
			name:        "param keeps internal whitespace",
			content:     "$palindrome  abc  cba",
			prefix:      "$",
			wantOK:      true,
			wantCommand: "palindrome",
			wantParam:   " abc  cba", // exactly one leading space consumed by SplitN
		},
		{
			name:        "command only, no param",
			content:     "$weast",
			prefix:      "$",
			wantOK:      true,
			wantCommand: "weast",
			wantParam:   "",
		},
		{
			name:    "no prefix match",
			content: "hello world",
			prefix:  "$",
			wantOK:  false,
		},
		{
			name:    "empty content",
			content: "",
			prefix:  "$",
			wantOK:  false,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			gotCmd, gotParam, gotOK := splitPrefixCommand(c.content, c.prefix)
			if gotOK != c.wantOK {
				t.Fatalf("ok = %v, want %v", gotOK, c.wantOK)
			}
			if !c.wantOK {
				return
			}
			if gotCmd != c.wantCommand {
				t.Errorf("command = %q, want %q", gotCmd, c.wantCommand)
			}
			if gotParam != c.wantParam {
				t.Errorf("param = %q, want %q", gotParam, c.wantParam)
			}
		})
	}
}

package prefix

import (
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// TestPrefixNamesMatchSwitch ensures the Names slice stays in sync with the
// case labels in ParsePrefixCmds. If a case is added or removed without
// updating Names, the bot's startup count will drift — this test catches it.
func TestPrefixNamesMatchSwitch(t *testing.T) {
	src, err := os.ReadFile("prefixCmds.go")
	if err != nil {
		t.Fatalf("read prefixCmds.go: %v", err)
	}
	re := regexp.MustCompile(`case "([a-z0-9_-]+)":`)
	var fromSwitch []string
	for _, m := range re.FindAllStringSubmatch(string(src), -1) {
		fromSwitch = append(fromSwitch, m[1])
	}
	fromNames := append([]string(nil), Names...)
	sort.Strings(fromSwitch)
	sort.Strings(fromNames)
	if strings.Join(fromSwitch, ",") != strings.Join(fromNames, ",") {
		t.Fatalf("Names drifted from switch cases:\n  switch: %v\n  names:  %v", fromSwitch, fromNames)
	}
}

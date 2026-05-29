package prefix

import (
	"strings"
	"testing"
)

func TestBuildReleaseNotesEmbed(t *testing.T) {
	a := buildReleaseNotesEmbed()
	b := buildReleaseNotesEmbed()

	// Fresh instance per call — sendReleaseNotes stamps the invoking user as
	// Author, so a shared pointer would leak that mutation across sends.
	if a == b {
		t.Fatal("expected a new embed pointer per call")
	}
	a.Author.Name = "tester"
	if b.Author.Name != "" {
		t.Error("mutating one embed's author bled into another (shared state)")
	}

	if len(a.Fields) == 0 {
		t.Fatal("expected release-notes fields")
	}
	for _, f := range a.Fields {
		if strings.TrimSpace(f.Name) == "" || strings.TrimSpace(f.Value) == "" {
			t.Errorf("field has empty name/value: %+v", f)
		}
	}

	// No IDs are registered in the test binary, so command mentions degrade to
	// plain "/cmd" text — confirms the helper.CommandMention wiring is in place.
	found := false
	for _, f := range a.Fields {
		if strings.Contains(f.Value, "/tuuck cmd-list") {
			found = true
		}
	}
	if !found {
		t.Error("expected a field carrying the /tuuck cmd-list mention fallback")
	}
}

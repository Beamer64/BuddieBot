package helper

import (
	"testing"
)

// TestIsLaunchedByDebugger_EnvVar covers the BUDDIEBOT_FORCE_DEV escape hatch
// — setting it to any non-empty value makes the function return true without
// requiring Delve. This lets devs run from the IDE's plain "Run" mode (no
// debugger attached) while still landing on the Test config path.
//
// t.Setenv is per-test scoped and automatically restored at test end, so
// this can't pollute later tests in the same package.
func TestIsLaunchedByDebugger_EnvVar(t *testing.T) {
	t.Setenv("BUDDIEBOT_FORCE_DEV", "1")
	if !IsLaunchedByDebugger() {
		t.Error("expected true when BUDDIEBOT_FORCE_DEV=1")
	}

	// Any non-empty value triggers the override — we don't gate on "1" or
	// "true", just non-empty. A typo like "yse" would still flip it (the
	// risk of accidentally landing on test config is much lower than the
	// risk of accidentally landing on prod, so we lean permissive).
	t.Setenv("BUDDIEBOT_FORCE_DEV", "yes")
	if !IsLaunchedByDebugger() {
		t.Error("expected true when BUDDIEBOT_FORCE_DEV=yes")
	}

	// Explicit empty value does NOT trigger the override — falls through to
	// the dlv parent-process check. We can't assert what that returns from a
	// test (the result depends on how `go test` was launched), so we just
	// exercise the code path to confirm no panic.
	t.Setenv("BUDDIEBOT_FORCE_DEV", "")
	_ = IsLaunchedByDebugger()
}

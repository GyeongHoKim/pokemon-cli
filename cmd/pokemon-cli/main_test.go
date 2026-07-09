package main

import (
	"slices"
	"testing"
)

func TestSyncOutputSafeEnvironStripsSessionSignals(t *testing.T) {
	t.Setenv("WT_SESSION", "some-guid")
	t.Setenv("TERM_PROGRAM", "vscode")
	t.Setenv("TERM", "xterm-256color")

	got := syncOutputSafeEnviron()

	for _, kv := range got {
		if slices.Contains([]string{"WT_SESSION=some-guid", "TERM_PROGRAM=vscode"}, kv) {
			t.Errorf("syncOutputSafeEnviron() kept %q, want stripped", kv)
		}
	}
	if !slices.Contains(got, "TERM=xterm-256color") {
		t.Error("syncOutputSafeEnviron() dropped real TERM value, want preserved")
	}
	if !slices.Contains(got, "SSH_TTY=/dev/pts/0") {
		t.Error("syncOutputSafeEnviron() missing stubbed SSH_TTY")
	}
}

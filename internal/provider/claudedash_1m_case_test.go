package provider

import "testing"

// A provider may name a model with the mark in capitals
// (relay/claude-opus-4.6[1M]). Claude Code takes the mark off before it
// asks, so the gateway is given the dashed id with no mark
// (relay/claude-opus-4-6); that one must come back to the served id, or it
// is passed on as a model the provider has not.
func TestClaude1MMarkInCapitals(t *testing.T) {
	for _, ref := range []string{
		"relay/claude-opus-4.6[1M]",
		"relay/claude-opus-4.6[1m]",
		"claude-opus-4.6[1M]",
	} {
		if !claudeHas1M(ref) {
			t.Errorf("claudeHas1M(%q) = false, want true", ref)
		}
		bare := ref[:len(ref)-len(claude1MMark)]
		if got := claudeCut1M(ref); got != bare {
			t.Errorf("claudeCut1M(%q) = %q, want %q", ref, got, bare)
		}
	}
	// an id with no mark, and a shorter one, are as they are
	for _, ref := range []string{"claude-opus-4.6", "claude-opus-4", ""} {
		if claudeHas1M(ref) {
			t.Errorf("claudeHas1M(%q) = true, want false", ref)
		}
		if got := claudeCut1M(ref); got != ref {
			t.Errorf("claudeCut1M(%q) = %q, want %q", ref, got, ref)
		}
	}
}

// claudeSplit reads a capital mark as the mark: the id it gives back is the
// model, and the rest says the mark the way Claude Code writes it.
func TestClaudeSplitCapitalMark(t *testing.T) {
	for in, want := range map[string][2]string{
		"relay/claude-opus-4.6[1M]":         {"relay/claude-opus-4.6", claude1MMark},
		"relay/claude-opus-4.6[1m]":         {"relay/claude-opus-4.6", claude1MMark},
		"relay/claude-opus-4.6":             {"relay/claude-opus-4.6", ""},
		"relay/claude-opus-4.6-us:high[1M]": {"relay/claude-opus-4.6-us", ":high" + claude1MMark},
		"relay/claude-sonnet-4.5[1M]":       {"relay/claude-sonnet-4.5", claude1MMark},
		"claude-opus-4-6":                   {"claude-opus-4-6", ""},
		"r/claude-opus-4-6-us[1m]":          {"r/claude-opus-4-6-us", claude1MMark},
		"r/claude-opus-4-6-us:high[1m]":     {"r/claude-opus-4-6-us", ":high" + claude1MMark},
		"r/claude-opus-4-6-us:high":         {"r/claude-opus-4-6-us", ":high"},
		"r/claude-opus-4-6-us":              {"r/claude-opus-4-6-us", ""},
	} {
		if id, rest := claudeSplit(in); id != want[0] || rest != want[1] {
			t.Errorf("claudeSplit(%q) = %q %q, want %q %q", in, id, rest, want[0], want[1])
		}
	}
}

// A served id marked [1M] is spelled dashed with the mark taken off, and
// the dashed spelling comes back to it with the mark as the provider writes
// it: [1M], which is what the provider serves.
func TestClaudeSpelledAndUndashedAcross1MCase(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("XDG_CONFIG_HOME", home+"/.config")
	t.Setenv("XDG_CACHE_HOME", home+"/.cache")
	if err := Save(Provider{ID: "relay", Name: "Relay", Anthropic: "https://example.test", Key: "k",
		Models: []string{"claude-opus-4.6[1M]", "claude-sonnet-4.5[1M]", "gpt-5.1[1M]"}}); err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		served, dashed, listed string
	}{
		// what magpie serves, what Claude Code sends once it takes the mark
		// off, and what /model lists (the dashed id, marked: Claude Code
		// takes an unmarked Claude model for 200K)
		{"relay/claude-opus-4.6[1M]", "relay/claude-opus-4-6", "relay/claude-opus-4-6[1m]"},
		{"relay/claude-sonnet-4.5[1M]", "relay/claude-sonnet-4-5", "relay/claude-sonnet-4-5[1m]"},
	} {
		if got := ClaudeSpelled(c.served); got != c.listed {
			t.Errorf("ClaudeSpelled(%q) = %q, want %q", c.served, got, c.listed)
		}
		// the bare dashed id is what Claude Code asks with
		got, ok := ClaudeUndashed(c.dashed)
		if !ok || got != c.served {
			t.Errorf("ClaudeUndashed(%q) = %q %v, want %q true", c.dashed, got, ok, c.served)
		}
		// and with the mark Claude Code puts on the id it lists
		got, ok = ClaudeUndashed(c.listed)
		if !ok || got != c.served {
			t.Errorf("ClaudeUndashed(%q) = %q %v, want %q true", c.listed, got, ok, c.served)
		}
		// an effort fixed on it is kept
		got, ok = ClaudeUndashed(c.dashed + ":high")
		if !ok || got != c.served+":high" {
			t.Errorf("ClaudeUndashed(%q) = %q %v, want %q true", c.dashed+":high", got, ok, c.served+":high")
		}
	}
	// another vendor's dotted version is as it is, either way
	if got := ClaudeSpelled("relay/gpt-5.1[1M]"); got != "relay/gpt-5.1[1M]" {
		t.Errorf("ClaudeSpelled(relay/gpt-5.1[1M]) = %q", got)
	}
}

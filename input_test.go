package main

import (
	"bufio"
	"strings"
	"testing"
)

// Under `go test` stdin is not a terminal, so these exercise the documented
// non-terminal fallback path.

func TestReadSecretTrimsAndReturnsLine(t *testing.T) {
	r := bufio.NewReader(strings.NewReader("hunter2\n"))
	got, err := ReadSecret(r, "pw: ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "hunter2" {
		t.Fatalf("got %q, want %q", got, "hunter2")
	}
}

func TestReadSecretReadsSequentialPrompts(t *testing.T) {
	r := bufio.NewReader(strings.NewReader("first\nsecond\n"))
	for _, want := range []string{"first", "second"} {
		got, err := ReadSecret(r, "pw: ")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	}
}

func TestReadSecretConfirmedAcceptsMatch(t *testing.T) {
	r := bufio.NewReader(strings.NewReader("Passw0rd\nPassw0rd\n"))
	got, err := ReadSecretConfirmed(r, "new: ", "confirm: ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "Passw0rd" {
		t.Fatalf("got %q, want %q", got, "Passw0rd")
	}
}

func TestReadSecretConfirmedRejectsMismatch(t *testing.T) {
	r := bufio.NewReader(strings.NewReader("Passw0rd\nPassw0rdd\n"))
	if _, err := ReadSecretConfirmed(r, "new: ", "confirm: "); err == nil {
		t.Fatal("expected mismatch to be rejected, got nil error")
	}
}

func TestReadSecretConfirmedRejectsMismatchSameLength(t *testing.T) {
	r := bufio.NewReader(strings.NewReader("Passw0rd\nPassw0rx\n"))
	if _, err := ReadSecretConfirmed(r, "new: ", "confirm: "); err == nil {
		t.Fatal("expected same-length mismatch to be rejected, got nil error")
	}
}

func TestSecureCompare(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{
		{"", "", true},
		{"abc", "abc", true},
		{"abc", "abd", false},
		{"abc", "abcd", false},
		{"abcd", "abc", false},
		{"abc", "", false},
	}
	for _, c := range cases {
		if got := SecureCompare(c.a, c.b); got != c.want {
			t.Errorf("SecureCompare(%q, %q) = %v, want %v", c.a, c.b, got, c.want)
		}
	}
}

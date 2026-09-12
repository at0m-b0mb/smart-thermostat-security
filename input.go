package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// ReadSecret prompts for a secret (password or PIN) and reads it without
// echoing the characters to the terminal.
//
// Echoing secrets is a real exposure even on a trusted machine: the value
// lands in terminal scrollback, in any screen recording or screen share, and
// in view of anyone behind the operator. Every other control in this system
// assumes the credential is known only to its owner, so the credential must
// not be printed while it is being entered.
//
// When stdin is not a terminal (piped input, automated test harness) there is
// no echo to suppress, so the value is read normally and the caller is warned
// that input is visible.
func ReadSecret(reader *bufio.Reader, prompt string) (string, error) {
	fmt.Print(prompt)

	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		fmt.Println("[warning: stdin is not a terminal; input will be visible]")
		secret, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(secret), nil
	}

	// ReadPassword reads straight from the fd. The shared bufio.Reader is
	// drained line-by-line on an interactive terminal, so there is no
	// buffered input for it to strand here.
	raw, err := term.ReadPassword(fd)
	fmt.Println() // ReadPassword consumes the newline without echoing it
	if err != nil {
		return "", errors.New("failed to read secret input")
	}
	return strings.TrimSpace(string(raw)), nil
}

// ReadSecretConfirmed prompts twice and fails unless both entries match.
// With echo off the operator cannot see a typo, so a second entry is the only
// protection against locking yourself out with a mistyped new credential.
func ReadSecretConfirmed(reader *bufio.Reader, prompt, confirmPrompt string) (string, error) {
	first, err := ReadSecret(reader, prompt)
	if err != nil {
		return "", err
	}
	second, err := ReadSecret(reader, confirmPrompt)
	if err != nil {
		return "", err
	}
	if !SecureCompare(first, second) {
		return "", errors.New("entries do not match")
	}
	return first, nil
}

// terminalState captures the tty mode at startup so a signal-driven exit can
// put the terminal back. Without this, Ctrl+C during a masked prompt leaves
// the user's shell with echo disabled.
var terminalState *term.State

func saveTerminalState() {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return
	}
	if state, err := term.GetState(fd); err == nil {
		terminalState = state
	}
}

func restoreTerminalState() {
	if terminalState == nil {
		return
	}
	_ = term.Restore(int(os.Stdin.Fd()), terminalState)
}

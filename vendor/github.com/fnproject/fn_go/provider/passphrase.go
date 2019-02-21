package provider

import (
	"errors"
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	"syscall"
)

// PassPhraseSource provides a passphrase from an external source (e.g. a user prompt or a terminal)
type PassPhraseSource interface {
	//ChallengeForPassPhrase requests a passphrase with a given prompt from the user
	ChallengeForPassPhrase(id, prompt string) (string, error)
}

// NopPassPhraseSource always returns an error when request for a passphrase
type NopPassPhraseSource struct{}

//TerminalPassPhraseSource requests a passphrase from the terminal
type TerminalPassPhraseSource struct{}

func (*NopPassPhraseSource) ChallengeForPassPhrase(id, msg string) (string, error) {
	return "", errors.New("no pass phrase available")
}

func (*TerminalPassPhraseSource) ChallengeForPassPhrase(id, msg string) (string, error) {
	fmt.Print(msg)
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}
	password := string(bytePassword)
	fmt.Println()

	return password, nil
}

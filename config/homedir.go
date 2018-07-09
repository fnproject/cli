package config

import (
	"log"

	homedir "github.com/mitchellh/go-homedir"
)

func GetHomeDir() string {
	home, err := homedir.Dir()
	if err != nil {
		log.Fatalf("Could not get home directory: %s\n", err)
	}

	return home
}

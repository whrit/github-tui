package config

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
)

type github struct {
	Owner string
	Repo  string
	Token string `yaml:"token"`
}

type app struct {
	File string `yaml:"file"`
}

const readThisMessage = "read this https://github.com/skanehira/github-tui?tab=readme-ov-file#settings to know more"

var (
	GitHub github
	App    app
)

func Init() {
	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatal(err)
	}

	ghtDir := filepath.Join(configDir, "ght")
	if err := os.MkdirAll(ghtDir, 0o700); err != nil {
		log.Fatal(err)
	}

	logFile := filepath.Join(ghtDir, "debug.log")
	output, err := os.Create(logFile)
	if err != nil {
		log.Fatal(err)
	}

	log.SetOutput(io.MultiWriter(output, os.Stderr))

	configFile := filepath.Join(ghtDir, "config.yaml")

	b, err := os.ReadFile(configFile)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Fatal(err)
		}

		// First run: prompt for token and write config
		token := promptForToken()
		if err := writeConfig(ghtDir, token); err != nil {
			log.Fatalf("could not write config: %v", err)
		}
		// Populate GitHub.Token so the rest of Init() can proceed
		GitHub.Token = token
		App.File = configFile
		return
	}

	var conf struct {
		GitHub github `yaml:"github"`
	}

	if err := yaml.Unmarshal(b, &conf); err != nil {
		log.Fatalf("cannot deserialize config file: %s", err)
	}

	if conf.GitHub.Token == "" {
		log.Fatalf("github token is empty, %s", readThisMessage)
	}

	App.File = configFile
	GitHub = conf.GitHub
}

// promptForToken reads a GitHub PAT from stdin interactively.
func promptForToken() string {
	fmt.Fprintln(os.Stderr, "\nWelcome to ght! No config file found.")
	fmt.Fprintln(os.Stderr, "You need a GitHub Personal Access Token (classic) with repo, workflow, and read:org scopes.")
	fmt.Fprintln(os.Stderr, "Create one at: https://github.com/settings/tokens")
	fmt.Fprint(os.Stderr, "\nEnter your GitHub token: ")

	scanner := bufio.NewScanner(os.Stdin)
	for {
		scanner.Scan()
		token := strings.TrimSpace(scanner.Text())
		if token != "" {
			return token
		}
		fmt.Fprint(os.Stderr, "Token cannot be empty. Enter your GitHub token: ")
	}
}

// writeConfig writes a minimal config.yaml to the ght config directory.
func writeConfig(dir, token string) error {
	configFile := filepath.Join(dir, "config.yaml")
	content := fmt.Sprintf("github:\n  token: %s\n", token)
	return os.WriteFile(configFile, []byte(content), 0o600)
}

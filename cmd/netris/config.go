package main

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"

	"codeberg.org/tslocum/netris/pkg/event"
	"gopkg.in/yaml.v2"
)

type appConfig struct {
	Input  map[event.GameAction][]string // Keybinds
	Colors map[event.GameColor]string
	Name   string
}

var config = &appConfig{
	Input:  make(map[event.GameAction][]string),
	Colors: make(map[event.GameColor]string),
	Name:   "Anonymous",
}

var regexpColor = regexp.MustCompile(`^#([0-9a-f]{3}|[0-9a-f]{6})$`)

func defaultConfigPath() (string, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	} else if homedir == "" {
		return "", errors.New("a blank path to user homedir was returned")
	}

	dir := path.Join(homedir, ".config", "netris")
	_, err = os.Stat(path.Join(dir, "config.yaml"))
	if err == nil {
		return path.Join(dir, "config.yaml"), nil
	}

	_, err = os.Stat(path.Join(dir, "config.yml"))
	if err == nil {
		return path.Join(dir, "config.yml"), nil
	}

	return path.Join(dir, "config.yaml"), nil
}

func serializeConfig(path string) error {
	if path == "" {
		return errors.New("failed to serialize: no path specified")

	}

	out, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %s", err)
	}

	os.MkdirAll(filepath.Dir(path), 0755)
	err = os.WriteFile(path, out, 0644)
	if err != nil {
		return fmt.Errorf("failed to write to %s: %s", path, err)
	}

	return nil
}

func deserializeConfig(path string) error {
	if path == "" {
		return errors.New("failed to deserialize: no path specified")
	}

	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil
	}

	configData, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file: %s", err)
	}

	err = yaml.Unmarshal(configData, config)
	if err != nil {
		return fmt.Errorf("failed to parse file: %s", err)
	}

	return nil
}

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"gitlab.com/tslocum/netris/pkg/event"
	"gopkg.in/yaml.v2"
)

type appConfig struct {
	Input map[event.GameAction][]string // Keybinds
	Name  string
}

var config = &appConfig{}

func defaultConfigPath() string {
	homedir, err := os.UserHomeDir()
	if err == nil && homedir != "" {
		return path.Join(homedir, ".config", "netris", "config.yaml")
	}

	return ""
}

func readConfig(configPath string) error {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if configPath != defaultConfigPath() {
			return fmt.Errorf("failed to read configuration: %s", err)
		}
		return nil
	}

	configData, err := ioutil.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read configuration: %s", err)
	}

	err = yaml.Unmarshal(configData, config)
	if err != nil {
		return fmt.Errorf("failed to parse configuration: %s", err)
	}

	return nil
}

func saveConfig(configPath string) error {
	config.Name = nickname

	out, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %s", err)
	}

	os.MkdirAll(path.Dir(configPath), 0755) // Ignore error

	err = ioutil.WriteFile(configPath, out, 0644)
	if err != nil {
		return fmt.Errorf("failed to write to %s: %s", configPath, err)
	}
	return nil
}

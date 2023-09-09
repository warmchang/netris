package main

import (
	"regexp"

	"code.rocket9labs.com/tslocum/netris/pkg/event"
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

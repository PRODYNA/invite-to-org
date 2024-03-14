package main

import (
	config "github.com/prodyna/invite-to-org/config"
	"log/slog"
	"os"
)

func main() {
	c, err := config.New()
	if err != nil {
		slog.Error("Unable to create config", "error", err)
		os.Exit(1)
	}
	slog.Debug("Config", "config", c)

}

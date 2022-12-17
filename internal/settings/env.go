package settings

import (
	"fmt"

	"github.com/caarlos0/env/v6"
)

func LoadEnvSettings() (Settings, error) {
	var settings Settings
	if err := env.Parse(&settings); err != nil {
		return Settings{}, fmt.Errorf("load settings from environment: %w", err)
	}

	return settings, nil
}

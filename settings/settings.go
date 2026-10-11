// Package settings provides configuration values for the app.
package settings

import "os"

const envWorkDir = "JAO_WORKING_DIR"

const defaultWorkDir = "/opt/jao"

type Settings struct {
	WorkDir string
}

func New() Settings {
	workDir := os.Getenv(envWorkDir)

	if len(workDir) == 0 {
		workDir = defaultWorkDir
	}

	settings := Settings{
		WorkDir: workDir,
	}
	return settings
}

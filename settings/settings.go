// Package settings provides configuration values for the app.
package settings

import "os"

const envWorkDir = "JAO_WORKING_DIR"

// envPublicHost holds the public IP or domain of the node.
const envPublicHost = "JAO_PUBLIC_HOST"

const defaultWorkDir = "/opt/jao"

type Settings struct {
	WorkDir    string
	PublicHost string
}

func New() Settings {
	workDir := os.Getenv(envWorkDir)

	if len(workDir) == 0 {
		workDir = defaultWorkDir
	}

	settings := Settings{
		WorkDir:    workDir,
		PublicHost: os.Getenv(envPublicHost),
	}
	return settings
}

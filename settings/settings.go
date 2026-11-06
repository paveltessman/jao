// Package settings provides configuration values for the app.
package settings

import "os"

const (
	envWorkDir    = "JAO_WORKING_DIR"
	envPublicHost = "JAO_PUBLIC_HOST" // public IP or domain of the node
	envHttpAddr   = "JAO_HTTP_LISTEN_ADDR"
)

const (
	defaultWorkDir  = "/opt/jao"
	defaultHTTPAddr = "127.0.0.1:8001"
)

type Settings struct {
	WorkDir        string
	PublicHost     string
	HttpListenAddr string
}

func New() Settings {
	workDir := os.Getenv(envWorkDir)

	if len(workDir) == 0 {
		workDir = defaultWorkDir
	}

	httpAddr := os.Getenv(envHttpAddr)
	if httpAddr == "" {
		httpAddr = defaultHTTPAddr
	}

	settings := Settings{
		WorkDir:        workDir,
		PublicHost:     os.Getenv(envPublicHost),
		HttpListenAddr: httpAddr,
	}
	return settings
}

// Package settings provides configuration values for the app.
package settings

import (
	"errors"
	"os"
)

const (
	envWorkDir    = "JAO_WORKING_DIR"
	envPublicHost = "JAO_PUBLIC_HOST" // public IP or domain of the node
	envHttpAddr   = "JAO_HTTP_LISTEN_ADDR"
	envSecretKey  = "JAO_SECRET_KEY"
)

const (
	defaultWorkDir  = "/opt/jao"
	defaultHTTPAddr = "127.0.0.1:8001"
)

var (
	ErrSecretNotSet = errors.New("settings: secret key is not set")
)

type Settings struct {
	WorkDir        string
	PublicHost     string
	HttpListenAddr string
	SecretKey      string
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

	secretKey := os.Getenv(envSecretKey)
	if secretKey == "" {
		panic(ErrSecretNotSet)
	}

	settings := Settings{
		WorkDir:        workDir,
		PublicHost:     os.Getenv(envPublicHost),
		HttpListenAddr: httpAddr,
		SecretKey:      secretKey,
	}
	return settings
}

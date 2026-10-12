package workdir

import (
	"embed"
	"fmt"
	"io/fs"
)

//go:embed embed/*
var embedded embed.FS

type SubDir string

const AWG SubDir = "awg"

func FS(subDir SubDir) (fs.FS, error) {
	sub, err := fs.Sub(embedded, string(subDir))
	if err != nil {
		return nil, fmt.Errorf("workdir: can't root the fs at %s: %w", subDir, err)
	}
	return sub, nil
}

package workdir

import (
	"embed"
	"fmt"
	"io/fs"
	"path"
)

//go:embed embed/*
var embedded embed.FS

const root = "embed"

type SubDir string

const AWG SubDir = "awg"

func FS(subDir SubDir) (fs.FS, error) {
	sub, err := fs.Sub(embedded, path.Join(root, string(subDir)))
	if err != nil {
		return nil, fmt.Errorf("workdir: can't root the fs at %s: %w", subDir, err)
	}
	return sub, nil
}

package awg

import (
	"embed"
	"fmt"
	"io/fs"
)

//go:embed image/*
var image embed.FS

func BuildContext() (fs.FS, error) {
	sub, err := fs.Sub(image, "image")
	if err != nil {
		return nil, fmt.Errorf("awg: can't root the fs at image: %w", err)
	}
	return sub, nil
}

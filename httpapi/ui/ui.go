package ui

import (
	"io/fs"
	"net/http"

	"github.com/paveltessman/jao/frontend"
)

func Route() http.Handler {

	frontendDir, err := fs.Sub(frontend.FS, "out")
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()

	mux.Handle("/", http.FileServer(http.FS(frontendDir)))

	return mux
}

package api

import (
	"errors"
	"log"
	"net/http"

	"github.com/paveltessman/jao/keys"
	"github.com/paveltessman/jao/protocol/awg"
)

var ErrBadPathValue = errors.New("bad path value")

type ErrorSpec struct {
	internal   bool
	statusCode int
}

var errorSpecs = map[error]ErrorSpec{
	ErrBadPathValue:  {internal: false, statusCode: http.StatusBadRequest},
	keys.ErrNotFound: {internal: false, statusCode: http.StatusBadRequest},
	awg.ErrBadEnv:    {internal: true, statusCode: http.StatusInternalServerError},
}

func errorSpec(err error) (ErrorSpec, bool) {
	for sentinel, spec := range errorSpecs {
		if errors.Is(err, sentinel) {
			return spec, true
		}
	}
	return ErrorSpec{}, false
}

func statusCode(err error) int {
	if err == nil {
		return http.StatusOK
	}

	spec, ok := errorSpec(err)
	if !ok {
		return http.StatusInternalServerError
	}
	return spec.statusCode
}

func errorMessage(err error) string {
	if err == nil {
		return ""
	}

	spec, ok := errorSpec(err)
	if !ok || spec.internal {
		return "internal server error"
	}
	return err.Error()
}

func logError(err error) {
	if err == nil {
		return
	}

	spec, ok := errorSpec(err)
	if !ok || (ok && spec.internal) {
		log.Println(err)
	}
}

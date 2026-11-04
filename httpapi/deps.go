package httpapi

import "github.com/paveltessman/jao/keys"

type Deps struct {
	keyService *keys.Service
}

func NewDeps(keyService *keys.Service) Deps {
	if keyService == nil {
		panic("key service can't be nil")
	}
	return Deps{keyService: keyService}
}

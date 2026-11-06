package shared

import "github.com/paveltessman/jao/keys"

type Deps struct {
	KeyService keys.ServiceProto
}

func NewDeps(keyService *keys.Service) Deps {
	if keyService == nil {
		panic("key service can't be nil")
	}
	return Deps{KeyService: keyService}
}

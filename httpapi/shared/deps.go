package shared

import (
	"github.com/paveltessman/jao/auth"
	"github.com/paveltessman/jao/keys"
)

type Deps struct {
	KeyService  keys.ServiceProto
	AuthService *auth.Service
}

func NewDeps(keyService *keys.Service, authService *auth.Service) Deps {
	if keyService == nil {
		panic("key service can't be nil")
	}
	if authService == nil {
		panic("auth service can't be nil")
	}
	return Deps{KeyService: keyService, AuthService: authService}
}

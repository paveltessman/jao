package keys

import "context"

// Installed describes a protocol container on the node.
type Installed struct {
	Name     string `json:"name"`
	Protocol string `json:"proto"`
}

type ServiceProto interface {
	Discover(ctx context.Context) ([]Installed, error)
	CreateKey(ctx context.Context, name string) (string, error)
}

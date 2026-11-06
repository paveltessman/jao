package keys

import (
	"context"

	"github.com/paveltessman/jao/keys"
)

var _ keys.ServiceProto = (*FakeService)(nil)

type FakeService struct {
	Installed    []keys.Installed
	DiscoverErr  error
	CreatedKey   string
	CreateKeyErr error
}

func (s *FakeService) Discover(ctx context.Context) ([]keys.Installed, error) {
	return s.Installed, s.DiscoverErr
}

func (s *FakeService) CreateKey(ctx context.Context, name string) (string, error) {
	return s.CreatedKey, s.CreateKeyErr
}

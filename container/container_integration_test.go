//go:build integration

package container

import (
	"context"
	"testing"
	"time"
)

func newService(t *testing.T) *Service {
	t.Helper()

	svc, err := New()
	if err != nil {
		t.Fatalf("create service: %v", err)
	}
	t.Cleanup(func() {
		if err := svc.Close(); err != nil {
			t.Errorf("close service: %v", err)
		}
	})

	return svc
}

func TestPing(t *testing.T) {

	svc := newService(t)

	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	ping, err := svc.Ping(ctx)
	if err != nil {
		t.Fatalf("ping: %v", err)
	}

	if ping.APIVersion == "" {
		t.Errorf("want API version, got empty string")
	}
	if ping.OSType == "" {
		t.Errorf("want OS type, got empty string")
	}
}

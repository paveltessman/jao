package api

import (
	"errors"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/paveltessman/jao/httpserver/shared"
	"github.com/paveltessman/jao/keys"
	keystest "github.com/paveltessman/jao/testkit/keys"
)

func deps(keySvc *keystest.FakeService) shared.Deps {
	d := shared.Deps{
		KeyService: keySvc,
	}
	return d
}

func TestListContainers(t *testing.T) {
	r := httptest.NewRequest("GET", "/containers/", nil)

	cases := []struct {
		desc        string
		installed   []keys.Installed
		discoverErr error
	}{
		{"happy path", []keys.Installed{{Name: "aoeo", Protocol: "hoba"}}, nil},
		{"no protocol is installed", []keys.Installed{}, keys.ErrNoProtocol},
	}

	for _, c := range cases {
		keysSvc := &keystest.FakeService{Installed: c.installed, DiscoverErr: c.discoverErr}
		svc := Service{deps(keysSvc)}

		containers, err := svc.listContainers(r)

		if !errors.Is(err, c.discoverErr) {
			t.Errorf("%s: error mismatch: got %q, want %q", c.desc, err, c.discoverErr)
		}

		if err != nil {
			continue
		}

		installed := containers.([]keys.Installed)

		if !slices.Equal(installed, c.installed) {
			t.Errorf("%s: installed containers mismatch: got %q, want %q", c.desc, installed, c.installed)
		}
	}
}

func TestCreateKey(t *testing.T) {

	r := httptest.NewRequest("POST", "/containers/jao-awg/", nil)

	cases := []struct {
		desc          string
		createdKey    string
		createdKeyErr error
		pathValue     string
	}{
		{"happy path", "some client config", nil, "jao-awg"},
		{"bad container name", "", keys.ErrNotFound, "nope"},
		{"empty path value", "", ErrBadPathValue, ""},
	}

	for _, c := range cases {

		r.SetPathValue("name", c.pathValue)
		keysSvc := &keystest.FakeService{CreatedKey: c.createdKey, CreateKeyErr: c.createdKeyErr}
		svc := Service{deps(keysSvc)}

		payload, err := svc.createKey(r)

		if !errors.Is(err, c.createdKeyErr) {
			t.Errorf("%s: error mismatch: got %q, want %q", c.desc, err, c.createdKeyErr)
		}

		if err == nil && (payload != c.createdKey) {
			t.Errorf("%s: payload mismatch: got %q, want %q", c.desc, payload, c.createdKey)
		}
	}
}

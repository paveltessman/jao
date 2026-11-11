package password

import (
	"errors"
	"testing"
)

func TestHashVerify(t *testing.T) {

	const plain = "opa"

	encoded, err := Hash(plain)
	if err != nil {
		t.Fatalf("unable to create hash: %v", err)
	}

	matched, err := Verify(plain, encoded)
	if err != nil {
		t.Fatalf("unable to verify: %v", err)
	}

	if !matched {
		t.Fatalf("not matched")
	}
}

func TestVerifyWithWrongPassword(t *testing.T) {

	encoded, err := Hash("opa")
	if err != nil {
		t.Fatalf("unable to create hash: %v", err)
	}

	matched, err := Verify("opa x2", encoded)
	if err != nil {
		t.Fatalf("unable to verify: %v", err)
	}

	if matched {
		t.Fatalf("incorrect match")
	}
}

func TestVerifyWithInvalidHash(t *testing.T) {

	_, err := Verify("opa x3", "opa x4")
	if !errors.Is(err, ErrParseHash) {
		t.Fatalf("expected error, got: %v", err)
	}
}

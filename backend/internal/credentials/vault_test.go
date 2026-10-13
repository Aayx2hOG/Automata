package credentials

import (
	"bytes"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestEncryptionBindsOwnerAndCredential(t *testing.T) {
	v, err := New(nil, strings.Repeat("ab", 32))
	if err != nil {
		t.Fatal(err)
	}
	owner, id := uuid.New(), uuid.New()
	sealed := v.seal(owner, id, "Bearer secret-token")
	if bytes.Contains(sealed, []byte("secret-token")) {
		t.Fatal("stored plaintext")
	}
	if bytes.Equal(sealed, v.seal(owner, id, "Bearer secret-token")) {
		t.Fatal("reused nonce")
	}
	if plain, err := v.open(owner, id, sealed); err != nil || plain != "Bearer secret-token" {
		t.Fatalf("roundtrip: %q %v", plain, err)
	}
	for _, pair := range [][2]uuid.UUID{{uuid.New(), id}, {owner, uuid.New()}} {
		if _, err := v.open(pair[0], pair[1], sealed); err == nil {
			t.Fatal("accepted different credential identity")
		}
	}
	sealed[len(sealed)-1] ^= 1
	if _, err := v.open(owner, id, sealed); err == nil {
		t.Fatal("accepted corrupted ciphertext")
	}
	if _, err := New(nil, "short"); err == nil {
		t.Fatal("accepted weak key")
	}
}

func TestCredentialValidation(t *testing.T) {
	valid := Input{Name: "API", Host: "api.example.com", Header: "Authorization", Value: "Bearer secret"}
	if err := valid.Validate(); err != nil {
		t.Fatal(err)
	}
	for _, change := range []func(*Input){
		func(i *Input) { i.Host = "example.com/path" },
		func(i *Input) { i.Host = "user@example.com" },
		func(i *Input) { i.Header = "Host" },
		func(i *Input) { i.Header = "bad header" },
		func(i *Input) { i.Value = "secret\r\nInjected: true" },
	} {
		input := valid
		change(&input)
		if err := input.Validate(); err == nil {
			t.Fatalf("accepted invalid input: header=%q host=%q", input.Header, input.Host)
		}
	}
}

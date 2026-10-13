package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"strings"
	"testing"

	"github.com/Aayx2hOG/automata/internal/credentials"
	"github.com/google/uuid"
)

func TestCredentialOwnershipRotationAndDeletion(t *testing.T) {
	r, schedule := outboxFixture(t)
	ctx := context.Background()
	var owner uuid.UUID
	if err := r.pool.QueryRow(ctx, `SELECT owner_id FROM workflows WHERE id=$1`, schedule.WorkflowID).Scan(&owner); err != nil {
		t.Fatal(err)
	}
	vault, err := credentials.New(r.pool, strings.Repeat("ab", 32))
	if err != nil {
		t.Fatal(err)
	}
	input := credentials.Input{Name: "Test API", Host: "api.example.com", Header: "Authorization", Value: "Bearer secret-one"}
	saved, err := vault.Save(ctx, owner, uuid.Nil, input)
	if err != nil {
		t.Fatal(err)
	}
	destination, _ := url.Parse("https://api.example.com/data")
	if header, value, err := vault.Resolve(ctx, owner, saved.ID, destination); err != nil || header != "Authorization" || value != input.Value {
		t.Fatalf("resolve: %v", err)
	}
	items, err := vault.List(ctx, owner)
	data, _ := json.Marshal(items)
	if err != nil || len(items) != 1 || strings.Contains(string(data), "secret-one") || strings.Contains(string(data), "encrypted_value") {
		t.Fatal("list leaked credential or omitted metadata")
	}
	other := uuid.New()
	if _, _, err := vault.Resolve(ctx, other, saved.ID, destination); !errors.Is(err, credentials.ErrNotFound) {
		t.Fatalf("cross-owner resolve: %v", err)
	}
	if _, err := vault.Save(ctx, other, saved.ID, input); !errors.Is(err, credentials.ErrNotFound) {
		t.Fatalf("cross-owner update: %v", err)
	}
	if err := vault.Delete(ctx, other, saved.ID); !errors.Is(err, credentials.ErrNotFound) {
		t.Fatalf("cross-owner delete: %v", err)
	}
	for _, raw := range []string{"http://api.example.com/data", "https://other.example.com", "https://api.example.com:8443"} {
		u, _ := url.Parse(raw)
		if _, _, err := vault.Resolve(ctx, owner, saved.ID, u); err == nil {
			t.Fatalf("allowed destination %s", raw)
		}
	}
	input.Value = "Bearer secret-two"
	if _, err := vault.Save(ctx, owner, saved.ID, input); err != nil {
		t.Fatal(err)
	}
	if _, value, err := vault.Resolve(ctx, owner, saved.ID, destination); err != nil || value != input.Value {
		t.Fatal("rotation did not take effect")
	}
	if err := vault.Delete(ctx, owner, saved.ID); err != nil {
		t.Fatal(err)
	}
	if _, _, err := vault.Resolve(ctx, owner, saved.ID, destination); !errors.Is(err, credentials.ErrNotFound) {
		t.Fatal("deleted credential remained usable")
	}
}

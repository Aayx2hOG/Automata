package services

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"github.com/Aayx2hOG/automata/internal/models"
	"github.com/Aayx2hOG/automata/internal/repositories"
	"github.com/google/uuid"
	"strconv"
	"testing"
	"time"
)

type webhookRepo struct {
	repositories.WorkflowRepository
	wf models.Workflow
}

func (r webhookRepo) GetById(context.Context, uuid.UUID) (*models.Workflow, error) { return &r.wf, nil }
func TestWebhookSignature(t *testing.T) {
	s := &WorkflowService{workflows: webhookRepo{wf: models.Workflow{WebhookSecret: "test-secret", IsActive: true}}}
	body := []byte(`{"hello":"world"}`)
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	mac := hmac.New(sha256.New, []byte("test-secret"))
	mac.Write([]byte(ts + "."))
	mac.Write(body)
	sig := hex.EncodeToString(mac.Sum(nil))
	for _, tc := range []struct {
		ts, sig string
		body    []byte
		valid   bool
	}{{ts, sig, body, true}, {ts, "", body, false}, {"1", sig, body, false}, {ts, sig, []byte(`{}`), false}} {
		err := s.VerifyWebhook(context.Background(), uuid.New(), tc.ts, tc.sig, tc.body)
		if (err == nil) != tc.valid {
			t.Fatalf("verification: %v", err)
		}
	}
}

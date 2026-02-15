package gate

import (
	"errors"
	"path/filepath"
	"testing"
	"time"

	schemavoice "github.com/davidahmann/gait/core/schema/v1/voice"
	"github.com/davidahmann/gait/core/sign"
)

func TestCommitmentIntentToIntent(t *testing.T) {
	intent, err := CommitmentIntentToIntent(schemavoice.CommitmentIntent{
		SchemaID:        commitmentIntentSchemaID,
		SchemaVersion:   commitmentIntentSchemaV1,
		CreatedAt:       time.Date(2026, time.February, 15, 0, 0, 0, 0, time.UTC),
		ProducerVersion: "test",
		CallID:          "call_demo",
		TurnIndex:       2,
		CallSeq:         5,
		CommitmentClass: "quote",
		Context: schemavoice.CommitmentContext{
			Identity:  "agent.voice",
			Workspace: "/srv/voice",
			RiskClass: "high",
			SessionID: "sess_demo",
			RequestID: "req_demo",
		},
		QuoteMinCents: 1000,
		QuoteMaxCents: 2200,
	})
	if err != nil {
		t.Fatalf("commitment intent to intent: %v", err)
	}
	if intent.ToolName != "voice.commitment.quote" {
		t.Fatalf("unexpected tool name: %s", intent.ToolName)
	}
	if intent.Context.Identity != "agent.voice" || intent.Context.Workspace != "/srv/voice" {
		t.Fatalf("unexpected context mapping: %#v", intent.Context)
	}
	if intent.IntentDigest == "" || intent.ArgsDigest == "" {
		t.Fatalf("expected digests to be populated")
	}
	if _, ok := intent.Args["quote_min_cents"]; !ok {
		t.Fatalf("expected quote_min_cents in normalized args")
	}
}

func TestNormalizeCommitmentIntentRejectsInvalidClass(t *testing.T) {
	_, err := NormalizeCommitmentIntent(schemavoice.CommitmentIntent{
		SchemaID:        commitmentIntentSchemaID,
		SchemaVersion:   commitmentIntentSchemaV1,
		CreatedAt:       time.Date(2026, time.February, 15, 0, 0, 0, 0, time.UTC),
		ProducerVersion: "test",
		CallID:          "call_demo",
		TurnIndex:       0,
		CallSeq:         1,
		CommitmentClass: "unknown",
		Context: schemavoice.CommitmentContext{
			Identity:  "agent.voice",
			Workspace: "/srv/voice",
			RiskClass: "high",
		},
	})
	if err == nil {
		t.Fatalf("expected invalid commitment_class error")
	}
}

func TestMintAndValidateSayToken(t *testing.T) {
	keyPair, err := sign.GenerateKeyPair()
	if err != nil {
		t.Fatalf("generate key pair: %v", err)
	}
	workDir := t.TempDir()
	tokenPath := filepath.Join(workDir, "say_token.json")
	now := time.Date(2026, time.February, 15, 0, 0, 0, 0, time.UTC)
	result, err := MintSayToken(MintSayTokenOptions{
		ProducerVersion:    "test",
		CommitmentClass:    "refund",
		IntentDigest:       "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		PolicyDigest:       "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		CallID:             "call_demo",
		TurnIndex:          3,
		CallSeq:            8,
		RefundCeilingCents: 5000,
		TTL:                30 * time.Minute,
		Now:                now,
		SigningPrivateKey:  keyPair.Private,
		TokenPath:          tokenPath,
	})
	if err != nil {
		t.Fatalf("mint say token: %v", err)
	}
	if result.TokenPath != tokenPath {
		t.Fatalf("unexpected token path: %s", result.TokenPath)
	}
	parsed, err := ReadSayToken(tokenPath)
	if err != nil {
		t.Fatalf("read say token: %v", err)
	}
	if err := ValidateSayToken(parsed, keyPair.Public, SayTokenValidationOptions{
		Now:                     now.Add(5 * time.Minute),
		ExpectedIntentDigest:    result.Token.IntentDigest,
		ExpectedPolicyDigest:    result.Token.PolicyDigest,
		ExpectedCallID:          "call_demo",
		ExpectedTurnIndex:       3,
		ExpectedCallSeq:         8,
		ExpectedCommitmentClass: "refund",
	}); err != nil {
		t.Fatalf("validate say token: %v", err)
	}
}

func TestValidateSayTokenMismatchAndExpired(t *testing.T) {
	keyPair, err := sign.GenerateKeyPair()
	if err != nil {
		t.Fatalf("generate key pair: %v", err)
	}
	now := time.Date(2026, time.February, 15, 0, 0, 0, 0, time.UTC)
	result, err := MintSayToken(MintSayTokenOptions{
		ProducerVersion:   "test",
		CommitmentClass:   "quote",
		IntentDigest:      "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		PolicyDigest:      "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		CallID:            "call_demo",
		TurnIndex:         2,
		CallSeq:           4,
		QuoteMinCents:     1000,
		QuoteMaxCents:     1200,
		TTL:               1 * time.Minute,
		Now:               now,
		SigningPrivateKey: keyPair.Private,
	})
	if err != nil {
		t.Fatalf("mint say token: %v", err)
	}
	err = ValidateSayToken(result.Token, keyPair.Public, SayTokenValidationOptions{
		ExpectedIntentDigest: "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
	})
	if err == nil {
		t.Fatalf("expected intent mismatch error")
	}
	var tokenErr *SayTokenError
	if !errors.As(err, &tokenErr) || tokenErr.Code != SayTokenCodeIntentMismatch {
		t.Fatalf("unexpected token error: %v", err)
	}
	err = ValidateSayToken(result.Token, keyPair.Public, SayTokenValidationOptions{
		Now:               now.Add(2 * time.Minute),
		ExpectedTurnIndex: -1,
	})
	if err == nil {
		t.Fatalf("expected expired token error")
	}
	if !errors.As(err, &tokenErr) || tokenErr.Code != SayTokenCodeExpired {
		t.Fatalf("unexpected expired token error: %v", err)
	}
}

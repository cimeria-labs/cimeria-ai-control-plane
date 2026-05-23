package handler

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/multica-ai/multica/server/pkg/db/generated"
)

func TestSeedSDRAgentsDefinitions(t *testing.T) {
	if len(sdrAgentDefs) != 5 {
		t.Fatalf("expected 5 SDR agent definitions, got %d", len(sdrAgentDefs))
	}
	names := map[string]bool{}
	for _, def := range sdrAgentDefs {
		if def.Name == "" {
			t.Error("agent name should not be empty")
		}
		if def.Description == "" {
			t.Errorf("agent %q should have a description", def.Name)
		}
		names[def.Name] = true
	}
	for _, expected := range []string{"Hunter", "Qualificador", "Copywriter", "Closer", "Nurture"} {
		if !names[expected] {
			t.Errorf("missing SDR agent: %s", expected)
		}
	}
}

func TestDefaultScoreRules(t *testing.T) {
	if len(defaultScoreRules) != 7 {
		t.Fatalf("expected 7 default score rules, got %d", len(defaultScoreRules))
	}
	eventTypes := map[string]bool{}
	for _, rule := range defaultScoreRules {
		if rule.EventType == "" {
			t.Error("rule event_type should not be empty")
		}
		eventTypes[rule.EventType] = true
	}
	for _, expected := range []string{"opened", "clicked", "replied", "forwarded", "bounced", "complained", "unsubscribed"} {
		if !eventTypes[expected] {
			t.Errorf("missing score rule for event type: %s", expected)
		}
	}
}

func TestSeedSDRAgentsNoDB(t *testing.T) {
	// Verify the function exists and the definitions are correct.
	// Actual DB seeding is tested via integration tests.
	ctx := context.Background()
	// Create a mock qtx that will fail (expected), just verify no panic
	qtx := &db.Queries{}
	seedSDRAgents(ctx, qtx, pgtype.UUID{}, pgtype.UUID{})
	// Should not panic even with nil/empty queries
}

func TestSeedDefaultScoreRulesNoDB(t *testing.T) {
	ctx := context.Background()
	qtx := &db.Queries{}
	seedDefaultScoreRules(ctx, qtx, pgtype.UUID{})
	// Should not panic even with nil/empty queries
}
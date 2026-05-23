package sdr

import (
	"testing"

	"github.com/multica-ai/multica/server/internal/events"
	db "github.com/multica-ai/multica/server/pkg/db/generated"
	"github.com/multica-ai/multica/server/pkg/protocol"
)

func TestEngineRegistersListeners(t *testing.T) {
	bus := events.New()
	queries := &db.Queries{}
	_ = NewEngine(queries, bus)

	eventTypes := []string{
		protocol.EventLeadCreated,
		protocol.EventTaskCompleted,
		protocol.EventEmailBounced,
		protocol.EventEmailComplained,
	}
	for _, eventType := range eventTypes {
		t.Run(eventType, func(t *testing.T) {
			bus.Publish(events.Event{Type: eventType, WorkspaceID: "ws-test", Payload: map[string]any{}})
		})
	}
}
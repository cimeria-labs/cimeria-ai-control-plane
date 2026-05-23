package sdr

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/multica-ai/multica/server/internal/events"
	"github.com/multica-ai/multica/server/internal/util"
	db "github.com/multica-ai/multica/server/pkg/db/generated"
	"github.com/multica-ai/multica/server/pkg/protocol"
)

const defaultApprovalLimit = 20

// Engine is the SDR orchestrator. It listens to lead/email events and bridges
// them into the existing issue/agent task system. It does NOT execute logic
// itself - it creates issues assigned to SDR agents, and the LLM (via the
// daemon) does the real work.
type Engine struct {
	Queries *db.Queries
	Bus     *events.Bus
}

// NewEngine creates the SDR engine and wires its event subscriptions.
func NewEngine(queries *db.Queries, bus *events.Bus) *Engine {
	e := &Engine{Queries: queries, Bus: bus}
	e.register()
	return e
}

func (e *Engine) register() {
	// Lead created -> create an issue assigned to the Hunter agent
	e.Bus.Subscribe(protocol.EventLeadCreated, e.onLeadCreated)

	// Task completed -> check if it's an SDR issue and advance the pipeline
	e.Bus.Subscribe(protocol.EventTaskCompleted, e.onTaskCompleted)

	// Email bounced/complained -> cancel lead
	e.Bus.Subscribe(protocol.EventEmailBounced, e.onBounceOrComplaint)
	e.Bus.Subscribe(protocol.EventEmailComplained, e.onBounceOrComplaint)
}

// toMap converts an arbitrary value to map[string]any by round-tripping through
// JSON. This handles both map[string]any values and typed structs (like
// LeadResponse) that have json tags.
func toMap(v any) (map[string]any, bool) {
	if m, ok := v.(map[string]any); ok {
		return m, true
	}
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, false
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, false
	}
	return m, true
}

// onLeadCreated creates an issue assigned to the Hunter agent when a new lead
// enters the system. The existing issue/task pipeline handles the rest.
func (e *Engine) onLeadCreated(ev events.Event) {
	ctx := context.Background()
	payload, ok := ev.Payload.(map[string]any)
	if !ok {
		return
	}
	leadData, ok := toMap(payload["lead"])
	if !ok {
		slog.Warn("sdr: failed to parse lead data from event payload", "payload", payload)
		return
	}
	leadID, _ := leadData["id"].(string)
	workspaceID := ev.WorkspaceID
	if leadID == "" || workspaceID == "" {
		return
	}

	// Skip rejected leads - curator already disqualified them
	if status, _ := leadData["status"].(string); status == "rejected" {
		slog.Debug("sdr: skipping rejected lead", "lead_id", leadID)
		return
	}

	// Find the Hunter agent in this workspace
	agents, err := e.Queries.ListAgents(ctx, parseUUID(workspaceID))
	if err != nil || len(agents) == 0 {
		slog.Warn("sdr: no agents found in workspace", "workspace_id", workspaceID, "error", err)
		return
	}

	var hunterID pgtype.UUID
	for _, a := range agents {
		if a.Name == AgentHunter {
			hunterID = a.ID
			break
		}
	}
	if !hunterID.Valid {
		slog.Warn("sdr: Hunter agent not found in workspace", "workspace_id", workspaceID)
		return
	}

	// Create an issue assigned to Hunter, linked to the lead via origin
	leadEmail, _ := leadData["email"].(string)
	leadName, _ := leadData["name"].(string)
	company, _ := leadData["company"].(string)
	title := fmt.Sprintf("SDR: Qualify lead %s", leadEmail)
	description := fmt.Sprintf("New lead captured. Evaluate ICP fit and enrich data.\n\n**Lead:** %s (%s)\n**Company:** %s\n**Lead ID:** %s",
		leadName, leadEmail, company, leadID)

	issueNumber, err := e.Queries.IncrementIssueCounter(ctx, parseUUID(workspaceID))
	if err != nil {
		slog.Warn("sdr: failed to increment issue counter", "workspace_id", workspaceID, "error", err)
		return
	}

	issue, err := e.Queries.CreateIssueWithOrigin(ctx, db.CreateIssueWithOriginParams{
		WorkspaceID:  parseUUID(workspaceID),
		Title:        title,
		Description:  pgtype.Text{String: description, Valid: true},
		Status:       "todo",
		Priority:     "high",
		AssigneeType: pgtype.Text{String: "agent", Valid: true},
		AssigneeID:   hunterID,
		CreatorType:  "agent",
		CreatorID:    hunterID,
		Position:     0,
		Number:       issueNumber,
		OriginType:   pgtype.Text{String: "lead", Valid: true},
		OriginID:     parseUUID(leadID),
	})
	if err != nil {
		slog.Warn("sdr: FAILED_CREATE_HUNTER_ISSUE_V99", "lead_id", leadID, "error", err)
		return
	}

	slog.Info("sdr: created Hunter issue for lead", "issue_id", uuidStr(issue.ID), "lead_id", leadID, "workspace_id", workspaceID)

	// Publish issue:created so the task system picks it up
	e.Bus.Publish(events.Event{
		Type:        protocol.EventIssueCreated,
		WorkspaceID: workspaceID,
		ActorType:   "agent",
		ActorID:     "sdr-engine",
		Payload:     map[string]any{"issue": issueToMinimalMap(issue)},
	})
}

// onTaskCompleted advances the SDR pipeline when an agent finishes a task.
// If the completed task's issue has origin_type="lead", we advance the lead
// to the next pipeline stage.
func (e *Engine) onTaskCompleted(ev events.Event) {
	ctx := context.Background()
	payload, ok := ev.Payload.(map[string]any)
	if !ok {
		return
	}
	taskID, _ := payload["task_id"].(string)
	if taskID == "" {
		return
	}

	task, err := e.Queries.GetAgentTask(ctx, parseUUID(taskID))
	if err != nil {
		return
	}

	// Load the issue to check origin
	issue, err := e.Queries.GetIssue(ctx, task.IssueID)
	if err != nil {
		return
	}

	// Only handle SDR-originated issues
	if !issue.OriginType.Valid || issue.OriginType.String != "lead" {
		return
	}

	leadID := issue.OriginID
	if !leadID.Valid {
		return
	}

	// Find which agent completed - determine next step
	agent, err := e.Queries.GetAgent(ctx, task.AgentID)
	if err != nil {
		return
	}

	// Update lead state machine based on which agent finished
	workspaceID := uuidStr(issue.WorkspaceID)
	lead, err := e.Queries.GetLeadInWorkspace(ctx, db.GetLeadInWorkspaceParams{
		ID:          leadID,
		WorkspaceID: issue.WorkspaceID,
	})
	if err != nil {
		slog.Warn("sdr: task completed but lead not found", "lead_id", uuidStr(leadID), "error", err)
		return
	}

	nextAgentName := ""
	newStatus := ""
	switch agent.Name {
	case AgentHunter:
		nextAgentName = AgentQualificador
		newStatus = "qualified"
	case AgentQualificador:
		nextAgentName = AgentCopywriter
		newStatus = "copy_ready"
	case AgentCopywriter:
		nextAgentName = AgentCloser
		newStatus = "strategy_ready"
	case AgentCloser:
		nextAgentName = AgentNurture
		newStatus = "email_sent"
	case AgentNurture:
		newStatus = "nurturing"
	}

	// Update lead state
	if newStatus != "" {
		_, err := e.Queries.UpdateLead(ctx, db.UpdateLeadParams{
			ID:                 lead.ID,
			WorkspaceID:        lead.WorkspaceID,
			StateMachineStatus: pgtype.Text{String: newStatus, Valid: true},
			LastEvent:          pgtype.Text{String: "sdr." + agent.Name + ".completed", Valid: true},
		})
		if err != nil {
			slog.Warn("sdr: failed to update lead state", "lead_id", uuidStr(leadID), "error", err)
		}
		slog.Info("sdr: lead state advanced", "lead_id", uuidStr(leadID), "agent", agent.Name, "new_status", newStatus)
	}

	// Create issue for next agent if there is one
	if nextAgentName != "" {
		agents, err := e.Queries.ListAgents(ctx, issue.WorkspaceID)
		if err != nil {
			return
		}
		var nextAgentID pgtype.UUID
		for _, a := range agents {
			if a.Name == nextAgentName {
				nextAgentID = a.ID
				break
			}
		}
		if !nextAgentID.Valid {
			slog.Warn("sdr: next agent not found", "agent", nextAgentName, "workspace_id", workspaceID)
			return
		}

		nextTitle := fmt.Sprintf("SDR: %s for lead %s", nextAgentName, lead.Email)
		nextDesc := fmt.Sprintf("Lead %s advanced from %s. Pipeline stage: %s.\n**Lead ID:** %s",
			lead.Email, agent.Name, newStatus, uuidStr(leadID))

		issueNumber, err := e.Queries.IncrementIssueCounter(ctx, issue.WorkspaceID)
		if err != nil {
			return
		}

		nextIssue, err := e.Queries.CreateIssueWithOrigin(ctx, db.CreateIssueWithOriginParams{
			WorkspaceID:  issue.WorkspaceID,
			Title:        nextTitle,
			Description:  pgtype.Text{String: nextDesc, Valid: true},
			Status:       "todo",
			Priority:     "high",
			AssigneeType: pgtype.Text{String: "agent", Valid: true},
			AssigneeID:   nextAgentID,
			CreatorType:  "agent",
			CreatorID:    nextAgentID,
			Position:     0,
			Number:       issueNumber,
			OriginType:   pgtype.Text{String: "lead", Valid: true},
			OriginID:     leadID,
		})
		if err != nil {
			slog.Warn("sdr: failed to create next pipeline issue", "agent", nextAgentName, "error", err)
			return
		}

		slog.Info("sdr: created next pipeline issue", "agent", nextAgentName, "issue_id", uuidStr(nextIssue.ID), "lead_id", uuidStr(leadID))

		e.Bus.Publish(events.Event{
			Type:        protocol.EventIssueCreated,
			WorkspaceID: workspaceID,
			ActorType:   "agent",
			ActorID:     "sdr-engine",
			Payload:     map[string]any{"issue": issueToMinimalMap(nextIssue)},
		})
	}
}

// onBounceOrComplaint cancels the lead when email delivery fails permanently.
func (e *Engine) onBounceOrComplaint(ev events.Event) {
	ctx := context.Background()
	payload, ok := ev.Payload.(map[string]any)
	if !ok {
		return
	}
	leadID, _ := payload["lead_id"].(string)
	workspaceID := ev.WorkspaceID
	if leadID == "" || workspaceID == "" {
		return
	}

	lead, err := e.Queries.GetLeadInWorkspace(ctx, db.GetLeadInWorkspaceParams{
		ID:          parseUUID(leadID),
		WorkspaceID: parseUUID(workspaceID),
	})
	if err != nil {
		return
	}

	if lead.Status == "cancelled" || lead.Status == "rejected" {
		return
	}

	eventType := "email.bounced"
	if ev.Type == protocol.EventEmailComplained {
		eventType = "email.complained"
	}

	_, err = e.Queries.UpdateLead(ctx, db.UpdateLeadParams{
		ID:                 lead.ID,
		WorkspaceID:        lead.WorkspaceID,
		Status:             pgtype.Text{String: "cancelled", Valid: true},
		StateMachineStatus: pgtype.Text{String: "cancelled", Valid: true},
		LastEvent:          pgtype.Text{String: eventType, Valid: true},
	})
	if err != nil {
		slog.Warn("sdr: failed to cancel lead after bounce/complaint", "lead_id", leadID, "error", err)
		return
	}

	slog.Info("sdr: lead cancelled after bounce/complaint", "lead_id", leadID, "event", ev.Type)
	e.Bus.Publish(events.Event{
		Type:        protocol.EventLeadRejected,
		WorkspaceID: workspaceID,
		ActorType:   "agent",
		ActorID:     "nurture",
		Payload:     map[string]any{"lead_id": leadID, "reason": eventType},
	})
}

// issueToMinimalMap creates a minimal map for event payloads.
func issueToMinimalMap(issue db.Issue) map[string]any {
	return map[string]any{
		"id":            uuidStr(issue.ID),
		"workspace_id":  uuidStr(issue.WorkspaceID),
		"title":         issue.Title,
		"status":        issue.Status,
		"priority":      issue.Priority,
		"assignee_type": textStr(issue.AssigneeType),
		"assignee_id":   uuidStr(issue.AssigneeID),
		"origin_type":   textStr(issue.OriginType),
		"origin_id":     uuidStr(issue.OriginID),
	}
}

func textStr(t pgtype.Text) string {
	if !t.Valid {
		return ""
	}
	return t.String
}

func parseUUID(s string) pgtype.UUID {
	return util.ParseUUID(s)
}

func uuidStr(u pgtype.UUID) string {
	return util.UUIDToString(u)
}

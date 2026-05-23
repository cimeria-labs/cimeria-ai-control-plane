package handler

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/multica-ai/multica/server/pkg/db/generated"
	"github.com/multica-ai/multica/server/pkg/protocol"
)

// 1x1 transparent GIF (43 bytes)
var trackingPixel = []byte{
	0x47, 0x49, 0x46, 0x38, 0x39, 0x61, 0x01, 0x00, 0x01, 0x00,
	0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff, 0xff, 0xff, 0x21,
	0xf9, 0x04, 0x01, 0x00, 0x00, 0x00, 0x00, 0x2c, 0x00, 0x00,
	0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x02, 0x02, 0x44,
	0x01, 0x00, 0x3b,
}

// clientIP extracts the client IP from the request, respecting X-Forwarded-For
// when present. Returns nil if parsing fails.
func clientIP(r *http.Request) *netip.Addr {
	ipStr := strings.TrimSpace(r.Header.Get("X-Forwarded-For"))
	if ipStr != "" {
		// X-Forwarded-For may contain multiple IPs; use the first.
		if idx := strings.Index(ipStr, ","); idx != -1 {
			ipStr = strings.TrimSpace(ipStr[:idx])
		}
	} else {
		// RemoteAddr is host:port; strip the port.
		ipStr = r.RemoteAddr
		if idx := strings.LastIndex(ipStr, ":"); idx != -1 {
			ipStr = ipStr[:idx]
		}
	}
	addr, err := netip.ParseAddr(ipStr)
	if err != nil {
		return nil
	}
	return &addr
}

// TrackPixel handles the open-tracking pixel. It records an "opened" event
// and returns a 1x1 transparent GIF. The pixel URL is invisible to users.
func (h *Handler) TrackPixel(w http.ResponseWriter, r *http.Request) {
	logID := chi.URLParam(r, "id")
	if logID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	logUUID := parseUUID(logID)
	if !logUUID.Valid {
		writeError(w, http.StatusBadRequest, "invalid email log ID")
		return
	}

	logEntry, err := h.Queries.GetEmailLog(ctx, logUUID)
	if err != nil {
		slog.Warn("tracking pixel: email log not found", "email_log_id", logID, "error", err)
		writeTrackingPixel(w)
		return
	}

	// Record the open event.
	event, err := h.Queries.CreateEmailEvent(ctx, db.CreateEmailEventParams{
		EmailLogID: logUUID,
		EventType:  "opened",
		IpAddress:  clientIP(r),
		UserAgent:  pgtype.Text{String: r.UserAgent(), Valid: true},
	})
	if err != nil {
		slog.Warn("failed to create email open event", "email_log_id", logID, "error", err)
	} else {
		h.afterEmailEvent(ctx, logEntry, event, protocol.EventEmailOpened)
	}

	// Update email_log opened_at if first open.
	if _, err := h.Queries.UpdateEmailLogStatus(ctx, db.UpdateEmailLogStatusParams{
		ID:       logUUID,
		OpenedAt: pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
	}); err != nil {
		slog.Warn("failed to update email log opened_at", "email_log_id", logID, "error", err)
	}

	writeTrackingPixel(w)
}

// TrackClick handles click-tracking redirects. It records a "clicked" event
// and 302-redirects to the original URL.
func (h *Handler) TrackClick(w http.ResponseWriter, r *http.Request) {
	logID := chi.URLParam(r, "id")
	encodedURL := r.URL.Query().Get("url")
	if logID == "" || encodedURL == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	originalURL, err := base64.URLEncoding.DecodeString(encodedURL)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	parsedURL, err := url.Parse(string(originalURL))
	if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	logUUID := parseUUID(logID)
	if !logUUID.Valid {
		writeError(w, http.StatusBadRequest, "invalid email log ID")
		return
	}

	logEntry, err := h.Queries.GetEmailLog(ctx, logUUID)
	if err != nil {
		slog.Warn("tracking click: email log not found", "email_log_id", logID, "error", err)
		w.Header().Set("Location", string(originalURL))
		w.WriteHeader(http.StatusFound)
		return
	}

	// Record the click event.
	event, err := h.Queries.CreateEmailEvent(ctx, db.CreateEmailEventParams{
		EmailLogID: logUUID,
		EventType:  "clicked",
		IpAddress:  clientIP(r),
		UserAgent:  pgtype.Text{String: r.UserAgent(), Valid: true},
		LinkUrl:    pgtype.Text{String: string(originalURL), Valid: true},
	})
	if err != nil {
		slog.Warn("failed to create email click event", "email_log_id", logID, "error", err)
	} else {
		h.afterEmailEvent(ctx, logEntry, event, protocol.EventEmailClicked)
	}

	// Update email_log clicked_at if first click.
	if _, err := h.Queries.UpdateEmailLogStatus(ctx, db.UpdateEmailLogStatusParams{
		ID:        logUUID,
		ClickedAt: pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
	}); err != nil {
		slog.Warn("failed to update email log clicked_at", "email_log_id", logID, "error", err)
	}

	w.Header().Set("Location", string(originalURL))
	w.WriteHeader(http.StatusFound)
}

// ResendWebhookPayload represents the JSON body sent by Resend webhooks.
type ResendWebhookPayload struct {
	Type string `json:"type"`
	Data struct {
		ID      string `json:"id"`
		To      string `json:"to"`
		Subject string `json:"subject"`
		// Bounce-specific fields
		Bounce struct {
			Type   string `json:"type"`
			Reason string `json:"reason"`
		} `json:"bounce"`
	} `json:"data"`
}

var errWebhookSecretMissing = errors.New("RESEND_WEBHOOK_SECRET is not configured")

// HandleResendWebhook receives Resend webhook events and updates email tracking state.
// Open and click events are tracked exclusively by TrackPixel and TrackClick;
// this handler only processes delivery-status events (delivered, bounced, complained).
func (h *Handler) HandleResendWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20)) // 1 MB max
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := verifyResendWebhookSignature(r, body); err != nil {
		if errors.Is(err, errWebhookSecretMissing) {
			writeError(w, http.StatusServiceUnavailable, "resend webhook secret is not configured")
			return
		}
		writeError(w, http.StatusUnauthorized, "invalid webhook signature")
		return
	}

	var payload ResendWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Find the email_log by Resend ID.
	logEntry, err := h.Queries.GetEmailLogByResendID(ctx, pgtype.Text{String: payload.Data.ID, Valid: true})
	if err != nil {
		slog.Warn("resend webhook: email_log not found for resend_id", "resend_id", payload.Data.ID, "error", err)
		w.WriteHeader(http.StatusOK)
		return
	}

	logUUID := logEntry.ID

	switch payload.Type {
	case "email.delivered":
		if _, err := h.Queries.UpdateEmailLogStatus(ctx, db.UpdateEmailLogStatusParams{
			ID:     logUUID,
			Status: pgtype.Text{String: "delivered", Valid: true},
		}); err != nil {
			slog.Warn("failed to update email log status to delivered", "email_log_id", uuidToString(logUUID), "error", err)
		}
		event, err := h.Queries.CreateEmailEvent(ctx, db.CreateEmailEventParams{
			EmailLogID: logUUID,
			EventType:  "delivered",
			UserAgent:  pgtype.Text{String: r.UserAgent(), Valid: true},
		})
		if err != nil {
			slog.Warn("failed to create email delivered event", "email_log_id", uuidToString(logUUID), "error", err)
		} else {
			h.afterEmailEvent(ctx, logEntry, event, protocol.EventEmailDelivered)
		}

	case "email.bounced":
		if _, err := h.Queries.UpdateEmailLogStatus(ctx, db.UpdateEmailLogStatusParams{
			ID:           logUUID,
			Status:       pgtype.Text{String: "bounced", Valid: true},
			BounceReason: pgtype.Text{String: payload.Data.Bounce.Reason, Valid: true},
		}); err != nil {
			slog.Warn("failed to update email log status to bounced", "email_log_id", uuidToString(logUUID), "error", err)
		}
		event, err := h.Queries.CreateEmailEvent(ctx, db.CreateEmailEventParams{
			EmailLogID: logUUID,
			EventType:  "bounced",
			UserAgent:  pgtype.Text{String: r.UserAgent(), Valid: true},
		})
		if err != nil {
			slog.Warn("failed to create email bounced event", "email_log_id", uuidToString(logUUID), "error", err)
		} else {
			h.afterEmailEvent(ctx, logEntry, event, protocol.EventEmailBounced)
		}

	case "email.complained":
		if _, err := h.Queries.UpdateEmailLogStatus(ctx, db.UpdateEmailLogStatusParams{
			ID:     logUUID,
			Status: pgtype.Text{String: "suppressed", Valid: true},
		}); err != nil {
			slog.Warn("failed to update email log status to suppressed", "email_log_id", uuidToString(logUUID), "error", err)
		}
		event, err := h.Queries.CreateEmailEvent(ctx, db.CreateEmailEventParams{
			EmailLogID: logUUID,
			EventType:  "complained",
			UserAgent:  pgtype.Text{String: r.UserAgent(), Valid: true},
		})
		if err != nil {
			slog.Warn("failed to create email complained event", "email_log_id", uuidToString(logUUID), "error", err)
		} else {
			h.afterEmailEvent(ctx, logEntry, event, protocol.EventEmailComplained)
		}

	default:
		slog.Debug("resend webhook: unhandled event type", "type", payload.Type, "resend_id", payload.Data.ID)
	}

	w.WriteHeader(http.StatusOK)
}

func writeTrackingPixel(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "image/gif")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	w.WriteHeader(http.StatusOK)
	w.Write(trackingPixel)
}

func verifyResendWebhookSignature(r *http.Request, body []byte) error {
	secret := strings.TrimSpace(os.Getenv("RESEND_WEBHOOK_SECRET"))
	if secret == "" {
		return errWebhookSecretMissing
	}
	msgID := r.Header.Get("svix-id")
	timestamp := r.Header.Get("svix-timestamp")
	signature := r.Header.Get("svix-signature")
	if msgID == "" || timestamp == "" || signature == "" {
		return errors.New("missing svix signature headers")
	}
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return err
	}
	if diff := time.Since(time.Unix(ts, 0)); diff > 5*time.Minute || diff < -5*time.Minute {
		return errors.New("stale svix timestamp")
	}
	secretBytes, err := decodeSvixSecret(secret)
	if err != nil {
		return err
	}
	signed := bytes.Join([][]byte{[]byte(msgID), []byte(timestamp), body}, []byte("."))
	mac := hmac.New(sha256.New, secretBytes)
	mac.Write(signed)
	expected := mac.Sum(nil)
	for _, candidate := range strings.Fields(signature) {
		parts := strings.SplitN(candidate, ",", 2)
		if len(parts) != 2 || parts[0] != "v1" {
			continue
		}
		got, err := base64.StdEncoding.DecodeString(parts[1])
		if err == nil && hmac.Equal(got, expected) {
			return nil
		}
	}
	return errors.New("signature mismatch")
}

func decodeSvixSecret(secret string) ([]byte, error) {
	secret = strings.TrimPrefix(secret, "whsec_")
	if decoded, err := base64.StdEncoding.DecodeString(secret); err == nil {
		return decoded, nil
	}
	return []byte(secret), nil
}

func (h *Handler) afterEmailEvent(ctx context.Context, logEntry db.EmailLog, event db.EmailEvent, eventType string) {
	workspaceID := uuidToString(logEntry.WorkspaceID)
	payload := map[string]any{
		"email_log_id": uuidToString(logEntry.ID),
		"event_id":     uuidToString(event.ID),
		"event_type":   event.EventType,
		"to":           logEntry.To,
		"subject":      logEntry.Subject,
	}
	if logEntry.IssueID.Valid {
		payload["issue_id"] = uuidToString(logEntry.IssueID)
	}
	if logEntry.LeadID.Valid {
		payload["lead_id"] = uuidToString(logEntry.LeadID)
	}
	if event.LinkUrl.Valid {
		payload["link_url"] = event.LinkUrl.String
	}
	h.publish(eventType, workspaceID, "system", "", payload)

	if !logEntry.LeadID.Valid {
		return
	}
	h.publish(protocol.EventLeadInteraction, workspaceID, "system", "", payload)

	score, err := h.Queries.CalculateLeadDynamicScore(ctx, db.CalculateLeadDynamicScoreParams{
		LeadID:      logEntry.LeadID,
		WorkspaceID: logEntry.WorkspaceID,
	})
	if err != nil {
		slog.Warn("failed to calculate lead dynamic score", "lead_id", uuidToString(logEntry.LeadID), "error", err)
		return
	}
	previous, _ := h.Queries.GetLeadInWorkspace(ctx, db.GetLeadInWorkspaceParams{
		ID:          logEntry.LeadID,
		WorkspaceID: logEntry.WorkspaceID,
	})
	lead, err := h.Queries.UpdateLeadDynamicScore(ctx, db.UpdateLeadDynamicScoreParams{
		ID:           logEntry.LeadID,
		WorkspaceID:  logEntry.WorkspaceID,
		DynamicScore: score,
		LastEvent:    pgtype.Text{String: event.EventType, Valid: true},
	})
	if err != nil {
		slog.Warn("failed to update lead dynamic score", "lead_id", uuidToString(logEntry.LeadID), "error", err)
		return
	}
	leadPayload := map[string]any{"lead": leadToResponse(lead), "email_event": payload}
	if score >= 7 && previous.Status != "hot" {
		h.publish(protocol.EventLeadHot, workspaceID, "system", "", leadPayload)
		h.createLeadHotInbox(ctx, lead, payload)
	}
}

func (h *Handler) createLeadHotInbox(ctx context.Context, lead db.Lead, eventPayload map[string]any) {
	members, err := h.Queries.ListMembers(ctx, lead.WorkspaceID)
	if err != nil || len(members) == 0 {
		if err != nil {
			slog.Warn("failed to list members for hot lead inbox", "lead_id", uuidToString(lead.ID), "error", err)
		}
		return
	}
	recipient := members[0]
	for _, member := range members {
		if member.Role == "owner" || member.Role == "admin" {
			recipient = member
			break
		}
	}
	details, _ := json.Marshal(map[string]any{
		"lead_id":     uuidToString(lead.ID),
		"lead_email":  lead.Email,
		"lead_name":   lead.Name,
		"company":     lead.Company,
		"score":       lead.Score,
		"dynamic":     lead.DynamicScore,
		"email_event": eventPayload,
	})
	item, err := h.Queries.CreateInboxItem(ctx, db.CreateInboxItemParams{
		WorkspaceID:   lead.WorkspaceID,
		RecipientType: "member",
		RecipientID:   recipient.UserID,
		Type:          protocol.EventLeadHot,
		Severity:      "action_required",
		Title:         "Hot lead ready for handoff",
		Body:          pgtype.Text{String: lead.Email + " crossed the hot-lead threshold.", Valid: true},
		ActorType:     pgtype.Text{String: "system", Valid: true},
		Details:       details,
	})
	if err != nil {
		slog.Warn("failed to create hot lead inbox item", "lead_id", uuidToString(lead.ID), "error", err)
		return
	}
	h.publish(protocol.EventInboxNew, uuidToString(lead.WorkspaceID), "system", "", map[string]any{
		"item": inboxToResponse(item),
	})
}

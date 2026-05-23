package handler

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/mail"
	"net/url"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/multica-ai/multica/server/pkg/db/generated"
)

type SendEmailRequest struct {
	To              []string `json:"to"`
	Subject         string   `json:"subject"`
	Body            string   `json:"body"`
	HTMLBody        string   `json:"html_body"`
	IssueID         *string  `json:"issue_id"`
	LeadID          *string  `json:"lead_id"`
	TrackingEnabled *bool    `json:"tracking_enabled"`
}

type SendEmailResponse struct {
	ID      string   `json:"id"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Status  string   `json:"status"`
}

const maxRecipients = 10

func (h *Handler) SendEmail(w http.ResponseWriter, r *http.Request) {
	workspaceID := h.resolveWorkspaceID(r)
	if workspaceID == "" {
		writeError(w, http.StatusBadRequest, "workspace_id is required")
		return
	}

	// Verify workspace membership (agents are also allowed via X-Agent-ID header).
	if _, ok := h.workspaceMember(w, r, workspaceID); !ok {
		return
	}

	var req SendEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate recipients.
	if len(req.To) == 0 {
		writeError(w, http.StatusBadRequest, "to is required")
		return
	}
	if len(req.To) > maxRecipients {
		writeError(w, http.StatusBadRequest, "too many recipients (max 10)")
		return
	}
	for _, addr := range req.To {
		if _, err := mail.ParseAddress(addr); err != nil {
			writeError(w, http.StatusBadRequest, "invalid email address: "+addr)
			return
		}
	}

	// Validate subject.
	req.Subject = strings.TrimSpace(req.Subject)
	if req.Subject == "" {
		writeError(w, http.StatusBadRequest, "subject is required")
		return
	}

	// Validate body: at least one of body or html_body must be provided.
	if req.Body == "" && req.HTMLBody == "" {
		writeError(w, http.StatusBadRequest, "body or html_body is required")
		return
	}

	// Resolve actor for audit logging.
	userID, _ := requireUserID(w, r)
	if userID == "" {
		return
	}
	actorType, actorID := h.resolveActor(r, userID, workspaceID)
	trackingEnabled := true
	if req.TrackingEnabled != nil {
		trackingEnabled = *req.TrackingEnabled
	}

	// For each recipient, create email_log, inject tracking, send.
	var firstResendID string
	ctx := r.Context()
	for _, to := range req.To {
		// 1. Create email_log entry.
		bodyPreview := req.Body
		if bodyPreview == "" {
			bodyPreview = req.HTMLBody
		}
		if len(bodyPreview) > 200 {
			bodyPreview = bodyPreview[:200]
		}

		logEntry, err := h.Queries.CreateEmailLog(ctx, db.CreateEmailLogParams{
			WorkspaceID:     parseUUID(workspaceID),
			IssueID:         optionalUUID(req.IssueID),
			LeadID:          optionalUUID(req.LeadID),
			SenderID:        parseUUID(actorID),
			SenderType:      pgtype.Text{String: actorType, Valid: true},
			To:              to,
			Subject:         req.Subject,
			BodyPreview:     pgtype.Text{String: bodyPreview, Valid: true},
			EmailType:       "generic",
			Status:          pgtype.Text{String: "sent", Valid: true},
			TrackingEnabled: pgtype.Bool{Bool: trackingEnabled, Valid: true},
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to create email log: "+err.Error())
			return
		}

		// 2. Inject tracking pixel and click links into HTML body.
		// Use BACKEND_ORIGIN so tracking URLs hit the API server, not the frontend.
		baseURL := os.Getenv("BACKEND_ORIGIN")
		if baseURL == "" {
			baseURL = "http://localhost:8080"
		}
		htmlBody := req.HTMLBody
		if trackingEnabled && htmlBody != "" {
			htmlBody = injectTracking(htmlBody, uuidToString(logEntry.ID), baseURL)
		}

		// 3. Send via Resend.
		resendID, err := h.EmailService.SendGenericEmail([]string{to}, req.Subject, req.Body, htmlBody)
		if err != nil {
			_, _ = h.Queries.UpdateEmailLogStatus(ctx, db.UpdateEmailLogStatusParams{
				ID:     logEntry.ID,
				Status: pgtype.Text{String: "failed", Valid: true},
			})
			writeError(w, http.StatusInternalServerError, "failed to send email: "+err.Error())
			return
		}

		// 4. Update email_log with resend_id.
		_, _ = h.Queries.UpdateEmailLogStatus(ctx, db.UpdateEmailLogStatusParams{
			ID:       logEntry.ID,
			ResendID: pgtype.Text{String: resendID, Valid: true},
		})

		if firstResendID == "" {
			firstResendID = resendID
		}
	}

	status := "sent"
	if firstResendID == "dev-mode" {
		status = "dev-mode"
	}

	writeJSON(w, http.StatusOK, SendEmailResponse{
		ID:      firstResendID,
		To:      req.To,
		Subject: req.Subject,
		Status:  status,
	})
}

// injectTracking adds an open-tracking pixel and replaces <a href="..."> links
// with click-tracking redirects. baseURL should be the public origin (e.g. https://app.multica.ai).
func injectTracking(htmlBody, emailLogID, baseURL string) string {
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	// 1. Append 1x1 tracking pixel at the end of the body.
	pixelURL := baseURL + "/track/pixel/" + emailLogID
	pixelTag := `<img src="` + pixelURL + `" width="1" height="1" alt="" style="display:block;" />`
	htmlBody = strings.TrimSuffix(htmlBody, "</body>")
	if !strings.Contains(htmlBody, "</body>") {
		htmlBody += pixelTag
	} else {
		htmlBody = strings.Replace(htmlBody, "</body>", pixelTag+"</body>", 1)
	}

	// 2. Replace all <a href="..."> links with click-tracking redirects.
	// Simple regex-free replacement using strings.Index for safety.
	var result strings.Builder
	remaining := htmlBody
	for {
		idx := strings.Index(remaining, `href="`)
		if idx == -1 {
			result.WriteString(remaining)
			break
		}
		quoteStart := idx + len(`href="`)
		quoteEnd := strings.Index(remaining[quoteStart:], `"`)
		if quoteEnd == -1 {
			result.WriteString(remaining)
			break
		}
		quoteEnd += quoteStart

		originalURL := remaining[quoteStart:quoteEnd]
		// Skip already-tracking URLs, anchors, and mailto links.
		if strings.HasPrefix(originalURL, baseURL+"/track/click/") ||
			strings.HasPrefix(originalURL, "#") ||
			strings.HasPrefix(originalURL, "mailto:") {
			result.WriteString(remaining[:quoteEnd+1])
			remaining = remaining[quoteEnd+1:]
			continue
		}

		encoded := base64.URLEncoding.EncodeToString([]byte(originalURL))
		trackedURL := baseURL + "/track/click/" + emailLogID + "?url=" + url.QueryEscape(encoded)

		result.WriteString(remaining[:quoteStart])
		result.WriteString(trackedURL)
		result.WriteString("\"")
		remaining = remaining[quoteEnd+1:]
	}

	return result.String()
}

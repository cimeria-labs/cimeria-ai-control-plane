package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/multica-ai/multica/server/internal/cli"
)

var emailCmd = &cobra.Command{
	Use:   "email",
	Short: "Send emails from the workspace",
}

var emailSendCmd = &cobra.Command{
	Use:   "send",
	Short: "Send an email",
	Long:  "Send an email to one or more recipients from the workspace.",
	Example: `  # Send a plain-text email
  $ multica email send --to "lead@example.com" --subject "Follow-up" --body "Hello, just checking in."

  # Send an HTML email
  $ multica email send --to "lead@example.com" --subject "Proposal" --html "<h1>Hi</h1><p>Here is our proposal.</p>"

  # Pipe body from stdin
  $ echo "Long email body..." | multica email send --to "lead@example.com" --subject "Test" --body-stdin

  # Multiple recipients
  $ multica email send --to "a@x.com,b@x.com" --subject "Update" --body "Hi all"`,
	RunE: runEmailSend,
}

func init() {
	emailCmd.AddCommand(emailSendCmd)

	emailSendCmd.Flags().String("to", "", "Recipient email(s), comma-separated (required)")
	emailSendCmd.Flags().String("subject", "", "Email subject line (required)")
	emailSendCmd.Flags().String("body", "", "Plain-text email body")
	emailSendCmd.Flags().String("html", "", "HTML email body")
	emailSendCmd.Flags().Bool("body-stdin", false, "Read plain-text body from stdin")
	emailSendCmd.Flags().Bool("html-stdin", false, "Read HTML body from stdin")
	emailSendCmd.Flags().String("output", "json", "Output format: json or table")
}

func runEmailSend(cmd *cobra.Command, args []string) error {
	toFlag, _ := cmd.Flags().GetString("to")
	if toFlag == "" {
		return fmt.Errorf("--to is required")
	}
	subject, _ := cmd.Flags().GetString("subject")
	if subject == "" {
		return fmt.Errorf("--subject is required")
	}

	// Parse comma-separated recipients.
	recipients := splitAndTrim(toFlag)
	if len(recipients) == 0 {
		return fmt.Errorf("--to must contain at least one email address")
	}

	body, _ := cmd.Flags().GetString("body")
	html, _ := cmd.Flags().GetString("html")
	bodyStdin, _ := cmd.Flags().GetBool("body-stdin")
	htmlStdin, _ := cmd.Flags().GetBool("html-stdin")

	if bodyStdin {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("read stdin: %w", err)
		}
		body = string(data)
	}
	if htmlStdin {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("read stdin: %w", err)
		}
		html = string(data)
	}

	if body == "" && html == "" {
		return fmt.Errorf("--body, --html, --body-stdin, or --html-stdin is required")
	}

	client, err := newAPIClient(cmd)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	reqBody := map[string]any{
		"to":       recipients,
		"subject":  subject,
		"body":     body,
		"html_body": html,
	}

	var result map[string]any
	if err := client.PostJSON(ctx, "/api/emails/send", reqBody, &result); err != nil {
		return fmt.Errorf("send email: %w", err)
	}

	output, _ := cmd.Flags().GetString("output")
	switch output {
	case "table":
		headers := []string{"ID", "Status", "Subject", "To"}
		rows := [][]string{
			{
				strVal(result, "id"),
				strVal(result, "status"),
				strVal(result, "subject"),
				strings.Join(recipients, ", "),
			},
		}
		cli.PrintTable(os.Stdout, headers, rows)
		return nil
	default:
		return cli.PrintJSON(os.Stdout, result)
	}
}

// splitAndTrim splits a comma-separated string and trims whitespace from each element.
func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
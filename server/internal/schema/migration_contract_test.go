package schema

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func readMigrationSet(t *testing.T) string {
	t.Helper()

	candidates := []string{
		filepath.FromSlash("../../migrations"),
		filepath.FromSlash("server/migrations"),
	}

	var migrationsDir string
	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			migrationsDir = candidate
			break
		}
	}
	if migrationsDir == "" {
		t.Fatalf("migrations directory not found from %s", mustGetwd(t))
	}

	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		t.Fatalf("read migrations directory: %v", err)
	}

	var builder strings.Builder
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".up.sql") {
			continue
		}
		content, err := os.ReadFile(filepath.Join(migrationsDir, name))
		if err != nil {
			t.Fatalf("read migration %s: %v", name, err)
		}
		builder.WriteString("\n-- ")
		builder.WriteString(name)
		builder.WriteByte('\n')
		builder.Write(content)
		builder.WriteByte('\n')
	}

	return strings.ToLower(builder.String())
}

func mustGetwd(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	return wd
}

func requireSQL(t *testing.T, sql string, required ...string) {
	t.Helper()
	for _, needle := range required {
		if !strings.Contains(sql, strings.ToLower(needle)) {
			t.Fatalf("expected migration SQL to contain %q", needle)
		}
	}
}

func TestLeadReadinessMigrationsContainGeneratedCodeSchema(t *testing.T) {
	sql := readMigrationSet(t)

	requireSQL(t, sql,
		"ADD COLUMN IF NOT EXISTS budget",
		"ADD COLUMN IF NOT EXISTS authority",
		"ADD COLUMN IF NOT EXISTS need",
		"ADD COLUMN IF NOT EXISTS timeline",
		"ADD COLUMN IF NOT EXISTS company_size",
		"ADD COLUMN IF NOT EXISTS industry",
		"ADD COLUMN IF NOT EXISTS pain_points",
		"ADD COLUMN IF NOT EXISTS icp_fit",
		"ADD COLUMN IF NOT EXISTS lead_temperature",
		"ADD COLUMN IF NOT EXISTS curated_at",
		"ADD COLUMN IF NOT EXISTS curated_by",
		"ADD COLUMN IF NOT EXISTS import_batch_id",
		"CREATE TABLE IF NOT EXISTS lead_source",
		"CREATE TABLE IF NOT EXISTS lead_import_batch",
		"CREATE TABLE IF NOT EXISTS lead_curator_rule",
		"provider TEXT NOT NULL",
		"auto_approve BOOLEAN NOT NULL DEFAULT false",
		"enrichment_enabled BOOLEAN NOT NULL DEFAULT true",
		"metadata JSONB NOT NULL DEFAULT '{}'",
		"DROP CONSTRAINT IF EXISTS lead_import_batch_provider_check",
		"DROP CONSTRAINT IF EXISTS lead_import_batch_status_check",
		"'apollo'",
		"'preview'",
	)
}

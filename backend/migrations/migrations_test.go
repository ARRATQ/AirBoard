package migrations

import (
	"fmt"
	"io/fs"
	"regexp"
	"sort"
	"strings"
	"testing"
)

var fileName = regexp.MustCompile(`^(\d{5})_[a-z0-9_]+\.sql$`)

func sqlFiles(t *testing.T) []string {
	t.Helper()
	names, err := fs.Glob(FS, "*.sql")
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	if len(names) == 0 {
		t.Fatal("aucune migration embarquée")
	}
	sort.Strings(names)
	return names
}

func TestMigrationsAreNumberedWithoutGaps(t *testing.T) {
	for i, name := range sqlFiles(t) {
		m := fileName.FindStringSubmatch(name)
		if m == nil {
			t.Errorf("%s : nom attendu NNNNN_description.sql (minuscules, chiffres, _)", name)
			continue
		}
		if want := fmt.Sprintf("%05d", i+1); m[1] != want {
			t.Errorf("%s : numéro %s attendu (pas de trou ni de doublon)", name, want)
		}
	}
}

func TestMigrationsAreWellFormed(t *testing.T) {
	for _, name := range sqlFiles(t) {
		t.Run(name, func(t *testing.T) {
			raw, err := fs.ReadFile(FS, name)
			if err != nil {
				t.Fatalf("lecture: %v", err)
			}
			body := string(raw)
			checks := []struct {
				desc string
				ok   bool
			}{
				{"une annotation -- +goose Up", strings.Count(body, "-- +goose Up") == 1},
				{"une annotation -- +goose Down", strings.Count(body, "-- +goose Down") == 1},
				{"Up avant Down", strings.Index(body, "-- +goose Up") < strings.Index(body, "-- +goose Down")},
				{"StatementBegin/End appariés",
					strings.Count(body, "-- +goose StatementBegin") == strings.Count(body, "-- +goose StatementEnd")},
				{"pas de NO TRANSACTION (chaque migration doit être atomique)",
					!strings.Contains(body, "+goose NO TRANSACTION")},
			}
			for _, c := range checks {
				if !c.ok {
					t.Errorf("attendu : %s", c.desc)
				}
			}
		})
	}
}

// Les migrations doivent pouvoir s'appliquer sur une base de production construite
// par l'ancien AutoMigrate : toute création ou suppression directe doit être
// conditionnelle (les EXECUTE dans les fonctions pg_temp sont gardés par des tests).
func TestMigrationsUpAreIdempotent(t *testing.T) {
	ddl := regexp.MustCompile(`(?i)^\s*(CREATE (UNIQUE )?INDEX|CREATE TABLE|CREATE SEQUENCE|DROP INDEX|DROP TABLE)\b|ADD COLUMN|DROP CONSTRAINT`)
	guard := regexp.MustCompile(`(?i)IF (NOT )?EXISTS`)
	for _, name := range sqlFiles(t) {
		raw, err := fs.ReadFile(FS, name)
		if err != nil {
			t.Fatalf("lecture %s: %v", name, err)
		}
		up := string(raw)
		if i := strings.Index(up, "-- +goose Down"); i >= 0 {
			up = up[:i]
		}
		for n, line := range strings.Split(up, "\n") {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "--") || strings.HasPrefix(trimmed, "EXECUTE ") {
				continue
			}
			if ddl.MatchString(line) && !guard.MatchString(line) {
				t.Errorf("%s:%d : %q sans IF [NOT] EXISTS", name, n+1, strings.TrimSpace(line))
			}
		}
	}
}

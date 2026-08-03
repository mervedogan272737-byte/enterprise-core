package database

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

func RunMigrations(
	ctx context.Context,
	db *pgxpool.Pool,
	migrationsDir string,
) error {
	if err := ensureMigrationsTable(ctx, db); err != nil {
		return err
	}

	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("read migrations directory: %w", err)
	}

	migrationFiles := make([]string, 0)

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}

		migrationFiles = append(
			migrationFiles,
			file.Name(),
		)
	}

	sort.Strings(migrationFiles)

	for _, fileName := range migrationFiles {
		if err := runMigration(
			ctx,
			db,
			migrationsDir,
			fileName,
		); err != nil {
			return err
		}
	}

	return nil
}

func ensureMigrationsTable(
	ctx context.Context,
	db *pgxpool.Pool,
) error {
	_, err := db.Exec(
		ctx,
		`
CREATE TABLE IF NOT EXISTS schema_migrations (
id BIGSERIAL PRIMARY KEY,
version VARCHAR(255) NOT NULL UNIQUE,
applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)
`,
	)

	if err != nil {
		return fmt.Errorf(
			"create schema_migrations table: %w",
			err,
		)
	}

	return nil
}

func runMigration(
	ctx context.Context,
	db *pgxpool.Pool,
	migrationsDir string,
	fileName string,
) error {
	var exists bool

	err := db.QueryRow(
		ctx,
		`
SELECT EXISTS (
SELECT 1
FROM schema_migrations
WHERE version = $1
)
`,
		fileName,
	).Scan(&exists)

	if err != nil {
		return fmt.Errorf(
			"check migration %s: %w",
			fileName,
			err,
		)
	}

	if exists {
		return nil
	}

	filePath := filepath.Join(
		migrationsDir,
		fileName,
	)

	sqlBytes, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf(
			"read migration %s: %w",
			fileName,
			err,
		)
	}

	migrationSQL := strings.TrimSpace(
		string(sqlBytes),
	)

	if migrationSQL == "" {
		return fmt.Errorf(
			"migration %s is empty",
			fileName,
		)
	}

	tx, err := db.Begin(ctx)
	if err != nil {
		return fmt.Errorf(
			"begin migration %s: %w",
			fileName,
			err,
		)
	}

	defer tx.Rollback(ctx)

	if _, err := tx.Exec(
		ctx,
		migrationSQL,
	); err != nil {
		return fmt.Errorf(
			"execute migration %s: %w",
			fileName,
			err,
		)
	}

	if _, err := tx.Exec(
		ctx,
		`
INSERT INTO schema_migrations (
version
)
VALUES (
$1
)
`,
		fileName,
	); err != nil {
		return fmt.Errorf(
			"record migration %s: %w",
			fileName,
			err,
		)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf(
			"commit migration %s: %w",
			fileName,
			err,
		)
	}

	fmt.Printf(
		"migration applied: %s\n",
		fileName,
	)

	return nil
}

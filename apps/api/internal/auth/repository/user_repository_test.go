package repository

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func testDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	databaseURL := os.Getenv(
		"DATABASE_URL",
	)

	if databaseURL == "" {
		databaseURL = "postgres://enterprise_user:enterprise_password@localhost:5432/enterprise_db?sslmode=disable"
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	db, err := pgxpool.New(
		ctx,
		databaseURL,
	)
	if err != nil {
		t.Fatalf(
			"failed to create database pool: %v",
			err,
		)
	}

	if err := db.Ping(ctx); err != nil {
		db.Close()

		t.Fatalf(
			"failed to ping database: %v",
			err,
		)
	}

	return db
}

func TestRepository_CreateUser(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	repo := NewRepository(db)

	ctx := context.Background()

	email := "repo-create-" +
		strings.ReplaceAll(
			time.Now().Format(
				"20060102150405.000000000",
			),
			".",
			"",
		) +
		"@example.com"

	user, err := repo.CreateUser(
		ctx,
		email,
		"hashed-password",
		"Repository Test User",
	)
	if err != nil {
		t.Fatalf(
			"CreateUser failed: %v",
			err,
		)
	}

	if user.ID == "" {
		t.Fatal("expected user ID")
	}

	if user.Email != email {
		t.Fatalf(
			"expected email %q, got %q",
			email,
			user.Email,
		)
	}

	if user.PasswordHash != "hashed-password" {
		t.Fatalf("password hash was not stored correctly")
	}

	if user.FullName != "Repository Test User" {
		t.Fatalf(
			"expected full name %q, got %q",
			"Repository Test User",
			user.FullName,
		)
	}

	if user.Role != "user" {
		t.Fatalf(
			"expected default role user, got %q",
			user.Role,
		)
	}

	if !user.IsActive {
		t.Fatal("expected newly created user to be active")
	}
}

func TestRepository_FindByEmail(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	repo := NewRepository(db)

	ctx := context.Background()

	email := "repo-find-" +
		strings.ReplaceAll(
			time.Now().Format(
				"20060102150405.000000000",
			),
			".",
			"",
		) +
		"@example.com"

	created, err := repo.CreateUser(
		ctx,
		email,
		"hashed-password",
		"Find Test User",
	)
	if err != nil {
		t.Fatalf(
			"CreateUser failed: %v",
			err,
		)
	}

	found, err := repo.FindByEmail(
		ctx,
		email,
	)
	if err != nil {
		t.Fatalf(
			"FindByEmail failed: %v",
			err,
		)
	}

	if found.ID != created.ID {
		t.Fatalf(
			"expected ID %q, got %q",
			created.ID,
			found.ID,
		)
	}

	if found.Email != email {
		t.Fatalf(
			"expected email %q, got %q",
			email,
			found.Email,
		)
	}

	if found.PasswordHash != "hashed-password" {
		t.Fatal("password hash was not retrieved correctly")
	}
}

func TestRepository_FindByEmail_NotFound(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	repo := NewRepository(db)

	_, err := repo.FindByEmail(
		context.Background(),
		"does-not-exist-repository-test@example.com",
	)

	if err == nil {
		t.Fatal(
			"expected error for non-existing user",
		)
	}
}

func TestRepository_CreateUser_DuplicateEmail(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	repo := NewRepository(db)

	ctx := context.Background()

	email := "repo-duplicate-" +
		strings.ReplaceAll(
			time.Now().Format(
				"20060102150405.000000000",
			),
			".",
			"",
		) +
		"@example.com"

	_, err := repo.CreateUser(
		ctx,
		email,
		"hash-one",
		"First User",
	)
	if err != nil {
		t.Fatalf(
			"first CreateUser failed: %v",
			err,
		)
	}

	_, err = repo.CreateUser(
		ctx,
		email,
		"hash-two",
		"Second User",
	)
	if err == nil {
		t.Fatal(
			"expected duplicate email error",
		)
	}
}

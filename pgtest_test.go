package pgtest_test

import (
	"context"
	"github.com/ibrhmkoz/pgtest"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	url pgtest.DesiredStateURL = "file://schema.sql"
)

func TestPGTest(t *testing.T) {
	ctx := context.Background()
	pool := pgtest.New(t, ctx, pgtest.WithDesiredState(url))

	// Given
	_, err := pool.Exec(ctx, `insert into clinics (id, name, email) values ($1, $2, $3)`, "test-clinic-uuid", "Test Clinic", "test@example.com")
	require.NoError(t, err)

	// When
	var name string
	var email string
	err = pool.QueryRow(ctx, `select name, email from clinics`).Scan(&name, &email)

	// Then
	require.NoError(t, err)
	require.Equal(t, "Test Clinic", name)
	require.Equal(t, "test@example.com", email)
}

func TestWithReferentialIntegrityEnabled(t *testing.T) {
	ctx := context.Background()
	pool := pgtest.New(t, ctx, pgtest.WithDesiredState(url))

	_, err := pool.Exec(ctx, `insert into doctors (id, clinic_id) values ($1, $2)`, "test-doctor-uuid", "non-existent-clinic-uuid")
	require.Error(t, err)
}

func TestWithReferentialIntegrityDisabled(t *testing.T) {
	ctx := context.Background()
	pool := pgtest.New(t, ctx,
		pgtest.WithReferentialIntegrityDisabled(),
		pgtest.WithDesiredState(url),
	)

	_, err := pool.Exec(ctx, `insert into doctors (id, clinic_id) values ($1, $2)`, "test-doctor-uuid", "non-existent-clinic-uuid")
	require.NoError(t, err)
}

func TestWithVersion(t *testing.T) {
	ctx := context.Background()
	pool := pgtest.New(t, ctx,
		pgtest.WithVersion("14"),
		pgtest.WithDesiredState(url),
	)

	// Given
	_, err := pool.Exec(ctx, `insert into clinics (id, name, email) values ($1, $2, $3)`, "test-clinic-uuid", "Test Clinic", "test@example.com")
	require.NoError(t, err)

	// When
	var name string
	var email string
	err = pool.QueryRow(ctx, `select name, email from clinics`).Scan(&name, &email)

	// Then
	require.NoError(t, err)
	require.Equal(t, "Test Clinic", name)
	require.Equal(t, "test@example.com", email)
}

func TestWithMigrator(t *testing.T) {
	ctx := context.Background()
	migrations := "file://migrations"
	pool := pgtest.New(t, ctx,
		pgtest.WithMigrator(pgtest.MigratorGoMigrator),
		pgtest.WithDesiredState(migrations),
	)

	// Test that the tables were created
	var tableCount int
	err := pool.QueryRow(ctx, `select count(*) from information_schema.tables where table_schema = 'public'`).Scan(&tableCount)
	require.NoError(t, err)
	// Since go-migrate creates a table named schema_migrations, we should have 3 tables
	require.Equal(t, 3, tableCount)

	// Test inserting and querying data
	_, err = pool.Exec(ctx, `insert into clinics (id, name, email) values ($1, $2, $3)`, "test-clinic-uuid", "Test Clinic", "test@example.com")
	require.NoError(t, err)

	var name string
	var email string
	err = pool.QueryRow(ctx, `select name, email from clinics where id = $1`, "test-clinic-uuid").Scan(&name, &email)
	require.NoError(t, err)
	require.Equal(t, "Test Clinic", name)
	require.Equal(t, "test@example.com", email)
}

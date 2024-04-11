package pgtest

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	dsu DesiredStateURL = "file://schema.sql"
)

func TestPGTest(t *testing.T) {
	ctx := context.Background()
	pool := New(t, ctx, WithDesiredState(dsu))

	// Given
	_, err := pool.Exec(ctx, `INSERT INTO clinics (id, name, email) VALUES ($1, $2, $3)`, "test-clinic-uuid", "Test Clinic", "test@example.com")
	require.NoError(t, err)

	// When
	var name string
	var email string
	err = pool.QueryRow(ctx, `SELECT name, email FROM clinics`).Scan(&name, &email)

	// Then
	require.NoError(t, err)
	require.Equal(t, "Test Clinic", name)
	require.Equal(t, "test@example.com", email)
}

func TestWithReferentialIntegrityEnabled(t *testing.T) {
	ctx := context.Background()
	pool := New(t, ctx, WithDesiredState(dsu))

	_, err := pool.Exec(ctx, `INSERT INTO doctors (id, clinic_id) VALUES ($1, $2)`, "test-doctor-uuid", "non-existent-clinic-uuid")
	require.Error(t, err)
}

func TestWithReferentialIntegrityDisabled(t *testing.T) {
	ctx := context.Background()
	pool := New(t, ctx,
		WithReferentialIntegrityDisabled(),
		WithDesiredState(dsu),
	)

	_, err := pool.Exec(ctx, `INSERT INTO doctors (id, clinic_id) VALUES ($1, $2)`, "test-doctor-uuid", "non-existent-clinic-uuid")
	require.NoError(t, err)
}

func TestWithVersion(t *testing.T) {
	ctx := context.Background()
	pool := New(t, ctx,
		WithVersion("14"),
		WithDesiredState(dsu),
	)

	// Given
	_, err := pool.Exec(ctx, `INSERT INTO clinics (id, name, email) VALUES ($1, $2, $3)`, "test-clinic-uuid", "Test Clinic", "test@example.com")
	require.NoError(t, err)

	// When
	var name string
	var email string
	err = pool.QueryRow(ctx, `SELECT name, email FROM clinics`).Scan(&name, &email)

	// Then
	require.NoError(t, err)
	require.Equal(t, "Test Clinic", name)
	require.Equal(t, "test@example.com", email)
}

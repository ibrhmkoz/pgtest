package pgtest

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	dsu DesiredStateURL = "file://pkg/pgtest/schema.sql"
)

func TestPGTest(t *testing.T) {
	Run(t,
		dsu,
		func(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
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
		},
	)
}

func TestWithReferentialIntegrityEnabled(t *testing.T) {
	Run(t,
		dsu,
		func(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
			_, err := pool.Exec(ctx, `INSERT INTO doctors (id, clinic_id) VALUES ($1, $2)`, "test-doctor-uuid", "non-existent-clinic-uuid")
			require.Error(t, err)
		},
	)
}

func TestWithReferentialIntegrityDisabled(t *testing.T) {
	Run(t,
		dsu,
		func(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
			_, err := pool.Exec(ctx, `INSERT INTO doctors (id, clinic_id) VALUES ($1, $2)`, "test-doctor-uuid", "non-existent-clinic-uuid")
			require.NoError(t, err)
		},
		WithReferentialIntegrityDisabled(),
	)
}

func TestWithVersion(t *testing.T) {
	Run(t,
		dsu,
		func(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
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
		},
		WithVersion("14"),
	)
}

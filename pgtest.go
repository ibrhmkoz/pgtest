package pgtest

import (
	"ariga.io/atlas-go-sdk/atlasexec"
	"context"
	"fmt"
	"github.com/docker/docker/pkg/ioutils"
	"github.com/ibrhmkoz/pgtest/git"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"log"
	"testing"
	"time"
)

type Option func(*DBOptions)

// DBOptions holds the configuration options for the database.
type DBOptions struct {
	referentialIntegrityDisabled bool
	version                      Version
	url                          DesiredStateURL
}

// WithReferentialIntegrityDisabled is an option that disables referential integrity checks in the test database.
// Although referential integrity is a powerful feature provided by relational database management systems (RDBMS),
// it can sometimes unnecessarily complicate the setup of tests. In a relational database, the referential integrity
// chain between entities can be quite long, requiring a significant amount of setup data to satisfy all the foreign
// key constraints. This extensive setup may not be directly relevant to the specific functionality being tested.
// By disabling referential integrity checks, you can simplify the test setup and focus on testing the desired
// behavior without being burdened by the overhead of setting up the entire referential integrity chain.
func WithReferentialIntegrityDisabled() Option {
	return func(opts *DBOptions) {
		opts.referentialIntegrityDisabled = true
	}
}

type Version = string

func WithVersion(v Version) Option {
	return func(opts *DBOptions) {
		opts.version = v
	}
}

type DesiredStateURL = string

func WithDesiredState(url DesiredStateURL) Option {
	return func(opts *DBOptions) {
		opts.url = url
	}
}

func Run(t *testing.T, ctx context.Context, opts ...Option) *pgxpool.Pool {
	// Since integration tests are run in distinct containers, they can be run in parallel.
	t.Parallel()

	t.Helper()

	o := &DBOptions{
		version: "latest",
	}

	for _, opt := range opts {
		opt(o)
	}

	pc, err := spinContainer(ctx, o.version)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := pc.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	cs, err := pc.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}

	if o.url != "" {
		err = reconcileDB(cs, o.url)
		if err != nil {
			t.Fatal(err)
		}
	}

	config, err := pgxpool.ParseConfig(cs)
	if err != nil {
		t.Fatal(err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		pool.Close()
	})

	if o.referentialIntegrityDisabled {
		if err := disableReferentialIntegrity(ctx, pool); err != nil {
			t.Fatalf("failed to disable referential integrity: %v", err)
		}
	}

	return pool
}

func spinContainer(ctx context.Context, version string) (*postgres.PostgresContainer, error) {
	n := "postgres"
	u := "postgres"
	p := "postgres"

	testcontainers.Logger = log.New(&ioutils.NopWriter{}, "", 0)

	pc, err := postgres.RunContainer(ctx,
		testcontainers.WithImage(fmt.Sprintf("docker.io/postgres:%s", version)),
		postgres.WithDatabase(n),
		postgres.WithUsername(u),
		postgres.WithPassword(p),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	return pc, err
}

type ConnectionString = string

func reconcileDB(cs ConnectionString, dsu DesiredStateURL) (err error) {
	r, err := git.Root()
	if err != nil {
		return err
	}

	client, err := atlasexec.NewClient(r, "atlas")
	if err != nil {
		return fmt.Errorf("failed to initialize client: %v", err)
	}

	_, err = client.SchemaApply(context.Background(), &atlasexec.SchemaApplyParams{
		URL:    cs,
		DevURL: "docker://postgres",
		To:     dsu,
	})

	return err
}

type DropConstraintQuery = string

func disableReferentialIntegrity(ctx context.Context, pool *pgxpool.Pool) (err error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer func(tx pgx.Tx, ctx context.Context) {
		if err != nil {
			err = tx.Rollback(ctx)
		}
	}(tx, ctx)

	// Generate and execute commands to drop all foreign key constraints
	rows, err := tx.Query(ctx, `
        SELECT 'ALTER TABLE ' || nspname || '."' || relname || '" DROP CONSTRAINT "' || conname || '";'
        FROM pg_constraint
        INNER JOIN pg_class ON conrelid = pg_class.oid
        INNER JOIN pg_namespace ON pg_namespace.oid = pg_class.relnamespace
        WHERE contype = 'f';
    `)
	if err != nil {
		return err
	}
	defer rows.Close()

	var d []DropConstraintQuery
	for rows.Next() {
		var cmd string
		if err := rows.Scan(&cmd); err != nil {
			return err
		}
		d = append(d, cmd)
	}

	// Execute all drop constraint queries within the same transaction
	for _, q := range d {
		if _, err := tx.Exec(ctx, q); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

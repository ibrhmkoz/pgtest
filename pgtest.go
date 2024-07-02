package pgtest

import (
	"ariga.io/atlas-go-sdk/atlasexec"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/docker/docker/pkg/ioutils"
	"github.com/golang-migrate/migrate/v4"
	pgmigrate "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
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

type Option func(*dbOptions)

type Migrator int

const (
	MigratorAtlas Migrator = iota
	MigratorGoMigrator
)

type dbOptions struct {
	referentialIntegrityDisabled bool
	version                      Version
	url                          DesiredStateURL
	migrator                     Migrator
}

// WithReferentialIntegrityDisabled is an option that disables referential integrity checks in the test database.
// Although referential integrity is a powerful feature provided by relational database management systems (RDBMS),
// it can sometimes unnecessarily complicate the setup of tests. In a relational database, the referential integrity
// chain between entities can be quite long, requiring a significant amount of setup data to satisfy all the foreign
// key constraints. This extensive setup may not be directly relevant to the specific functionality being tested.
// By disabling referential integrity checks, you can simplify the test setup and focus on testing the desired
// behavior without being burdened by the overhead of setting up the entire referential integrity chain.
func WithReferentialIntegrityDisabled() Option {
	return func(opts *dbOptions) {
		opts.referentialIntegrityDisabled = true
	}
}

// Version specifies postgres version.
type Version = string

// WithVersion allows specifying which version of postgres to use.
func WithVersion(v Version) Option {
	return func(opts *dbOptions) {
		opts.version = v
	}
}

// DesiredStateURL points to either a file or a dir comprising sql files.
type DesiredStateURL = string

// WithDesiredState allows initializing test db with a desired state.
func WithDesiredState(url DesiredStateURL) Option {
	return func(opts *dbOptions) {
		opts.url = url
	}
}

// WithMigrator allows specifying which migrator to use.
func WithMigrator(m Migrator) Option {
	return func(opts *dbOptions) {
		opts.migrator = m
	}
}

// New prepares a brand-new postgres instance by getting it to the intended state.
// It delivers a pgx pool through which the client can interact with the test DB.
func New(t *testing.T, ctx context.Context, opts ...Option) *pgxpool.Pool {
	// Since integration tests are run in distinct containers, they can be run in parallel.
	t.Parallel()

	t.Helper()

	o := &dbOptions{
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

	root, err := git.Root()

	if o.url != "" {
		switch o.migrator {
		case MigratorAtlas:
			err = atlasMigrate(root, cs, o.url)
		case MigratorGoMigrator:
			err = goMigratorMigrate(cs, o.url)
		default:
			t.Fatal("unreachable")
		}

		if err != nil {
			t.Fatal(err)
		}
	}

	pool, err := pgxpool.New(ctx, cs)

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

type connectionString = string

func atlasMigrate(root git.AbsolutePath, cs connectionString, dsu DesiredStateURL) (err error) {
	client, err := atlasexec.NewClient(root, "atlas")
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

func goMigratorMigrate(cs connectionString, dsu DesiredStateURL) (err error) {
	db, err := sql.Open("postgres", cs)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}
	defer func(db *sql.DB) {
		err = db.Close()
	}(db)

	driver, err := pgmigrate.WithInstance(db, &pgmigrate.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		dsu,
		"postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %v", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to run migrations: %v", err)
	}

	return nil
}

type dropConstraintQuery = string

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
        select 'alter table ' || nspname || '."' || relname || '" drop constraint "' || conname || '";'
        from pg_constraint
        inner join pg_class on conrelid = pg_class.oid
        inner join pg_namespace on pg_namespace.oid = pg_class.relnamespace
        where contype = 'f';
    `)
	if err != nil {
		return err
	}
	defer rows.Close()

	var d []dropConstraintQuery
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

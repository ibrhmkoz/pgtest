# pgtest

`pgtest` is a Go package that provides a convenient way to set up and run tests against a PostgreSQL database using Docker containers. It leverages the power of [Testcontainers](https://www.testcontainers.org/) to spin up isolated PostgreSQL containers for each test, ensuring a clean and reproducible testing environment.

## Features

- Spin up isolated PostgreSQL containers for each test
- Apply database schema migrations using [Atlas](https://atlasgo.io/)
- Disable referential integrity checks for simplified test setup
- Specify desired PostgreSQL version for testing
- Run tests in parallel for improved performance

## Installation

To use `pgtest` in your Go project, you can install it using `go get`:

```shell
go get github.com/ibrhmkoz/pgtest
```

## Usage

Here's an example of how to use `pgtest` in your test code:

```go
package mytests

import (
    "context"
    "github.com/ibrhmkoz/pgtest"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/stretchr/testify/require"
    "testing"
)

const (
    dsu pgtest.DesiredStateURL = "file://path/to/schema.sql"
)

func TestMyFunction(t *testing.T) {
    pgtest.Run(t,
        dsu,
        func(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
            // Test code here
            // Use the pool to interact with the database

            // Example:
            _, err := pool.Exec(ctx, `INSERT INTO mytable (id, name) VALUES ($1, $2)`, 1, "John Doe")
            require.NoError(t, err)

            var name string
            err = pool.QueryRow(ctx, `SELECT name FROM mytable WHERE id = $1`, 1).Scan(&name)
            require.NoError(t, err)
            require.Equal(t, "John Doe", name)
        },
    )
}
```

In the above example, we define a test function `TestMyFunction` that uses `pgtest.Run` to set up a PostgreSQL container, apply the schema migrations specified in `schema.sql`, and run the test code. The test code receives a `*pgxpool.Pool` instance that can be used to interact with the database.

### Options

`pgtest` provides a few options to customize the behavior of the test setup:

- `WithReferentialIntegrityDisabled()`: Disables referential integrity checks in the test database, allowing for simplified test setup when dealing with complex referential integrity chains.
- `WithVersion(version string)`: Specifies the desired PostgreSQL version to use for testing. Defaults to "latest" if not provided.

## Contributing

Contributions to `pgtest` are welcome! If you find any issues or have suggestions for improvements, please open an issue or submit a pull request on the [GitHub repository](https://github.com/ibrhmkoz/pgtest).

## License

`pgtest` is open-source software released under the [MIT License](https://opensource.org/licenses/MIT).
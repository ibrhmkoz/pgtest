# pgtest

`pgtest` is a Go package that simplifies testing against PostgreSQL databases. It provides a clean and efficient testing environment by running tests in parallel, applying database schema migrations, and offering options to disable referential integrity checks and specify the desired PostgreSQL version. `pgtest` follows best practices for establishing database connections by leveraging the `pgx/pool` package, which offers superior performance and features compared to the standard library's `database/sql`.

## Features

- Run tests in parallel for improved performance
- Apply database schema migrations using [Atlas](https://atlasgo.io/)
- Disable referential integrity checks for simplified test setup
- Specify desired PostgreSQL version for testing
- Utilizes `pgx/pool` for efficient and feature-rich database connection management

## Installation

```shell
go get github.com/ibrhmkoz/pgtest
```

## Usage

```go
import (
"context"
"testing"
"github.com/ibrhmkoz/pgtest"
)

func TestMyFunction(t *testing.T) {
ctx := context.Background()
pool := pgtest.New(t, ctx, pgtest.WithDesiredState("file://path/to/schema.sql"))

// Test code here
// Use the pool to interact with the database
}
```

### Options

`pgtest` provides a few options to customize the behavior of the test setup:

- `WithReferentialIntegrityDisabled()`: Disables referential integrity checks in the test database, allowing for simplified test setup when dealing with complex referential integrity chains.
- `WithVersion(version string)`: Specifies the desired PostgreSQL version to use for testing. Defaults to "latest" if not provided.
- `WithDesiredState(url string)`: Specifies the desired state URL for the database schema. The URL can point to a SQL file or a directory consisting of SQL files.

## Example

```go
package pgtest

import (
    "context"
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
```

## Contributing

Contributions are welcome! Please open an issue or submit a pull request for any bugs, feature requests, or improvements.

## License

`pgtest` is released under the [MIT License](https://opensource.org/licenses/MIT).
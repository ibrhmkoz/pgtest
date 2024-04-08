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
import "github.com/ibrhmkoz/pgtest"

func TestMyFunction(t *testing.T) {
    pgtest.Run(t,
        "file://path/to/schema.sql",
        func(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
            // Test code here
            // Use the pool to interact with the database
        },
    )
}
```

### Options

`pgtest` provides a few options to customize the behavior of the test setup:

- `WithReferentialIntegrityDisabled()`: Disables referential integrity checks in the test database, allowing for simplified test setup when dealing with complex referential integrity chains.
- `WithVersion(version string)`: Specifies the desired PostgreSQL version to use for testing. Defaults to "latest" if not provided.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request for any bugs, feature requests, or improvements.

## License

`pgtest` is released under the [MIT License](https://opensource.org/licenses/MIT).
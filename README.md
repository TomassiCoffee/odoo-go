# odoo-go

Reusable Odoo JSON-2 client and schema-driven Go type generator.

## Library

```go
import odoo "github.com/TomassiCoffee/odoo-go"

client, err := odoo.NewClientFromEnv()
```

The client reads `ODOO_URL`, `ODOO_API_KEY` (or `ODOO_PASSWORD`/`ODOO_TOKEN`) and optionally `ODOO_DATABASE`/`ODOO_DB`.

## Generate database-specific models

Generated models belong to the consuming application, not this library. From a consuming module:

```bash
go run github.com/TomassiCoffee/odoo-go/cmd/odoo-gen \
  -output internal/odoomodels/models_gen.go \
  -package odoomodels \
  -cache internal/odoomodels/models_metadata.json
```

Optionally restrict introspection/generation:

```bash
go run github.com/TomassiCoffee/odoo-go/cmd/odoo-gen \
  -models account.account,account.journal,account.move,account.move.line,account.bank.statement.line
```

The command introspects `ir.model` and `ir.model.fields`; if Odoo is unavailable it can fall back to the cache file. Generated source imports this module and is formatted with `go/format`.

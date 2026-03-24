//go:build !nolint

package main

// Import the generated Swagger docs so the swaggo HTTP handler can serve the
// OpenAPI spec. Excluded from linting via the nolint build tag because
// swaggo/swag uses Go export-data format v2, which older golangci-lint
// binaries cannot read, causing cascade type-checking failures.
import _ "github.com/avito-internships/test-backend-1-cQu1x/docs"

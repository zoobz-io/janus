// Command adminspec dumps the admin API's OpenAPI specification to disk without
// starting a server or touching the database. It builds a bare rocco engine,
// applies the same OpenAPI metadata and endpoint set the admin server uses, and
// marshals the generated spec. The output feeds the @janus/admin-sdk client
// generator in the web monorepo.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/zoobz-io/openapi"
	"github.com/zoobz-io/rocco"

	adminhandlers "github.com/zoobz-io/janus/admin/handlers"
)

// defaultOut is where the spec lands when no path argument is given: the admin
// SDK package's committed spec snapshot.
const defaultOut = "web/packages/admin-sdk/openapi.json"

func main() {
	if err := run(); err != nil {
		log.Fatalf("adminspec: %v", err)
	}
}

func run() error {
	out := defaultOut
	if len(os.Args) > 1 {
		out = os.Args[1]
	}

	// A bare engine carries no DB, DI, or auth — endpoint registration only
	// records handler metadata (paths, request/response types, error defs), and
	// each handler resolves its dependencies lazily at request time. That is all
	// GenerateOpenAPI needs, so no runtime boot is required.
	e := rocco.NewEngine()
	adminhandlers.ConfigureOpenAPI(e)
	e.WithHandlers(adminhandlers.All()...)

	spec := e.GenerateOpenAPI(nil)
	patchSpec(spec)

	data, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal spec: %w", err)
	}
	data = append(data, '\n')

	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}
	if err := os.WriteFile(out, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", out, err)
	}

	fmt.Printf("wrote %s (%d bytes)\n", out, len(data))
	return nil
}

// patchSpec repairs known gaps in rocco's generated spec so the output is
// self-contained and validates against strict OpenAPI tooling.
//
// rocco (v0.1.19) inlines validation-error details as an array of
// ValidationFieldError and emits a $ref to that named schema, but never adds the
// schema to components. We backfill it here from the known rocco type shape.
func patchSpec(spec *openapi.OpenAPI) {
	if spec.Components == nil || spec.Components.Schemas == nil {
		return
	}
	if _, ok := spec.Components.Schemas["ValidationFieldError"]; ok {
		return
	}
	spec.Components.Schemas["ValidationFieldError"] = &openapi.Schema{
		Type: openapi.NewSchemaType("object"),
		Properties: map[string]*openapi.Schema{
			"field":   {Type: openapi.NewSchemaType("string"), Description: "The field that failed validation"},
			"message": {Type: openapi.NewSchemaType("string"), Description: "Description of the validation failure"},
		},
		Required: []string{"field", "message"},
	}
}

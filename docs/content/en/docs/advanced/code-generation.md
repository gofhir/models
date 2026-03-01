---
title: "Code Generation"
linkTitle: "Code Generation"
description: "How the cmd/generator tool reads FHIR StructureDefinitions and produces Go source code for resources, builders, serialization, and metadata."
weight: 3
---

All Go source code in the `r4`, `r4b`, and `r5` packages is automatically generated from the official FHIR StructureDefinitions. The `cmd/generator` tool orchestrates this process, reading JSON specification files from the `specs/` directory and producing Go source files in the target package directory.

## Running the Generator

The generator is invoked from the repository root using `go run`:

```bash
# Generate R4 code
go run cmd/generator/main.go r4

# Generate R4B code
go run cmd/generator/main.go r4b

# Generate R5 code
go run cmd/generator/main.go r5
```

The generator accepts a single argument -- the FHIR version to generate. It must be one of `r4`, `r4b`, or `r5`.

{{< callout type="info" >}}
You do not need to run the generator to use the library. All generated code is committed to the repository and published as Go modules. The generator is only needed when updating to a new FHIR specification release or modifying the generation templates.
{{< /callout >}}

## What Gets Generated

For each FHIR version, the generator produces the following files:

| Output | Description |
|--------|-------------|
| `resource_*.go` | One file per resource type containing the struct, JSON/XML marshaling, builder, and functional options |
| `datatypes.go` | All FHIR data types (HumanName, Address, CodeableConcept, Quantity, etc.) |
| `codesystems.go` | All FHIR code system enumerations as Go `string` types with constants |
| `interfaces.go` | The `Resource` and `DomainResource` interfaces |
| `registry.go` | The resource factory map and polymorphic deserialization functions |
| `fhirpath_model.go` | The `FHIRPathModelData` singleton with all type metadata maps |
| `summary.go` | The `SummaryFields` map with isSummary field lists for each resource |
| `marshal.go` | Custom JSON marshaling functions (`Marshal`, `MarshalIndent`) |
| `xml_helpers.go` | XML serialization helper functions and namespace constants |

Each resource file (e.g., `resource_patient.go`) contains:

1. **The resource struct** with JSON and XML tags
2. **Interface method implementations** (`GetResourceType`, `GetId`, `SetId`, `GetMeta`, `SetMeta`, and `DomainResource` methods where applicable)
3. **`MarshalJSON`/`UnmarshalJSON`** for automatic `resourceType` injection and contained resource handling
4. **`MarshalXML`/`UnmarshalXML`** for FHIR-conformant XML encoding with proper namespace and primitive attribute handling
5. **Backbone element structs** (e.g., `PatientContact`, `PatientCommunication`) with their own marshaling methods
6. **A fluent builder** (`PatientBuilder` with `NewPatientBuilder`, `Set*`, `Add*`, `Build`)
7. **Functional options** (`PatientOption` type, `NewPatient`, `WithPatient*` functions)

## Internal Pipeline

The generator follows a three-stage pipeline:

### 1. Parser

The parser reads FHIR StructureDefinition JSON files from the `specs/<version>/` directory. These files are the official machine-readable definitions published by HL7 for each FHIR version. The parser extracts:

- Resource definitions and their element hierarchies
- Data type definitions (primitive and complex)
- Code system value sets and their allowed codes
- Element metadata: cardinality, types, isSummary, reference targets, choice type variants

### 2. Analyzer

The analyzer processes the parsed data into an internal model suitable for code generation. Key transformations include:

- **Flattening element hierarchies** into Go-friendly struct field definitions
- **Resolving choice types** (e.g., `value[x]`) into separate fields per type variant
- **Building the type hierarchy** for interface satisfaction checks
- **Computing backbone element boundaries** to determine which nested elements need their own struct
- **Collecting FHIRPath metadata** into the six maps that populate `FHIRPathModelData`
- **Extracting summary flags** to build the `SummaryFields` map

### 3. Generator

The generator takes the analyzed model and renders Go source files using Go's `text/template` package. Templates handle:

- Struct field generation with correct Go types, JSON tags, and XML tags
- Interface method generation based on whether a resource is a base Resource or DomainResource
- Builder and functional option generation following consistent naming patterns
- XML marshal/unmarshal generation with FHIR-specific encoding rules
- Code system constant generation with Go-friendly names

After generation, the output files are formatted with `gofmt` to ensure consistent style.

## Configuration

The generator uses a `Config` struct to determine its behavior:

```go
type Config struct {
    SpecsDir    string // Path to specs/<version>/ directory
    OutputDir   string // Path to output package directory (e.g., ./r4)
    PackageName string // Go package name (e.g., "r4")
    Version     string // FHIR version identifier (e.g., "r4")
}
```

When invoked via `cmd/generator/main.go`, these paths are resolved relative to the repository root.

## Adding a New FHIR Version

To add support for a new FHIR version:

1. Download the StructureDefinition JSON files from the HL7 FHIR specification
2. Place them in `specs/<version>/`
3. Create the output directory (e.g., `r6/`)
4. Initialize a Go module in the output directory (`go mod init github.com/gofhir/models/r6`)
5. Run the generator: `go run cmd/generator/main.go r6`
6. Add the new module to `go.work`
7. Add a release-please entry in `release-please-config.json`

## Modifying Generated Code

{{< callout type="warning" >}}
Never edit the generated files directly. All files in `r4/`, `r4b/`, and `r5/` (except `helpers/`) begin with the comment `// Code generated by gofhir. DO NOT EDIT.` and will be overwritten when the generator runs.
{{< /callout >}}

To change the generated output, modify the templates and generation logic in the `internal/codegen/generator` package, then re-run the generator for all affected versions.

The `helpers/` subdirectory (e.g., `r4/helpers/`) is hand-written and is not affected by the generator. It provides convenience constants and functions that build on top of the generated types.

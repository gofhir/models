---
title: "Multi-Version Support"
linkTitle: "Multi-Version"
description: "Go workspace architecture for importing multiple FHIR versions side by side with independent module versioning."
weight: 4
---

The `gofhir/models` repository publishes each FHIR version as an independent Go module. This architecture allows you to import a single version, or import multiple versions in the same project without conflicts.

## Module Structure

Each FHIR version lives in its own directory at the repository root with its own `go.mod`:

```
models/
  go.work             # Go workspace file for local development
  r4/
    go.mod            # module github.com/gofhir/models/r4
    helpers/          # Hand-written helper constants
  r4b/
    go.mod            # module github.com/gofhir/models/r4b
  r5/
    go.mod            # module github.com/gofhir/models/r5
```

The root `go.work` file ties these modules together for local development:

```go
go 1.23

use (
    ./cmd/generator
    ./r4
    ./r4b
    ./r5
)
```

## Importing a Single Version

Most projects only need one FHIR version. Import the corresponding module directly:

```go
import "github.com/gofhir/models/r4"

patient := r4.NewPatient(
    r4.WithPatientId("example"),
)
```

Install with:

```bash
go get github.com/gofhir/models/r4
```

## Importing Multiple Versions

When you need to work with multiple FHIR versions (for example, a converter or a server that supports multiple versions), import them with package aliases:

```go
import (
    r4 "github.com/gofhir/models/r4"
    r5 "github.com/gofhir/models/r5"
)

func convertPatient(r4Patient *r4.Patient) *r5.Patient {
    r5Patient := r5.NewPatientBuilder().
        SetId(*r4Patient.Id).
        SetBirthDate(*r4Patient.BirthDate).
        Build()

    // Copy names
    for _, name := range r4Patient.Name {
        r5Patient.Name = append(r5Patient.Name, r5.HumanName{
            Family: name.Family,
            Given:  name.Given,
        })
    }

    return r5Patient
}
```

Install both modules:

```bash
go get github.com/gofhir/models/r4
go get github.com/gofhir/models/r5
```

{{< callout type="info" >}}
Since each FHIR version is a completely separate Go module, there are no type conflicts. An `r4.Patient` and an `r5.Patient` are distinct types, even though they share the same struct name within their packages.
{{< /callout >}}

## Package Aliasing Patterns

When importing multiple versions, the default package name matches the directory name (`r4`, `r4b`, `r5`), so aliasing is only necessary if you want custom names:

```go
import (
    // Default names -- no alias needed
    "github.com/gofhir/models/r4"
    "github.com/gofhir/models/r4b"
    "github.com/gofhir/models/r5"
)

// Or use descriptive aliases
import (
    fhirR4  "github.com/gofhir/models/r4"
    fhirR4B "github.com/gofhir/models/r4b"
    fhirR5  "github.com/gofhir/models/r5"
)
```

## Independent Versioning

Each module is versioned independently using [release-please](https://github.com/googleapis/release-please). This means:

- A bug fix in the R4 module does not force a version bump in R4B or R5
- Breaking changes in one version package do not affect others
- Each module has its own `CHANGELOG.md` tracking changes

Version tags follow the pattern `<component>/v<version>`:

```
r4/v0.3.0
r4b/v0.2.1
r5/v0.1.0
```

The `release-please-config.json` at the repository root configures this multi-package release strategy:

```json
{
  "packages": {
    "r4": {
      "release-type": "go",
      "component": "r4",
      "changelog-path": "CHANGELOG.md"
    },
    "r4b": {
      "release-type": "go",
      "component": "r4b",
      "changelog-path": "CHANGELOG.md"
    },
    "r5": {
      "release-type": "go",
      "component": "r5",
      "changelog-path": "CHANGELOG.md"
    }
  }
}
```

## Version-Specific Features

While the three packages share the same overall structure, each FHIR version has differences in its resource set and data types. For example:

- **R4** includes resources like `MedicinalProduct` and `EffectEvidenceSynthesis` that were removed in R5
- **R4B** is largely identical to R4 with incremental updates to certain resources
- **R5** introduces new resources and refactors others (e.g., the medication-related resources were restructured)

The generated code for each version precisely reflects the official StructureDefinitions for that FHIR release, so you always get the correct type definitions for the version you are targeting.

## Go Workspace for Development

If you are contributing to the `gofhir/models` project itself, the `go.work` file allows all modules to be developed together without publishing. The workspace ensures that local changes to shared infrastructure (like the generator) are immediately visible across all version packages.

```bash
# Run all R4 tests
cd r4 && go test ./...

# Run all R4B tests
cd r4b && go test ./...

# Run all R5 tests
cd r5 && go test ./...
```

The `go.work` file is not included in published modules and does not affect downstream consumers.

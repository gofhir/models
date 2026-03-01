---
title: "FHIR Versions"
linkTitle: "FHIR Versions"
description: "Understand the differences between FHIR R4, R4B, and R5 and how each version is packaged as a Go module."
weight: 3
---

The **gofhir/models** project supports three FHIR specification versions. Each version is published as an independent Go module with its own types, builders, code systems, and serialization functions.

## Supported Versions

| Version | FHIR Spec | Go Module Path | Status |
|---------|-----------|----------------|--------|
| **R4** | 4.0.1 | `github.com/gofhir/models/r4` | Stable, most widely adopted |
| **R4B** | 4.3.0 | `github.com/gofhir/models/r4b` | Stable, transitional release |
| **R5** | 5.0.0 | `github.com/gofhir/models/r5` | Stable, latest normative release |

## Version Differences

### FHIR R4 (4.0.1)

R4 is the most widely deployed FHIR version in production systems. It was published in 2019 and is used by the US Core Implementation Guide, the SMART on FHIR framework, and most commercial FHIR server implementations.

```go
import "github.com/gofhir/models/r4"

patient := r4.NewPatient(
    r4.WithPatientId("r4-example"),
    r4.WithPatientGender(r4.AdministrativeGenderMale),
)
```

The R4 package also includes an optional `helpers` sub-package with pre-built `CodeableConcept` values:

```go
import "github.com/gofhir/models/r4/helpers"

// Use a pre-built vital signs category
category := helpers.ObservationCategoryVitalSigns
```

### FHIR R4B (4.3.0)

R4B is a transitional release published in 2022. It is backwards-compatible with R4 for most resources, but introduces new resources and updates to terminology-related resources (CodeSystem, ValueSet, ConceptMap) that align with the R5 direction.

```go
import "github.com/gofhir/models/r4b"

patient := r4b.NewPatient(
    r4b.WithPatientId("r4b-example"),
    r4b.WithPatientGender(r4b.AdministrativeGenderFemale),
)
```

R4B is typically used when you need to support systems that are transitioning from R4 toward R5.

### FHIR R5 (5.0.0)

R5 is the latest normative release, published in 2023. It includes significant changes to several resources, new resources, updated code systems, and structural changes to backbone elements.

```go
import "github.com/gofhir/models/r5"

patient := r5.NewPatient(
    r5.WithPatientId("r5-example"),
    r5.WithPatientGender(r5.AdministrativeGenderMale),
)
```

Key differences in R5 include changes to Observation component structures, new resources like `SubscriptionTopic`, and updated terminology resources.

## Module Path Structure

Each version is its own Go module with a separate `go.mod` file. This means:

1. **Independent versioning** -- Each module has its own semantic version. An update to the R5 package does not require updating R4.
2. **No dependency conflicts** -- Importing multiple versions in the same project does not create diamond dependency issues.
3. **Minimal binary size** -- Your compiled binary only includes the FHIR version(s) you import.

```
github.com/gofhir/models/
    r4/                # module: github.com/gofhir/models/r4
        go.mod
        helpers/       # sub-package within the r4 module
    r4b/               # module: github.com/gofhir/models/r4b
        go.mod
    r5/                # module: github.com/gofhir/models/r5
        go.mod
```

## Choosing a Version

- **Use R4** if you are building against US Core, SMART on FHIR, or any system that implements FHIR R4. This is the safest default for most projects.
- **Use R4B** if your target system specifically requires the 4.3.0 specification, especially for updated terminology resources.
- **Use R5** if you are building greenfield applications or targeting systems that have adopted the latest FHIR standard.

## API Consistency Across Versions

All three packages expose the same API patterns:

- Struct types: `r4.Patient`, `r4b.Patient`, `r5.Patient`
- Functional options: `r4.NewPatient(opts...)`, `r4b.NewPatient(opts...)`, `r5.NewPatient(opts...)`
- Builders: `r4.NewPatientBuilder()`, `r4b.NewPatientBuilder()`, `r5.NewPatientBuilder()`
- Serialization: `r4.Marshal(v)`, `r4b.Marshal(v)`, `r5.Marshal(v)`
- Deserialization: `r4.UnmarshalResource(data)`, `r4b.UnmarshalResource(data)`, `r5.UnmarshalResource(data)`
- Code systems: `r4.AdministrativeGenderMale`, `r4b.AdministrativeGenderMale`, `r5.AdministrativeGenderMale`

The struct fields and code system values differ between versions to match the corresponding FHIR StructureDefinitions, but the construction and serialization patterns remain the same.

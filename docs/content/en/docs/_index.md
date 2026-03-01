---
title: "Documentation"
linkTitle: "Documentation"
description: "Complete documentation for gofhir/models -- type-safe Go structs for FHIR R4, R4B, and R5 resources."
weight: 1
---

Welcome to the **gofhir/models** documentation. This library provides auto-generated, type-safe Go structs for every resource and data type defined in the HL7 FHIR specification. It covers FHIR R4, R4B, and R5, each published as an independent Go module.

## Where to Start

{{< cards cols="2" >}}
  {{< card link="getting-started" title="Getting Started" subtitle="Install the library, create your first FHIR resource, and learn about supported FHIR versions." icon="play" >}}
  {{< card link="resource-construction" title="Resource Construction" subtitle="Explore three patterns for building resources: struct literals, fluent builders, and functional options." icon="puzzle" >}}
{{< /cards >}}

## Key Features

- **All FHIR versions** -- R4 (4.0.1), R4B (4.3.0), and R5 (5.0.0) with every resource, backbone element, data type, and code system.
- **Three construction patterns** -- Choose between direct struct literals, fluent builder chains, or functional options depending on your use case.
- **JSON serialization** -- HTML-safe marshaling that preserves FHIR narrative XHTML content, with automatic `resourceType` injection.
- **XML serialization** -- Full FHIR-conformant XML marshaling and unmarshaling via `MarshalXML` and `UnmarshalXML` on every type.
- **Polymorphic deserialization** -- `UnmarshalResource(data)` automatically detects `resourceType` and returns the correct Go struct.
- **Precision-preserving decimals** -- A custom `Decimal` type stores the exact textual representation (e.g., `"1.50"` keeps the trailing zero).
- **FHIRPath model metadata** -- Runtime type information for FHIRPath engines, including choice type paths, type hierarchies, and reference targets.
- **Helper constants** -- Pre-built `CodeableConcept` values for common observation categories, condition categories, LOINC codes, and UCUM units.
- **Resource and DomainResource interfaces** -- Standard Go interfaces matching the FHIR type hierarchy for generic resource handling.

## Package Overview

| Package | Description |
|---------|-------------|
| `github.com/gofhir/models/r4` | FHIR R4 (4.0.1) resources, data types, code systems, builders, and serialization |
| `github.com/gofhir/models/r4b` | FHIR R4B (4.3.0) resources, data types, code systems, builders, and serialization |
| `github.com/gofhir/models/r5` | FHIR R5 (5.0.0) resources, data types, code systems, builders, and serialization |
| `github.com/gofhir/models/r4/helpers` | Pre-built CodeableConcepts for common categories (observation, condition, LOINC, UCUM) |

## Project Structure

Each FHIR version lives in its own directory at the repository root and is published as a separate Go module:

```
models/
  r4/            # github.com/gofhir/models/r4
    helpers/     # github.com/gofhir/models/r4/helpers
  r4b/           # github.com/gofhir/models/r4b
  r5/            # github.com/gofhir/models/r5
```

All types are automatically generated from the official FHIR StructureDefinitions by the code generator in the `cmd/generator` directory. You should never edit the generated source files directly.

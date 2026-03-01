---
title: "Advanced Topics"
linkTitle: "Advanced"
description: "Advanced features of gofhir/models including FHIRPath model metadata, summary fields, code generation, and multi-version support."
weight: 5
---

Beyond basic resource construction and serialization, the `gofhir/models` library provides several advanced features that support building production FHIR infrastructure. These capabilities are generated alongside the resource structs and expose the rich metadata embedded in the FHIR specification.

## Topics

{{< cards >}}
  {{< card link="fhirpath-model" title="FHIRPath Model" subtitle="Runtime type metadata for FHIRPath expression evaluation, including choice types, type hierarchies, and reference targets." icon="academic-cap" >}}
  {{< card link="summary-fields" title="Summary Fields" subtitle="Pre-computed summary field lists for implementing the _summary=true search parameter on a FHIR server." icon="document-text" >}}
  {{< card link="code-generation" title="Code Generation" subtitle="How the generator reads FHIR StructureDefinitions and produces all Go source code for resources, builders, and metadata." icon="cog" >}}
  {{< card link="multi-version" title="Multi-Version Support" subtitle="Go workspace architecture for importing multiple FHIR versions side by side with independent module versioning." icon="collection" >}}
{{< /cards >}}

## Overview

| Feature | Package Export | Purpose |
|---------|---------------|---------|
| FHIRPath model | `FHIRPathModel()` | Runtime type information for FHIRPath engines |
| Summary fields | `SummaryFields` | Field lists for `_summary=true` server behavior |
| Code generator | `cmd/generator` | Regenerate all Go code from FHIR StructureDefinitions |
| Multi-version | `go.work` | Import R4, R4B, and R5 in the same project |

Each of these features is designed to be used independently. You do not need a FHIRPath engine to use summary fields, and you do not need the code generator to use the library at all -- the generated code is already committed and published as Go modules.

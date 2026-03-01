---
title: "API Reference"
linkTitle: "API Reference"
description: "Complete API reference for gofhir/models interfaces, registry functions, builders, and FHIRPath model accessors."
weight: 6
---

This section provides a detailed API reference for the core types and functions exported by the `gofhir/models` packages. The same API surface is available in each version package (`r4`, `r4b`, `r5`), with type definitions reflecting the corresponding FHIR specification.

## Topics

{{< cards >}}
  {{< card link="resource-interfaces" title="Resource Interfaces" subtitle="The Resource and DomainResource interfaces that all FHIR resources implement." icon="code" >}}
  {{< card link="registry-functions" title="Registry Functions" subtitle="Factory, deserialization, and introspection functions for dynamic resource handling." icon="server" >}}
  {{< card link="builder-api" title="Builder API" subtitle="Fluent builder pattern with Set/Add methods and functional options for every resource type." icon="puzzle" >}}
  {{< card link="fhirpath-model-api" title="FHIRPath Model API" subtitle="Complete reference for the FHIRPathModelData accessor methods." icon="academic-cap" >}}
{{< /cards >}}

## API Summary

The library exports a small, focused API surface on top of the generated resource structs:

| Category | Key Exports | Purpose |
|----------|-------------|---------|
| Interfaces | `Resource`, `DomainResource` | Generic resource handling without type assertions |
| Registry | `NewResource`, `UnmarshalResource`, `GetResourceType`, `IsKnownResourceType`, `AllResourceTypes` | Dynamic resource creation and deserialization |
| Builders | `New<Resource>Builder`, `Set*`, `Add*`, `Build` | Fluent resource construction |
| Functional Options | `New<Resource>`, `With<Resource><Field>` | Concise resource creation with options |
| Serialization | `Marshal`, `MarshalIndent`, `MarshalResourceXML`, `MarshalResourceXMLIndent`, `UnmarshalResourceXML` | FHIR-conformant JSON and XML encoding |
| Metadata | `FHIRPathModel`, `SummaryFields` | Runtime type information and summary field lists |

All examples in this section use the `r4` package. The same patterns apply to `r4b` and `r5` with their respective type definitions.

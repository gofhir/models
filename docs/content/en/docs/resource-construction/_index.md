---
title: "Resource Construction"
linkTitle: "Resource Construction"
description: "Learn two patterns for building FHIR resources in Go: struct literals and fluent builders."
weight: 2
---

The **gofhir/models** library provides three distinct patterns for creating FHIR resources. All three produce the same Go struct types, so you can mix and match patterns within a single project. Choose the one that best fits your coding style and use case.

## Overview

| Pattern | Best For | Verbosity | Compile-time Safety |
|---------|----------|-----------|---------------------|
| [Struct Literals](struct-literals) | Full control, one-shot initialization | Medium (requires pointers) | Full |
| [Builder Pattern](builder-pattern) | Step-by-step construction, fluent chains | Low | Full |

Both patterns set fields on the same underlying struct (for example, `r4.Patient`), so serialization and deserialization work identically regardless of how the resource was created.

## Quick Comparison

### Struct Literal

```go
active := true
patient := r4.Patient{
    Id:           &id,
    Active:       &active,
}
```

### Builder Pattern

```go
patient := r4.NewPatientBuilder().
    SetId("patient-1").
    SetActive(true).
    Build()
```

## Guides

{{< cards >}}
  {{< card link="struct-literals" title="Struct Literals" subtitle="Direct struct initialization with full control over every field." icon="code" >}}
  {{< card link="builder-pattern" title="Builder Pattern" subtitle="Fluent chainable API for step-by-step resource construction." icon="cube" >}}
  {{< card link="working-with-primitives" title="Working with Primitives" subtitle="FHIR primitive types, pointers, the Decimal type, and extension elements." icon="variable" >}}
{{< /cards >}}

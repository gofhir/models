---
title: "Examples"
linkTitle: "Examples"
description: "Practical code examples for building FHIR resources with gofhir/models, from common data types to production patterns."
weight: 7
---

This section provides practical, runnable Go code examples that demonstrate how to use the `gofhir/models` library in real-world scenarios. The examples progress from common construction patterns to full production use cases.

## Topics

{{< cards >}}
  {{< card link="common-patterns" title="Common Patterns" subtitle="Building patients, observations, bundles, and working with CodeableConcept, Coding, and the helpers package." icon="code" >}}
  {{< card link="real-world-usage" title="Real-World Usage" subtitle="Production patterns including HTTP handlers, API response parsing, format conversion, and resource routing." icon="server" >}}
{{< /cards >}}

## Quick Start

If you are new to the library, here is the simplest possible example to get started:

```go
package main

import (
    "fmt"
    "github.com/gofhir/models/r4/v2"
)

func main() {
    // Create a Patient using the builder
    patient := r4.NewPatientBuilder().
        SetId("hello-fhir").
        SetActive(true).
        AddName(r4.HumanName{
            Family: ptrTo("World"),
            Given:  r4.PtrSlice("Hello"),
        }).
        Build()

    // Serialize to JSON
    data, _ := r4.Marshal(patient)
    fmt.Println(string(data))
}

func ptrTo[T any](v T) *T {
    return &v
}
```

Output:

```json
{
  "resourceType": "Patient",
  "id": "hello-fhir",
  "active": true,
  "name": [{"family": "World", "given": ["Hello"]}]
}
```

All examples in this section use the `r4` package. The same patterns apply to `r4b` and `r5`.

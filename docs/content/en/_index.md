---
title: "FHIR Models for Go"
description: "Type-safe Go structs for FHIR R4, R4B, and R5 resources with builders, serialization, and FHIRPath model metadata."
layout: hextra-home
---

<div class="hx:text-center hx:mt-24 hx:mb-6">
{{< hextra/hero-badge >}}
  <span>Open Source</span>
  {{< icon name="github" attributes="height=14" >}}
{{< /hextra/hero-badge >}}
</div>

<div class="hx:mt-6 hx:mb-6">
{{< hextra/hero-headline >}}
  FHIR Models for Go
{{< /hextra/hero-headline >}}
</div>

<div class="hx:mb-12">
{{< hextra/hero-subtitle >}}
  Type-safe Go structs for every FHIR R4, R4B, and R5 resource.&nbsp;<br class="sm:hx:block hx:hidden" />Build, serialize, and integrate with fluent builders and full JSON/XML support.
{{< /hextra/hero-subtitle >}}
</div>

<div class="hx:mb-6">
{{< hextra/hero-button text="Get Started" link="docs/getting-started" >}}
{{< hextra/hero-button text="View on GitHub" link="https://github.com/gofhir/models" style="alt" >}}
</div>

<div class="hx:mt-6"></div>

{{< hextra/feature-grid >}}
  {{< hextra/feature-card
    title="All FHIR Versions"
    icon="collection"
    subtitle="Complete Go structs for every resource and data type in FHIR R4 (4.0.1), R4B (4.3.0), and R5 (5.0.0). Each version is an independent Go module with its own release cycle."
  >}}
  {{< hextra/feature-card
    title="Three Construction Patterns"
    icon="puzzle"
    subtitle="Create FHIR resources your way: direct struct literals, fluent builder chains, or functional options. Every pattern produces the same type-safe structs."
  >}}
  {{< hextra/feature-card
    title="JSON & XML Serialization"
    icon="code"
    subtitle="Serialize and deserialize resources with full FHIR conformance. HTML-safe JSON marshaling preserves narrative XHTML, and polymorphic contained resources are handled automatically."
  >}}
{{< /hextra/feature-grid >}}

## Quick Start

{{< callout type="info" >}}
  Requires **Go 1.23** or later.
{{< /callout >}}

Install the package for the FHIR version you need:

```shell
go get github.com/gofhir/models/r4
```

Create a Patient resource and serialize it to JSON:

```go
package main

import (
    "fmt"

    "github.com/gofhir/models/r4"
)

func main() {
    patient := r4.NewPatient(
        r4.WithPatientId("example"),
        r4.WithPatientActive(true),
        r4.WithPatientGender(r4.AdministrativeGenderMale),
    )

    data, _ := r4.Marshal(patient)
    fmt.Println(string(data))
}
```

Output:

```json
{"resourceType":"Patient","id":"example","active":true,"gender":"male"}
```

{{< hextra/hero-button text="Read the full guide" link="docs/getting-started" >}}

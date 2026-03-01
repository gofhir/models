---
title: "Serialization"
linkTitle: "Serialization"
description: "JSON and XML serialization support for FHIR resources in the gofhir/models library."
weight: 3
---

The `gofhir/models` library provides comprehensive serialization support for FHIR resources in both JSON and XML formats. Every generated resource struct implements the standard Go marshaling interfaces, and the library also provides custom functions for FHIR-specific serialization requirements.

## Serialization Overview

FHIR resources can be exchanged in two wire formats: JSON and XML. The `gofhir/models` library supports both, with careful attention to the FHIR specification's requirements for each format.

### JSON Serialization

All resource structs implement `json.Marshaler` and `json.Unmarshaler` from Go's standard `encoding/json` package. This means you can use `json.Marshal()` and `json.Unmarshal()` directly with any resource type.

```go
import (
    "encoding/json"
    "github.com/gofhir/models/r4"
)

patient := &r4.Patient{
    ResourceType: "Patient",
    Id:           ptrTo("123"),
}
data, err := json.Marshal(patient)
```

In addition, the library provides `r4.Marshal()` and `r4.MarshalIndent()` functions that solve a specific problem with Go's standard JSON encoder: HTML entity escaping. These custom functions preserve HTML content in FHIR narrative fields exactly as the FHIR specification requires.

### XML Serialization

XML serialization is handled through dedicated helper functions in the `xml_helpers.go` module. The library provides `MarshalResourceXML()`, `MarshalResourceXMLIndent()`, and `UnmarshalResourceXML()` for working with the FHIR XML format, including proper namespace handling and the FHIR convention of encoding primitives as `<name value="..."/>` attributes.

### Polymorphic Deserialization

When working with raw FHIR data where the resource type is not known at compile time, the library provides a resource registry with functions like `UnmarshalResource()`, `GetResourceType()`, and `NewResource()` that enable dynamic dispatch to the correct Go struct based on the `resourceType` field.

## Topics

{{< cards >}}
  {{< card link="json-marshaling" title="JSON Marshaling" subtitle="Standard encoding/json compatibility with MarshalJSON and UnmarshalJSON." >}}
  {{< card link="xml-marshaling" title="XML Marshaling" subtitle="FHIR XML format with namespace handling and primitive encoding." >}}
  {{< card link="custom-marshal" title="Custom Marshal" subtitle="HTML-safe JSON output with r4.Marshal() for FHIR narrative content." >}}
  {{< card link="polymorphic-deserialization" title="Polymorphic Deserialization" subtitle="Resource registry for dynamic type resolution from raw JSON or XML." >}}
{{< /cards >}}

## Quick Comparison

| Method | HTML-Safe | Indented | Use Case |
|--------|-----------|----------|----------|
| `json.Marshal()` | No | No | General JSON output without HTML content |
| `json.MarshalIndent()` | No | Yes | Debug/display output without HTML content |
| `r4.Marshal()` | Yes | No | Production FHIR JSON output |
| `r4.MarshalIndent()` | Yes | Yes | Human-readable FHIR JSON output |
| `r4.MarshalResourceXML()` | N/A | No | Compact FHIR XML output |
| `r4.MarshalResourceXMLIndent()` | N/A | Yes | Human-readable FHIR XML output |

{{< callout type="info" >}}
For production systems exchanging FHIR resources, use `r4.Marshal()` instead of `json.Marshal()` to ensure narrative XHTML content is preserved correctly. See the [Custom Marshal](custom-marshal) page for details.
{{< /callout >}}

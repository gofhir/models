---
title: "Resource Interfaces"
linkTitle: "Resource Interfaces"
description: "The Resource and DomainResource interfaces that enable generic handling of all FHIR resource types."
weight: 1
---

The `gofhir/models` library defines two core interfaces that mirror the FHIR type hierarchy. These interfaces enable generic resource handling, polymorphic collections, and type-safe server infrastructure without resorting to `interface{}`.

## Resource Interface

The `Resource` interface is implemented by every FHIR resource type. It provides access to the base fields defined on the FHIR `Resource` abstract type.

```go
type Resource interface {
    GetResourceType() string
    GetId() *string
    SetId(string)
    GetMeta() *Meta
    SetMeta(*Meta)
}
```

### Methods

| Method | Return Type | Description |
|--------|-------------|-------------|
| `GetResourceType()` | `string` | Returns the FHIR resource type name (e.g., `"Patient"`, `"Observation"`) |
| `GetId()` | `*string` | Returns the resource's logical ID, or `nil` if not set |
| `SetId(string)` | -- | Sets the resource's logical ID |
| `GetMeta()` | `*Meta` | Returns the resource's metadata (version, lastUpdated, profiles, tags, security labels) |
| `SetMeta(*Meta)` | -- | Sets the resource's metadata |

### Example

```go
import "github.com/gofhir/models/r4"

func printResourceInfo(res r4.Resource) {
    fmt.Printf("Type: %s\n", res.GetResourceType())
    if id := res.GetId(); id != nil {
        fmt.Printf("ID: %s\n", *id)
    }
    if meta := res.GetMeta(); meta != nil && meta.VersionId != nil {
        fmt.Printf("Version: %s\n", *meta.VersionId)
    }
}

// Works with any resource type
patient := r4.NewPatient(r4.WithPatientId("p-123"))
printResourceInfo(patient) // Type: Patient, ID: p-123

obs := r4.NewObservation(r4.WithObservationId("obs-456"))
printResourceInfo(obs) // Type: Observation, ID: obs-456
```

## DomainResource Interface

The `DomainResource` interface extends `Resource` with fields from the FHIR `DomainResource` abstract type. It adds access to narrative text, contained resources, and extensions.

```go
type DomainResource interface {
    Resource
    GetText() *Narrative
    SetText(*Narrative)
    GetContained() []Resource
    GetExtension() []Extension
    GetModifierExtension() []Extension
}
```

### Methods

| Method | Return Type | Description |
|--------|-------------|-------------|
| `GetText()` | `*Narrative` | Returns the human-readable XHTML narrative |
| `SetText(*Narrative)` | -- | Sets the human-readable narrative |
| `GetContained()` | `[]Resource` | Returns the list of contained (inline) resources |
| `GetExtension()` | `[]Extension` | Returns standard extensions |
| `GetModifierExtension()` | `[]Extension` | Returns modifier extensions that change the meaning of the resource |

### Example

```go
import "github.com/gofhir/models/r4"

func extractNarrative(res r4.DomainResource) string {
    if text := res.GetText(); text != nil && text.Div != nil {
        return *text.Div
    }
    return ""
}

func listContainedTypes(res r4.DomainResource) []string {
    var types []string
    for _, contained := range res.GetContained() {
        types = append(types, contained.GetResourceType())
    }
    return types
}
```

## Which Resources Implement Which Interface

In FHIR R4, all 148 resource types implement the `Resource` interface. Most of them also implement `DomainResource`. The exceptions are the three infrastructure resources that inherit directly from `Resource` rather than `DomainResource`:

| Resource | Implements `Resource` | Implements `DomainResource` | Reason |
|----------|:---------------------:|:---------------------------:|--------|
| `Bundle` | Yes | No | Container for other resources, not a domain concept |
| `Binary` | Yes | No | Raw binary content, no narrative or extensions |
| `Parameters` | Yes | No | Operation input/output container |
| All others (~145) | Yes | Yes | Standard domain resources |

### Type Assertion

You can use Go type assertions to check whether a resource is a `DomainResource`:

```go
func processResource(res r4.Resource) {
    fmt.Printf("Processing %s/%s\n", res.GetResourceType(), safeId(res))

    if dr, ok := res.(r4.DomainResource); ok {
        // This is a DomainResource -- we can access text, contained, extensions
        if text := dr.GetText(); text != nil {
            fmt.Println("Has narrative text")
        }
        if exts := dr.GetExtension(); len(exts) > 0 {
            fmt.Printf("Has %d extensions\n", len(exts))
        }
    }
}

func safeId(res r4.Resource) string {
    if id := res.GetId(); id != nil {
        return *id
    }
    return "<no id>"
}
```

## Generic Collections

The interfaces make it straightforward to work with heterogeneous collections of resources:

```go
func storeResources(resources []r4.Resource) {
    for _, res := range resources {
        data, err := r4.Marshal(res)
        if err != nil {
            log.Printf("Failed to marshal %s: %v", res.GetResourceType(), err)
            continue
        }
        // Store data keyed by type and ID
        key := fmt.Sprintf("%s/%s", res.GetResourceType(), safeId(res))
        store.Put(key, data)
    }
}
```

{{< callout type="info" >}}
The `Resource` and `DomainResource` interfaces are generated from the FHIR specification and are identical across R4, R4B, and R5 in terms of their method signatures. The concrete types they reference (`Meta`, `Narrative`, `Extension`) are version-specific, so `r4.Resource` and `r5.Resource` are distinct interface types.
{{< /callout >}}

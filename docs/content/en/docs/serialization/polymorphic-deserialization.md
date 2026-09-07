---
title: "Polymorphic Deserialization"
linkTitle: "Polymorphic Deserialization"
description: "Resource registry functions for dynamic type resolution from raw FHIR JSON and XML data."
weight: 4
---

When working with FHIR data where the resource type is not known at compile time -- such as reading from a database, processing Bundle entries, or receiving arbitrary FHIR payloads from an API -- the library provides a resource registry that enables dynamic dispatch to the correct Go struct.

## Resource Registry Functions

The resource registry is defined in `registry.go` and provides five functions for working with FHIR resource types at runtime.

### UnmarshalResource

```go
func UnmarshalResource(data []byte) (Resource, error)
```

Deserializes raw JSON bytes into the correct concrete resource type. It first peeks at the `resourceType` field to determine the type, creates an empty instance via the factory registry, and then unmarshals the full JSON into that instance.

```go
package main

import (
    "fmt"
    "log"

    "github.com/gofhir/models/r4/v2"
)

func main() {
    jsonData := []byte(`{
        "resourceType": "Patient",
        "id": "example-1",
        "gender": "female",
        "birthDate": "1985-03-22"
    }`)

    resource, err := r4.UnmarshalResource(jsonData)
    if err != nil {
        log.Fatal(err)
    }

    // Use type assertion to access type-specific fields
    if patient, ok := resource.(*r4.Patient); ok {
        fmt.Println(*patient.Id)        // "example-1"
        fmt.Println(*patient.Gender)    // "female"
        fmt.Println(*patient.BirthDate) // "1985-03-22"
    }
}
```

### GetResourceType

```go
func GetResourceType(data []byte) (string, error)
```

Extracts the `resourceType` string from raw JSON without performing full deserialization. This is useful for routing, filtering, or validation before committing to a full unmarshal operation.

```go
jsonData := []byte(`{"resourceType": "Observation", "id": "obs-1"}`)

resourceType, err := r4.GetResourceType(jsonData)
if err != nil {
    log.Fatal(err)
}
fmt.Println(resourceType) // "Observation"
```

This function only parses the minimal struct needed to read the `resourceType` field, making it efficient for high-throughput scenarios where you need to inspect or route resources before deserializing them.

### NewResource

```go
func NewResource(resourceType string) (Resource, error)
```

Creates a new, empty instance of the specified resource type. Returns an error if the type name is not recognized.

```go
resource, err := r4.NewResource("Patient")
if err != nil {
    log.Fatal(err) // "unknown resource type: ..."
}

patient := resource.(*r4.Patient)
patient.Id = ptrTo("new-patient")
// resourceType needs no assignment; it is fixed by the Go type.
```

This function is the foundation for the other registry functions. It looks up the type name in the internal `resourceFactories` map and calls the corresponding factory function.

### IsKnownResourceType

```go
func IsKnownResourceType(resourceType string) bool
```

Returns `true` if the given resource type name is registered in the factory. Useful for input validation before attempting deserialization.

```go
fmt.Println(r4.IsKnownResourceType("Patient"))      // true
fmt.Println(r4.IsKnownResourceType("Observation"))   // true
fmt.Println(r4.IsKnownResourceType("FakeResource"))  // false
```

### AllResourceTypes

```go
func AllResourceTypes() []string
```

Returns a slice of all registered resource type names. The order is not guaranteed.

```go
types := r4.AllResourceTypes()
fmt.Println(len(types)) // 146 (for R4)

for _, t := range types {
    fmt.Println(t)
}
```

## A type this version does not define

A server one version ahead is the ordinary case, not an edge case: an R5 server
answering an R4 client will put resources like `InventoryItem` in a searchset
Bundle. Such a resource is **preserved**, not refused, so one unrecognised entry
does not destroy everything around it:

```go
var bundle r4.Bundle
err := json.Unmarshal(data, &bundle)   // no error

for _, entry := range bundle.Entry {
    switch res := entry.Resource.(type) {
    case *r4.Patient:
        // ...
    case *r4.UnknownResource:
        log.Printf("cannot interpret %s", res.Type)
    }
}
```

The document is kept byte for byte, so writing the Bundle back preserves the
resource complete — including the members this version has no field for. Before,
they were lost along with everything else.

```go
u := entry.Resource.(*r4.UnknownResource)
u.Type          // "InventoryItem"
u.Raw           // the original JSON
u.GetId()       // the id, which means the same in every FHIR version
```

Only `id` and `meta` are interpreted. They are what the `Resource` interface
exposes and they mean the same thing in every version, so reading them assumes
nothing about a type the library does not model.

### What is still an error

**A missing `resourceType`.** Unknown is not the same as absent: a document with
no type is not a resource in any version.

**`NewResource("Nonesuch")`.** That asks for an instance of a named type, and
handing back an empty shell would answer a question nobody asked. The fallback
belongs on the path that is decoding someone else's document.

### It cannot cross formats

Only one of `Raw` and `RawXML` is ever set, depending on how the resource
arrived. Converting between the two would mean knowing which members are
attributes, which are elements and which are primitives carrying a `value`
attribute — which is exactly what makes the type unknown. Asking for the other
format returns an error saying so rather than guessing:

```
InventoryItem was read from XML and cannot be written as JSON: converting it
would require knowing its structure, which is what makes it unknown
```

XML behaves the same way otherwise: `UnmarshalResourceXML` captures the element
and writes it back. That capture is semantic rather than byte-for-byte — it is
rebuilt from the decoder's token stream, which has no record of how an empty
element was spelled, so `<self/>` comes back as `<self></self>`. The two are the
same element in XML, and every member and its content survives.

## Type Assertion Patterns

Since `UnmarshalResource` returns the `Resource` interface, you need type assertions to access resource-specific fields. Here are common patterns:

### Single Type Assertion

When you expect a specific type:

```go
resource, err := r4.UnmarshalResource(jsonData)
if err != nil {
    log.Fatal(err)
}

patient, ok := resource.(*r4.Patient)
if !ok {
    log.Fatalf("expected Patient, got %s", resource.GetResourceType())
}
fmt.Println(*patient.Id)
```

### Type Switch

When handling multiple resource types:

```go
resource, err := r4.UnmarshalResource(jsonData)
if err != nil {
    log.Fatal(err)
}

switch r := resource.(type) {
case *r4.Patient:
    fmt.Printf("Patient: %s\n", *r.Id)
case *r4.Observation:
    fmt.Printf("Observation: %s\n", *r.Id)
case *r4.Encounter:
    fmt.Printf("Encounter: %s\n", *r.Id)
default:
    fmt.Printf("Other resource: %s\n", r.GetResourceType())
}
```

### Using the Resource Interface

For operations that apply to all resources, use the `Resource` interface methods without type assertion:

```go
resource, _ := r4.UnmarshalResource(jsonData)

// These methods are available on all resources
fmt.Println(resource.GetResourceType()) // e.g., "Patient"
fmt.Println(*resource.GetId())          // e.g., "123"

meta := resource.GetMeta()
if meta != nil {
    fmt.Println(*meta.VersionId)
}
```

The `Resource` interface provides:

```go
type Resource interface {
    GetResourceType() string
    GetId() *string
    SetId(string)
    GetMeta() *Meta
    SetMeta(*Meta)
}
```

For domain resources (resources with narrative and extensions), you can also assert to `DomainResource`:

```go
if dr, ok := resource.(r4.DomainResource); ok {
    text := dr.GetText()
    extensions := dr.GetExtension()
    contained := dr.GetContained()
    // ...
}
```

## Processing Bundle Entries

A common use case is processing entries from a FHIR Bundle. Each entry's `resource` field can contain any resource type:

```go
bundleJSON := []byte(`{
    "resourceType": "Bundle",
    "type": "searchset",
    "entry": [
        {"resource": {"resourceType": "Patient", "id": "p1"}},
        {"resource": {"resourceType": "Observation", "id": "o1"}}
    ]
}`)

var bundle r4.Bundle
if err := json.Unmarshal(bundleJSON, &bundle); err != nil {
    log.Fatal(err)
}

for _, entry := range bundle.Entry {
    if entry.Resource != nil {
        fmt.Printf("Type: %s, ID: %s\n",
            entry.Resource.GetResourceType(),
            *entry.Resource.GetId(),
        )
    }
}
```

## XML Polymorphic Deserialization

The registry also supports XML deserialization through `UnmarshalResourceXML`:

```go
xmlData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<Observation xmlns="http://hl7.org/fhir">
  <id value="obs-1"/>
  <status value="final"/>
</Observation>`)

resource, err := r4.UnmarshalResourceXML(xmlData)
if err != nil {
    log.Fatal(err)
}

obs := resource.(*r4.Observation)
fmt.Println(*obs.Id)     // "obs-1"
fmt.Println(*obs.Status) // "final"
```

`UnmarshalResourceXML` reads the root element name (e.g., `<Observation>`) to determine the resource type, then delegates to the same `NewResource` factory used by JSON deserialization.

## Routing Pattern

Combine `GetResourceType` with `NewResource` for efficient resource routing:

```go
func handleFHIRResource(data []byte) error {
    resourceType, err := r4.GetResourceType(data)
    if err != nil {
        return fmt.Errorf("cannot determine resource type: %w", err)
    }

    if !r4.IsKnownResourceType(resourceType) {
        return fmt.Errorf("unsupported resource type: %s", resourceType)
    }

    resource, err := r4.UnmarshalResource(data)
    if err != nil {
        return fmt.Errorf("failed to unmarshal %s: %w", resourceType, err)
    }

    switch resourceType {
    case "Patient":
        return processPatient(resource.(*r4.Patient))
    case "Observation":
        return processObservation(resource.(*r4.Observation))
    default:
        return processGenericResource(resource)
    }
}
```

{{< callout type="info" >}}
The resource registry is generated at code-generation time and includes all resource types defined in the corresponding FHIR version. For R4, this includes 146 resource types. The R4B and R5 packages each have their own registries with the resource types defined in those versions.
{{< /callout >}}

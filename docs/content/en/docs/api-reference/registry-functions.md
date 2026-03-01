---
title: "Registry Functions"
linkTitle: "Registry Functions"
description: "Factory, deserialization, and introspection functions for dynamic resource handling from the registry."
weight: 2
---

The registry provides a set of functions for working with FHIR resources dynamically -- when the resource type is not known at compile time. These functions are powered by an internal factory map that associates each resource type name with a constructor function.

All registry functions are exported from the version package (e.g., `r4.NewResource`, `r4.UnmarshalResource`).

## NewResource

Creates a new, empty instance of the specified resource type.

### Signature

```go
func NewResource(resourceType string) (Resource, error)
```

### Parameters

- `resourceType` -- The FHIR resource type name, e.g., `"Patient"`, `"Observation"`. Case-sensitive.

### Returns

- A `Resource` interface pointing to a newly allocated, zero-valued struct of the requested type.
- An error if the resource type name is not recognized.

### Example

```go
import "github.com/gofhir/models/r4"

res, err := r4.NewResource("Patient")
if err != nil {
    log.Fatal(err)
}
res.SetId("new-patient-1")

// Type-assert if you need the concrete type
patient := res.(*r4.Patient)
fmt.Println(patient.GetResourceType()) // "Patient"
```

---

## UnmarshalResource

Deserializes a JSON byte slice into the correct resource struct. It first peeks at the `resourceType` field in the JSON to determine the type, then unmarshals the full payload into the appropriate Go struct.

### Signature

```go
func UnmarshalResource(data []byte) (Resource, error)
```

### Parameters

- `data` -- A JSON byte slice containing a FHIR resource with a `resourceType` field.

### Returns

- A `Resource` interface pointing to the fully populated struct.
- An error if the JSON is invalid, the `resourceType` field is missing, or the type is not recognized.

### Example

```go
import "github.com/gofhir/models/r4"

jsonData := []byte(`{
    "resourceType": "Patient",
    "id": "example",
    "name": [{"family": "Smith", "given": ["John"]}]
}`)

res, err := r4.UnmarshalResource(jsonData)
if err != nil {
    log.Fatal(err)
}

patient := res.(*r4.Patient)
fmt.Println(*patient.Name[0].Family) // "Smith"
```

---

## GetResourceType

Extracts the `resourceType` field from a JSON byte slice without fully deserializing the resource. This is useful for routing or validation before committing to full deserialization.

### Signature

```go
func GetResourceType(data []byte) (string, error)
```

### Parameters

- `data` -- A JSON byte slice that should contain a `resourceType` field.

### Returns

- The resource type name as a string.
- An error if the JSON is invalid or the `resourceType` field is missing or empty.

### Example

```go
import "github.com/gofhir/models/r4"

data := []byte(`{"resourceType": "Observation", "id": "123"}`)

rt, err := r4.GetResourceType(data)
if err != nil {
    log.Fatal(err)
}
fmt.Println(rt) // "Observation"

// Use for routing before full deserialization
switch rt {
case "Patient":
    handlePatient(data)
case "Observation":
    handleObservation(data)
default:
    handleGeneric(data)
}
```

---

## IsKnownResourceType

Checks whether a given resource type name is recognized by the registry.

### Signature

```go
func IsKnownResourceType(resourceType string) bool
```

### Parameters

- `resourceType` -- The resource type name to check. Case-sensitive.

### Returns

- `true` if the type is a known FHIR resource for this version.
- `false` otherwise.

### Example

```go
import "github.com/gofhir/models/r4"

r4.IsKnownResourceType("Patient")      // true
r4.IsKnownResourceType("Observation")  // true
r4.IsKnownResourceType("HumanName")    // false (data type, not a resource)
r4.IsKnownResourceType("FakeResource") // false
```

---

## AllResourceTypes

Returns a slice containing all known resource type names for this FHIR version.

### Signature

```go
func AllResourceTypes() []string
```

### Returns

- A slice of all resource type names. The order is not guaranteed.

### Example

```go
import (
    "fmt"
    "sort"
    "github.com/gofhir/models/r4"
)

types := r4.AllResourceTypes()
sort.Strings(types)

fmt.Printf("R4 defines %d resource types\n", len(types))
// R4 defines 148 resource types

for _, t := range types[:5] {
    fmt.Println(t)
}
// Account
// ActivityDefinition
// AdverseEvent
// AllergyIntolerance
// Appointment
```

{{< callout type="info" >}}
The registry is initialized at package load time from a compile-time map. All registry functions are safe for concurrent use and do not require any initialization beyond importing the package.
{{< /callout >}}

## Common Patterns

### FHIR Server Resource Router

```go
import (
    "net/http"
    "github.com/gofhir/models/r4"
)

func handleCreate(w http.ResponseWriter, req *http.Request) {
    data, _ := io.ReadAll(req.Body)

    rt, err := r4.GetResourceType(data)
    if err != nil {
        http.Error(w, "Invalid FHIR resource", http.StatusBadRequest)
        return
    }

    if !r4.IsKnownResourceType(rt) {
        http.Error(w, "Unknown resource type: "+rt, http.StatusBadRequest)
        return
    }

    resource, err := r4.UnmarshalResource(data)
    if err != nil {
        http.Error(w, "Failed to parse resource", http.StatusBadRequest)
        return
    }

    // Process the resource...
    resource.SetId(generateId())
    // Store, validate, etc.
}
```

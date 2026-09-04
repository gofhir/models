---
title: "JSON Marshaling"
linkTitle: "JSON Marshaling"
description: "Standard encoding/json compatibility for FHIR resource serialization and deserialization."
weight: 1
---

All resource and data type structs in the `gofhir/models` library implement Go's standard `json.Marshaler` and `json.Unmarshaler` interfaces. This means you can use `encoding/json` directly with any FHIR type, and the library integrates seamlessly with existing Go code, HTTP handlers, and third-party libraries that expect standard JSON marshaling behavior.

## MarshalJSON and UnmarshalJSON

Every generated resource struct (such as `Patient`, `Observation`, `Bundle`) implements custom `MarshalJSON()` and `UnmarshalJSON()` methods. These methods handle FHIR-specific concerns such as:

- Serializing the `resourceType` discriminator field
- Marshaling polymorphic contained resources
- Handling the `value[x]` choice type pattern
- Encoding primitive extension elements (`_fieldName`)

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"

    "github.com/gofhir/models/r4/v2"
)

func ptrTo[T any](v T) *T {
    return &v
}

func main() {
    // Create a Patient resource
    patient := &r4.Patient{
        Id:           ptrTo("example-123"),
        Active:       ptrTo(true),
        Name: []r4.HumanName{
            {
                Use:    ptrTo(r4.NameUseOfficial),
                Family: ptrTo("Smith"),
                Given:  r4.PtrSlice("John", "Michael"),
            },
        },
        Gender:    ptrTo(r4.AdministrativeGenderMale),
        BirthDate: ptrTo("1990-01-15"),
    }

    // Marshal to JSON using the standard library
    data, err := json.Marshal(patient)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(string(data))
}
```

The output will be a compact JSON representation conforming to the FHIR JSON specification:

```json
{"resourceType":"Patient","id":"example-123","active":true,"name":[{"use":"official","family":"Smith","given":["John","Michael"]}],"gender":"male","birthDate":"1990-01-15"}
```

## JSON Tags and omitempty

All struct fields use appropriate `json` struct tags with `omitempty` for optional fields. This ensures that absent (nil) fields are omitted from the JSON output, producing clean and spec-compliant output.

```go
type Patient struct {
    ResourceType patientTypeMarker   `json:"resourceType"`  // zero-size; emits "Patient"
    Id           *string             `json:"id,omitempty"`
    Meta         *Meta               `json:"meta,omitempty"`
    Active       *bool               `json:"active,omitempty"`
    Name         []HumanName         `json:"name,omitempty"`
    Gender       *AdministrativeGender `json:"gender,omitempty"`
    BirthDate    *string             `json:"birthDate,omitempty"`
    // ... additional fields
}
```

`resourceType` does not use `omitempty`, so it is always present in the output. It is carried by a zero-size marker type rather than a string, which is why nothing needs to assign it — read it with `GetResourceType()`.

## Round-Trip Fidelity

The library aims for round-trip fidelity: marshaling a resource to JSON and then unmarshaling it back produces an identical struct. This is critical for FHIR systems that need to store and retrieve resources without data loss.

Measured over the official example corpora, 8751 of 8758 JSON files (99.9%) survive a round-trip unchanged. Classifying the 7 that do not, by every difference present rather than only the first:

| Defect | Files | Detected? | Effect |
|---|---:|---|---|
| Extensions on primitives dropped | 6 | silent | Where no companion field exists at all — primitives inside backbone elements, such as `_text` on `Questionnaire.item`. All 6 measured cases are backbone elements |
| Not a FHIR resource | 1 | error | `package-min-ver.json` is a package manifest with no `resourceType`; corpus noise, not a defect |

Regenerate these counts with `cd conformance && go test . -update-known` and read the lists.

Extensions on primitives at resource and datatype level mostly survive, but not universally — do not treat any of them as guaranteed until these lists shrink.

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"

    "github.com/gofhir/models/r4/v2"
)

func ptrTo[T any](v T) *T {
    return &v
}

func main() {
    // Create the original resource
    original := &r4.Patient{
        Id:           ptrTo("123"),
        Active:       ptrTo(true),
        Gender:       ptrTo(r4.AdministrativeGenderFemale),
    }

    // Marshal to JSON
    data, err := json.Marshal(original)
    if err != nil {
        log.Fatal(err)
    }

    // Unmarshal back to a struct
    var decoded r4.Patient
    if err := json.Unmarshal(data, &decoded); err != nil {
        log.Fatal(err)
    }

    // Verify round-trip fidelity
    fmt.Println(decoded.GetResourceType()) // "Patient"
    fmt.Println(*decoded.Id)             // "123"
    fmt.Println(*decoded.Active)         // true
    fmt.Println(*decoded.Gender)         // "female"
}
```

## Indented Output

For debugging or human-readable output, use `json.MarshalIndent()`:

```go
data, err := json.MarshalIndent(patient, "", "  ")
if err != nil {
    log.Fatal(err)
}
fmt.Println(string(data))
```

This produces formatted output:

```json
{
  "resourceType": "Patient",
  "id": "123",
  "active": true,
  "gender": "female"
}
```

## Unmarshaling from External Sources

When receiving FHIR JSON from an external API or file, unmarshal directly into the target struct:

```go
jsonData := []byte(`{
    "resourceType": "Patient",
    "id": "external-1",
    "name": [
        {
            "use": "official",
            "family": "Doe",
            "given": ["Jane"]
        }
    ],
    "gender": "female",
    "birthDate": "1985-03-22"
}`)

var patient r4.Patient
if err := json.Unmarshal(jsonData, &patient); err != nil {
    log.Fatal(err)
}

fmt.Println(*patient.Id)             // "external-1"
fmt.Println(*patient.Name[0].Family) // "Doe"
fmt.Println(*patient.Gender)         // "female"
```

## Integration with net/http

Because the structs implement the standard marshaling interfaces, they work directly with `json.NewEncoder` and `json.NewDecoder` for HTTP handlers:

```go
func handleGetPatient(w http.ResponseWriter, r *http.Request) {
    patient := &r4.Patient{
        Id:           ptrTo("http-example"),
    }

    w.Header().Set("Content-Type", "application/fhir+json")
    json.NewEncoder(w).Encode(patient)
}

func handleCreatePatient(w http.ResponseWriter, r *http.Request) {
    var patient r4.Patient
    if err := json.NewDecoder(r.Body).Decode(&patient); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    // Process the patient...
}
```

{{< callout type="info" >}}
When serving FHIR resources over HTTP, consider using `r4.Marshal()` instead of `json.NewEncoder` to avoid HTML escaping in narrative content. See the [Custom Marshal](../custom-marshal) page for details.
{{< /callout >}}

## Contained Resources

The library handles polymorphic contained resources during JSON serialization. The field's type is `ContainedList`, a named `[]Resource`: assignment from a `[]Resource`, `range`, `append`, indexing and passing to a `func([]Resource)` all work as they would on a plain slice, and `GetContained()` still returns `[]Resource`. The named type is what carries the `UnmarshalJSON` that dispatches each element, since `encoding/json` cannot build an interface value on its own.

```go
patient := &r4.Patient{
    Id:           ptrTo("with-contained"),
    Contained: []r4.Resource{
        &r4.Organization{
            Id:           ptrTo("org-1"),
            Name:         ptrTo("Example Hospital"),
        },
    },
    ManagingOrganization: &r4.Reference{
        Reference: ptrTo("#org-1"),
    },
}

data, _ := json.MarshalIndent(patient, "", "  ")
fmt.Println(string(data))
```

During unmarshaling, the `UnmarshalJSON` method reads the `resourceType` field from each contained entry and creates the appropriate concrete type.

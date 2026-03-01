---
title: "Quick Start"
linkTitle: "Quick Start"
description: "Create, serialize, and deserialize FHIR resources with working Go code examples."
weight: 2
---

This guide demonstrates the most common operations: creating a resource, marshaling it to JSON, and unmarshaling it back. All examples use the R4 package, but the API is identical for R4B and R5.

## 1. Create a Patient with a Struct Literal

The most direct way to create a FHIR resource is by initializing the struct fields. Optional FHIR fields are represented as Go pointers, so you need a helper function or the address-of operator to set them.

```go
package main

import (
    "fmt"

    "github.com/gofhir/models/r4"
)

// ptrTo is a generic helper that returns a pointer to the given value.
func ptrTo[T any](v T) *T {
    return &v
}

func main() {
    patient := r4.Patient{
        ResourceType: "Patient",
        Id:           ptrTo("123"),
        Active:       ptrTo(true),
        Name: []r4.HumanName{
            {Family: ptrTo("Smith"), Given: []string{"John"}},
        },
    }

    data, _ := r4.Marshal(&patient)
    fmt.Println(string(data))
}
```

Output:

```json
{"resourceType":"Patient","id":"123","active":true,"name":[{"family":"Smith","given":["John"]}]}
```

## 2. Create a Patient with Functional Options

Functional options eliminate pointer boilerplate. Each field has a corresponding `With<Resource><Field>()` function that sets the value and handles pointer wrapping internally.

```go
patient := r4.NewPatient(
    r4.WithPatientId("patient-123"),
    r4.WithPatientActive(true),
    r4.WithPatientGender(r4.AdministrativeGenderMale),
    r4.WithPatientBirthDate("1990-01-15"),
)
```

You can also add complex nested fields like names and identifiers:

```go
use := r4.NameUseOfficial
family := "Smith"

patient := r4.NewPatient(
    r4.WithPatientId("patient-456"),
    r4.WithPatientName(r4.HumanName{
        Use:    &use,
        Family: &family,
        Given:  []string{"John"},
    }),
    r4.WithPatientIdentifier(r4.Identifier{
        System: ptrTo("http://hospital.example.org/mrn"),
        Value:  ptrTo("12345"),
    }),
)
```

## 3. Create a Patient with the Fluent Builder

The builder pattern provides a chainable API. Start with `New<Resource>Builder()`, set fields with `Set` and `Add` methods, and finish with `.Build()`.

```go
family := "Garcia"
use := r4.NameUseOfficial

patient := r4.NewPatientBuilder().
    SetId("patient-789").
    SetActive(true).
    SetGender(r4.AdministrativeGenderFemale).
    SetBirthDate("1985-06-20").
    AddName(r4.HumanName{
        Use:    &use,
        Family: &family,
        Given:  []string{"Maria"},
    }).
    Build()
```

## 4. Marshal to JSON

Use `r4.Marshal` instead of `json.Marshal`. The library function disables HTML escaping so that FHIR narrative XHTML in `text.div` fields is preserved correctly. The `resourceType` field is always injected automatically.

```go
data, err := r4.Marshal(patient)
if err != nil {
    log.Fatal(err)
}
fmt.Println(string(data))
```

For pretty-printed output, use `MarshalIndent`:

```go
data, err := r4.MarshalIndent(patient, "", "  ")
if err != nil {
    log.Fatal(err)
}
fmt.Println(string(data))
```

Output:

```json
{
  "resourceType": "Patient",
  "id": "patient-789",
  "active": true,
  "name": [
    {
      "use": "official",
      "family": "Garcia",
      "given": ["Maria"]
    }
  ],
  "gender": "female",
  "birthDate": "1985-06-20"
}
```

## 5. Unmarshal from JSON

### Known Resource Type

If you know the resource type at compile time, unmarshal directly into the struct:

```go
var patient r4.Patient
err := json.Unmarshal(data, &patient)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Patient ID: %s\n", *patient.Id)
```

### Unknown Resource Type

When the resource type is not known in advance (for example, when reading from a FHIR server response), use `UnmarshalResource`. It inspects the `resourceType` field and returns the correct Go struct behind the `Resource` interface:

```go
resource, err := r4.UnmarshalResource(data)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Type: %s\n", resource.GetResourceType())

// Type-assert to access resource-specific fields
if patient, ok := resource.(*r4.Patient); ok {
    fmt.Printf("Patient ID: %s\n", *patient.Id)
}
```

You can also peek at the resource type without full deserialization:

```go
resourceType, err := r4.GetResourceType(data)
if err != nil {
    log.Fatal(err)
}
fmt.Println(resourceType) // "Patient"
```

## 6. Create an Observation

Here is a more complete example showing an Observation with a code, value, and category:

```go
codeSystem := "http://loinc.org"
codeCode := "8867-4"
codeDisplay := "Heart rate"

obs := r4.NewObservation(
    r4.WithObservationId("obs-123"),
    r4.WithObservationStatus(r4.ObservationStatusFinal),
    r4.WithObservationCode(r4.CodeableConcept{
        Coding: []r4.Coding{
            {System: &codeSystem, Code: &codeCode, Display: &codeDisplay},
        },
    }),
    r4.WithObservationEffectiveDateTime("2024-01-15T10:30:00Z"),
)

data, _ := r4.Marshal(obs)
fmt.Println(string(data))
```

## Next Steps

- Learn about [FHIR Versions](../fhir-versions) and how the three packages differ.
- Explore [Resource Construction](../../resource-construction) patterns in depth.

---
title: "Struct Literals"
linkTitle: "Struct Literals"
description: "Create FHIR resources using direct Go struct initialization with full control over every field."
weight: 1
---

Struct literals give you the most direct control over resource creation. You initialize the Go struct fields explicitly, which is familiar to any Go developer and provides full compile-time type checking.

## Basic Example

```go
package main

import (
    "fmt"

    "github.com/gofhir/models/r4"
)

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

{{< callout type="info" >}}
The `ResourceType` field is automatically injected during JSON marshaling. You can set it explicitly for clarity, but it is not required.
{{< /callout >}}

## Handling Pointers

In FHIR, most primitive fields are optional. The Go structs represent this with pointer types (`*string`, `*bool`, `*int`). You cannot take the address of a literal value directly in Go, so you need a helper pattern.

### The `ptrTo` Helper

The most common approach is a generic helper function:

```go
func ptrTo[T any](v T) *T {
    return &v
}
```

This works with any type:

```go
patient := r4.Patient{
    Id:        ptrTo("patient-1"),
    Active:    ptrTo(true),
    BirthDate: ptrTo("1990-05-15"),
    Gender:    ptrTo(r4.AdministrativeGenderMale),
}
```

### Local Variables

Alternatively, you can use local variables and take their address:

```go
id := "patient-1"
active := true
gender := r4.AdministrativeGenderFemale

patient := r4.Patient{
    Id:     &id,
    Active: &active,
    Gender: &gender,
}
```

This approach is useful when you need to reuse values or when the field type is a custom code system type.

## Complex Resources

Struct literals work well for building resources with nested data types:

```go
system := "http://hospital.example.org/mrn"
value := "MRN-12345"
use := r4.NameUseOfficial
family := "Johnson"
homeUse := r4.AddressUseHome

patient := r4.Patient{
    Id:     ptrTo("patient-complex"),
    Active: ptrTo(true),
    Identifier: []r4.Identifier{
        {
            System: &system,
            Value:  &value,
        },
    },
    Name: []r4.HumanName{
        {
            Use:    &use,
            Family: &family,
            Given:  []string{"Robert", "James"},
        },
    },
    Gender:    ptrTo(r4.AdministrativeGenderMale),
    BirthDate: ptrTo("1978-11-03"),
    Address: []r4.Address{
        {
            Use:        &homeUse,
            Line:       []string{"123 Main St", "Apt 4B"},
            City:       ptrTo("Springfield"),
            State:      ptrTo("IL"),
            PostalCode: ptrTo("62704"),
            Country:    ptrTo("US"),
        },
    },
    Telecom: []r4.ContactPoint{
        {
            System: ptrTo(r4.ContactPointSystemPhone),
            Value:  ptrTo("555-0123"),
            Use:    ptrTo(r4.ContactPointUseHome),
        },
        {
            System: ptrTo(r4.ContactPointSystemEmail),
            Value:  ptrTo("robert.johnson@example.com"),
        },
    },
}
```

## Observations with Struct Literals

```go
codeSystem := "http://loinc.org"
codeCode := "8480-6"
codeDisplay := "Systolic blood pressure"
unitSystem := "http://unitsofmeasure.org"

obs := r4.Observation{
    Status: ptrTo(r4.ObservationStatusFinal),
    Code: r4.CodeableConcept{
        Coding: []r4.Coding{
            {
                System:  &codeSystem,
                Code:    &codeCode,
                Display: &codeDisplay,
            },
        },
    },
    ValueQuantity: &r4.Quantity{
        Value:  r4.NewDecimalFromFloat64(120.0),
        Unit:   ptrTo("mmHg"),
        System: &unitSystem,
        Code:   ptrTo("mm[Hg]"),
    },
    EffectiveDateTime: ptrTo("2024-01-15T10:30:00Z"),
}
```

## When to Use Struct Literals

Struct literals are a good choice when:

- You want full visibility into every field being set.
- You are building a resource from a known, fixed set of values (such as test fixtures or seed data).
- You prefer explicit Go idioms over builder abstractions.

For situations where pointer boilerplate is cumbersome, consider the [Builder Pattern](../builder-pattern) or [Functional Options](../functional-options) instead.

---
title: "Builder Pattern"
linkTitle: "Builder Pattern"
description: "Use the fluent builder API to construct FHIR resources with chainable method calls."
weight: 2
---

The builder pattern provides a fluent, chainable API for constructing FHIR resources. Each resource type has a corresponding builder with `Set` methods for singular fields and `Add` methods for repeating fields. The builder handles pointer wrapping internally, eliminating the boilerplate required by struct literals.

## How It Works

Every resource type in the package has a builder:

1. Create a builder with `New<Resource>Builder()`.
2. Chain `Set<Field>()` calls for singular fields.
3. Chain `Add<Field>()` calls for repeating (slice) fields.
4. Call `.Build()` to get the final resource struct.

The builder returns `*<Resource>Builder` from every setter, so calls can be chained.

## Basic Example

```go
patient := r4.NewPatientBuilder().
    SetId("patient-789").
    SetActive(true).
    SetGender(r4.AdministrativeGenderFemale).
    SetBirthDate("1985-06-20").
    Build()

data, _ := r4.Marshal(patient)
fmt.Println(string(data))
```

Output:

```json
{"resourceType":"Patient","id":"patient-789","active":true,"gender":"female","birthDate":"1985-06-20"}
```

## Adding Complex Fields

For fields that contain data type structs (like `HumanName`, `Identifier`, or `Address`), use the `Add` methods. These append to the underlying slice:

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

Note that the data type structs passed to `Add` methods still use pointers for optional fields. The builder eliminates pointer boilerplate for primitive resource fields (string, bool, code types), but complex data type structs retain their standard Go representation.

## Adding Multiple Elements

Call `Add` methods multiple times to append to repeating fields:

```go
system := "http://hospital.example.org/mrn"
value1 := "MRN-001"
value2 := "MRN-002"

patient := r4.NewPatientBuilder().
    SetId("patient-multi").
    AddIdentifier(r4.Identifier{System: &system, Value: &value1}).
    AddIdentifier(r4.Identifier{System: &system, Value: &value2}).
    Build()

// patient.Identifier has 2 elements
```

## Building an Observation

The builder pattern works for all resource types. Here is an Observation with a vital sign measurement:

```go
codeSystem := "http://loinc.org"
codeCode := "8480-6"
codeDisplay := "Systolic blood pressure"
value := r4.NewDecimalFromFloat64(120.0)
unit := "mmHg"
unitSystem := "http://unitsofmeasure.org"
unitCode := "mm[Hg]"

obs := r4.NewObservationBuilder().
    SetId("obs-bp-001").
    SetStatus(r4.ObservationStatusFinal).
    SetCode(r4.CodeableConcept{
        Coding: []r4.Coding{
            {System: &codeSystem, Code: &codeCode, Display: &codeDisplay},
        },
    }).
    SetValueQuantity(r4.Quantity{
        Value:  value,
        Unit:   &unit,
        System: &unitSystem,
        Code:   &unitCode,
    }).
    SetEffectiveDateTime("2024-01-15T10:30:00Z").
    Build()
```

## JSON Round Trip

Resources built with the builder serialize and deserialize exactly like struct literals:

```go
family := "Johnson"
city := "Boston"
use := r4.AddressUseHome

original := r4.NewPatientBuilder().
    SetId("pt-json").
    SetActive(true).
    SetGender(r4.AdministrativeGenderMale).
    AddName(r4.HumanName{Family: &family, Given: []string{"Robert"}}).
    AddAddress(r4.Address{Use: &use, City: &city}).
    Build()

// Marshal
data, err := r4.Marshal(original)
if err != nil {
    log.Fatal(err)
}

// Unmarshal
var decoded r4.Patient
err = json.Unmarshal(data, &decoded)
if err != nil {
    log.Fatal(err)
}

fmt.Println(*decoded.Id)          // "pt-json"
fmt.Println(*decoded.Name[0].Family) // "Johnson"
```

## Empty Builder

Calling `Build()` without setting any fields returns a valid, empty resource:

```go
patient := r4.NewPatientBuilder().Build()
// patient.Id is nil, patient.Active is nil, patient.Name is empty
```

This is useful as a starting point when you need to conditionally populate fields.

## Available Methods

Every builder follows the same naming convention:

| Method Pattern | Purpose | Example |
|----------------|---------|---------|
| `Set<Field>(v)` | Set a singular field | `SetId("123")`, `SetActive(true)` |
| `Add<Field>(v)` | Append to a repeating field | `AddName(humanName)`, `AddIdentifier(id)` |
| `Build()` | Return the constructed resource | `Build()` |

The `Set` methods accept unwrapped values (e.g., `string` instead of `*string`) and handle pointer creation internally. The `Add` methods accept the data type struct directly and append it to the corresponding slice.

## When to Use the Builder

The builder pattern is ideal when:

- You are constructing resources step by step, possibly across multiple function calls.
- You want a fluent, readable chain of field assignments.
- You want to avoid pointer boilerplate for primitive fields.

For one-shot initialization with full control, consider [Struct Literals](../struct-literals). For composable configuration, see [Functional Options](../functional-options).

---
title: "Functional Options"
linkTitle: "Functional Options"
description: "Create FHIR resources using composable functional option functions for clean, configurable construction."
weight: 3
---

The functional options pattern uses Go functions to configure resource fields. Each resource type has a `New<Resource>(opts...)` constructor and a set of `With<Resource><Field>()` functions. This pattern, popularized by Dave Cheney and Rob Pike, produces clean call sites and makes options composable.

## How It Works

Each resource type provides:

1. An option type: `<Resource>Option` (a function that modifies the resource).
2. A constructor: `New<Resource>(opts ...<Resource>Option)` that creates a resource and applies the options.
3. Option functions: `With<Resource><Field>(value)` for each field on the resource.

```go
// PatientOption is a functional option for configuring a Patient.
type PatientOption func(*Patient)

// NewPatient creates a new Patient with the given options.
func NewPatient(opts ...PatientOption) *Patient {
    r := &Patient{}
    for _, opt := range opts {
        opt(r)
    }
    return r
}
```

## Basic Example

```go
patient := r4.NewPatient(
    r4.WithPatientId("patient-123"),
    r4.WithPatientActive(true),
    r4.WithPatientGender(r4.AdministrativeGenderMale),
    r4.WithPatientBirthDate("1990-01-15"),
)

fmt.Println(*patient.Id)     // "patient-123"
fmt.Println(*patient.Active) // true
fmt.Println(*patient.Gender) // "male"
```

The `With` functions handle pointer wrapping internally, so you pass plain values (`string`, `bool`, etc.) and the function creates the pointer for you.

## Adding Names and Identifiers

For complex data types like `HumanName` and `Identifier`, the option functions accept the struct directly. Each call appends to the slice:

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
    r4.WithPatientName(r4.HumanName{
        Family: &family,
        Given:  []string{"Johnny"},
    }),
)

// patient.Name has 2 elements
```

Adding identifiers works the same way:

```go
system := "http://hospital.example.org/mrn"
value := "12345"

patient := r4.NewPatient(
    r4.WithPatientIdentifier(r4.Identifier{
        System: &system,
        Value:  &value,
    }),
)
```

## Building an Observation

Functional options are available for every resource type. Here is an Observation with a LOINC code:

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
    r4.WithObservationSubject(r4.Reference{
        Reference: ptrTo("Patient/patient-123"),
    }),
)
```

## Composing Options

Because options are plain functions, you can store them in slices, pass them between functions, and compose them:

```go
// Define reusable option sets
func defaultPatientOptions() []r4.PatientOption {
    return []r4.PatientOption{
        r4.WithPatientActive(true),
        r4.WithPatientLanguage("en"),
    }
}

func main() {
    // Start with defaults and add specific options
    opts := defaultPatientOptions()
    opts = append(opts,
        r4.WithPatientId("patient-composed"),
        r4.WithPatientGender(r4.AdministrativeGenderFemale),
    )

    patient := r4.NewPatient(opts...)
}
```

This pattern is particularly useful in test code where you want to define base fixtures and override specific fields per test case.

## Conditional Options

You can conditionally include options using standard Go control flow:

```go
func createPatient(id string, birthDate string, deceased bool) *r4.Patient {
    opts := []r4.PatientOption{
        r4.WithPatientId(id),
        r4.WithPatientActive(true),
        r4.WithPatientBirthDate(birthDate),
    }

    if deceased {
        opts = append(opts, r4.WithPatientDeceasedBoolean(true))
    }

    return r4.NewPatient(opts...)
}
```

## Empty Resource

Calling the constructor with no options returns a valid, empty resource:

```go
patient := r4.NewPatient()
// patient.Id is nil
// patient.Active is nil
// patient.Name is empty
```

## Naming Convention

All option functions follow a consistent naming pattern:

| Pattern | Purpose | Example |
|---------|---------|---------|
| `With<Resource><Field>(v)` | Set a singular field | `WithPatientId("123")` |
| `With<Resource><Field>(v)` | Append to a repeating field | `WithPatientName(humanName)` |

For singular fields (like `Id`, `Active`, `Gender`), the option function sets the field. For repeating fields (like `Name`, `Identifier`, `Telecom`), the option function appends to the existing slice. Calling `WithPatientName` twice adds two names.

## When to Use Functional Options

Functional options are ideal when:

- You want clean, readable construction with no pointer boilerplate.
- You need to compose options from different sources (defaults, overrides, conditional logic).
- You are writing library code that accepts configuration from callers.
- You want to build test fixtures with reusable base configurations.

For full control over every field, consider [Struct Literals](../struct-literals). For step-by-step construction with a chainable API, see the [Builder Pattern](../builder-pattern).

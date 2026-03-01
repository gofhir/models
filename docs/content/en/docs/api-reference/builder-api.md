---
title: "Builder API"
linkTitle: "Builder API"
description: "Fluent builder pattern and functional options API for constructing FHIR resources."
weight: 3
---

Every resource type in `gofhir/models` provides two complementary construction APIs: a **fluent builder** and **functional options**. Both are generated automatically for all resources across R4, R4B, and R5.

## Fluent Builder Pattern

The builder pattern provides a chainable API for constructing resources step by step.

### Structure

For each resource type `<Resource>`, the library generates:

| Export | Type | Description |
|--------|------|-------------|
| `<Resource>Builder` | struct | The builder struct holding the resource under construction |
| `New<Resource>Builder()` | `*<Resource>Builder` | Constructor that creates a new builder with a zero-valued resource |
| `Set<Field>(v T)` | `*<Resource>Builder` | Sets a singular field (pointer or scalar) |
| `Add<Field>(v T)` | `*<Resource>Builder` | Appends to a repeated (slice) field |
| `Build()` | `*<Resource>` | Returns the constructed resource |

### Naming Convention

- **`Set`** is used for singular fields -- fields with a maximum cardinality of 1 (e.g., `Id`, `Gender`, `BirthDate`, `Status`).
- **`Add`** is used for repeated fields -- fields with a maximum cardinality greater than 1 (e.g., `Name`, `Identifier`, `Telecom`, `Extension`).

### Patient Example

```go
import "github.com/gofhir/models/r4"

patient := r4.NewPatientBuilder().
    SetId("patient-123").
    SetActive(true).
    SetGender(r4.AdministrativeGenderMale).
    SetBirthDate("1990-05-15").
    AddName(r4.HumanName{
        Family: ptrTo("Doe"),
        Given:  []string{"John", "Michael"},
    }).
    AddIdentifier(r4.Identifier{
        System: ptrTo("http://hospital.example.org/mrn"),
        Value:  ptrTo("MRN-12345"),
    }).
    AddTelecom(r4.ContactPoint{
        System: ptrTo(r4.ContactPointSystemPhone),
        Value:  ptrTo("+1-555-0100"),
        Use:    ptrTo(r4.ContactPointUseHome),
    }).
    AddAddress(r4.Address{
        Use:        ptrTo(r4.AddressUseHome),
        Line:       []string{"123 Main St"},
        City:       ptrTo("Springfield"),
        State:      ptrTo("IL"),
        PostalCode: ptrTo("62701"),
        Country:    ptrTo("US"),
    }).
    Build()

// Helper function for creating string pointers
func ptrTo[T any](v T) *T {
    return &v
}
```

### Observation Example

```go
import "github.com/gofhir/models/r4"

status := r4.ObservationStatusFinal
observation := r4.NewObservationBuilder().
    SetId("obs-001").
    SetStatus(status).
    SetCode(r4.CodeableConcept{
        Coding: []r4.Coding{{
            System:  ptrTo("http://loinc.org"),
            Code:    ptrTo("29463-7"),
            Display: ptrTo("Body weight"),
        }},
        Text: ptrTo("Body Weight"),
    }).
    SetSubject(r4.Reference{
        Reference: ptrTo("Patient/patient-123"),
    }).
    SetEffectiveDateTime("2024-01-15T10:30:00Z").
    SetValueQuantity(r4.Quantity{
        Value:  r4.NewDecimalFromFloat64(72.5),
        Unit:   ptrTo("kg"),
        System: ptrTo("http://unitsofmeasure.org"),
        Code:   ptrTo("kg"),
    }).
    AddCategory(r4.CodeableConcept{
        Coding: []r4.Coding{{
            System:  ptrTo("http://terminology.hl7.org/CodeSystem/observation-category"),
            Code:    ptrTo("vital-signs"),
            Display: ptrTo("Vital Signs"),
        }},
    }).
    Build()
```

## Functional Options Pattern

The functional options pattern provides a more concise syntax for creating resources in a single function call.

### Structure

For each resource type `<Resource>`, the library generates:

| Export | Type | Description |
|--------|------|-------------|
| `<Resource>Option` | `func(*<Resource>)` | The option function type |
| `New<Resource>(opts ...Option)` | `*<Resource>` | Constructor that applies all options |
| `With<Resource><Field>(v T)` | `<Resource>Option` | Option function for each field |

### Patient Example

```go
import "github.com/gofhir/models/r4"

patient := r4.NewPatient(
    r4.WithPatientId("patient-456"),
    r4.WithPatientActive(true),
    r4.WithPatientGender(r4.AdministrativeGenderFemale),
    r4.WithPatientBirthDate("1985-11-20"),
    r4.WithPatientName(r4.HumanName{
        Family: ptrTo("Smith"),
        Given:  []string{"Jane"},
    }),
    r4.WithPatientIdentifier(r4.Identifier{
        System: ptrTo("http://hospital.example.org/mrn"),
        Value:  ptrTo("MRN-67890"),
    }),
)
```

### Observation Example

```go
import "github.com/gofhir/models/r4"

observation := r4.NewObservation(
    r4.WithObservationId("obs-002"),
    r4.WithObservationStatus(r4.ObservationStatusFinal),
    r4.WithObservationCode(r4.CodeableConcept{
        Coding: []r4.Coding{{
            System:  ptrTo("http://loinc.org"),
            Code:    ptrTo("8310-5"),
            Display: ptrTo("Body temperature"),
        }},
    }),
    r4.WithObservationValueQuantity(r4.Quantity{
        Value:  r4.NewDecimalFromFloat64(37.2),
        Unit:   ptrTo("Cel"),
        System: ptrTo("http://unitsofmeasure.org"),
        Code:   ptrTo("Cel"),
    }),
)
```

## Choosing Between Patterns

| Criterion | Fluent Builder | Functional Options |
|-----------|:--------------:|:------------------:|
| Method chaining | Yes | No |
| Single expression | Multi-line chain | Single function call |
| Incremental construction | Natural fit | Less natural |
| Conditional fields | Add after builder creation | Compose option slices |
| Testing/mocking | Builder can be injected | Options can be collected |

### When to Use the Builder

The builder is well suited when you need to construct a resource incrementally, especially when some fields depend on conditions:

```go
builder := r4.NewPatientBuilder().
    SetId(id).
    SetActive(true)

if hasName {
    builder.AddName(name)
}
if hasAddress {
    builder.AddAddress(address)
}

patient := builder.Build()
```

### When to Use Functional Options

Functional options work well when constructing a resource in a single, declarative statement, or when options are composed from different sources:

```go
opts := []r4.PatientOption{
    r4.WithPatientId(id),
    r4.WithPatientActive(true),
}

if hasName {
    opts = append(opts, r4.WithPatientName(name))
}

patient := r4.NewPatient(opts...)
```

{{< callout type="info" >}}
Both the builder and functional options produce identical results. They both set fields on the same underlying struct. Choose whichever style better fits your code's readability requirements.
{{< /callout >}}

## Struct Literals

You can always construct resources directly using Go struct literals. This provides the most control and is the most familiar Go pattern:

```go
patient := &r4.Patient{
    ResourceType: "Patient",
    Id:           ptrTo("patient-789"),
    Active:       ptrTo(true),
    Gender:       ptrTo(r4.AdministrativeGenderMale),
    Name: []r4.HumanName{{
        Family: ptrTo("Johnson"),
        Given:  []string{"Robert"},
    }},
}
```

Note that when using struct literals, you must set `ResourceType` yourself. The builder and functional options set `ResourceType` automatically during JSON marshaling.

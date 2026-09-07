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
        Given:  r4.PtrSlice("Maria"),
    }).
    Build()
```

Data types have builders of their own, so the chain does not have to stop there:

```go
patient := r4.NewPatientBuilder().
    SetId("patient-789").
    SetGender(r4.AdministrativeGenderFemale).
    SetBirthDate("1985-06-20").
    AddName(r4.NewHumanNameBuilder().
        SetUse(r4.NameUseOfficial).
        SetFamily("Garcia").
        AddGiven("Maria").
        Build()).
    Build()
```

`Build()` on a data type returns a **value**, not a pointer, because that is what
every consumer takes: `Patient.Name` is `[]HumanName`, and even a pointer field
like `Range.Low` is set through `SetLow(Quantity)`, which takes the address
itself. So no dereference is needed at the call site.

Resources are the other way round — `Build()` returns `*Patient` — since they are
handled as pointers and satisfy the `Resource` interface on the pointer receiver.

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
    AddName(r4.HumanName{Family: &family, Given: r4.PtrSlice("Robert")}).
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
| `Add<Field>(v)` | Append to a repeating field | `AddName(humanName)` |
| `Set<Field>Ext(v)` | Set the extensions on a primitive | `SetBirthDateExt(element)` |
| `Add<Field>Ext(v)` | Append an extension slot to a repeating primitive | `AddGivenExt(&element)` |
| `Build()` | Return the constructed value | `Build()` |

The `Set` methods accept unwrapped values (`string` rather than `*string`) and
create the pointer internally.

### Every type has one

Builders are generated for resources, data types **and** backbone elements — 663
in R4, 681 in R4B, 834 in R5. A backbone is where most of a resource's data
actually lives, so stopping at the resource level would mean dropping into a
struct literal for `Patient.contact`, `Bundle.entry` or `Observation.component`:

```go
patient := r4.NewPatientBuilder().
    AddContact(r4.NewPatientContactBuilder().
        SetName(r4.NewHumanNameBuilder().SetFamily("Doe").Build()).
        AddTelecom(r4.NewContactPointBuilder().
            SetSystem(r4.ContactPointSystemPhone).
            SetValue("555-0100").
            Build()).
        Build()).
    Build()
```

### Extensions on primitives

A primitive carries its extensions in a companion field — `BirthDateExt` for
`birthDate`, serialized as `_birthDate`. That is the only way FHIR expresses an
extension on a primitive value, and the builder reaches it:

```go
patient := r4.NewPatientBuilder().
    SetBirthDateExt(r4.NewElementBuilder().
        AddExtension(r4.NewExtensionBuilder().
            SetUrl("http://hl7.org/fhir/StructureDefinition/data-absent-reason").
            SetValueCode("asked-declined").
            Build()).
        Build()).
    Build()

// {"resourceType":"Patient","_birthDate":{"extension":[{"url":"...","valueCode":"asked-declined"}]}}
```

Note there is no `birthDate` value in that output: the extension says *why* the
date is absent, which is the whole point.

For a **repeating** primitive the two slices are parallel by position, so
`Add<Field>Ext` attaches to the element added most recently and fills any earlier
gap with nil:

```go
name := r4.NewHumanNameBuilder().
    AddGiven("A").
    AddGiven("B").
    AddGivenExt(&ext).      // belongs to "B"
    Build()

// {"given":["A","B"],"_given":[null,{...}]}
```

Passing `nil` is meaningful there — it is a position with no extension.

### Choice fields are exclusive

A choice element holds exactly one variant, so setting one clears the others,
including the `_field` companion of a primitive variant:

```go
obs := r4.NewObservationBuilder().
    SetValueString("a").
    SetValueBoolean(true).      // clears valueString
    Build()

// {"resourceType":"Observation","valueBoolean":true}
```

Without that, a chain of setters produced a document with several variants
present, which no FHIR server accepts. A struct literal can still do it — the
guard is in the builder.

## When to Use the Builder

The builder pattern is ideal when:

- You are constructing resources step by step, possibly across multiple function calls.
- You want a fluent, readable chain of field assignments.
- You want to avoid pointer boilerplate for primitive fields.

For one-shot initialization with full control, consider [Struct Literals](../struct-literals).

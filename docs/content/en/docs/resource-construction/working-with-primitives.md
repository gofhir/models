---
title: "Working with Primitives"
linkTitle: "Working with Primitives"
description: "Understand how FHIR primitive types map to Go types, including pointers, the Decimal type, and extension elements."
weight: 4
---

FHIR defines a set of primitive data types (string, boolean, integer, decimal, date, dateTime, and others) that appear throughout every resource. This page explains how these primitives are represented in Go and how to work with them effectively.

## Primitive Type Mapping

FHIR primitive types map to Go types as follows:

| FHIR Type | Go Type | Example Value |
|-----------|---------|---------------|
| `string` | `*string` | `"John"` |
| `boolean` | `*bool` | `true` |
| `integer` | `*int` | `42` |
| `decimal` | `*Decimal` | `NewDecimalFromFloat64(98.6)` |
| `date` | `*string` | `"1990-01-15"` |
| `dateTime` | `*string` | `"2024-01-15T10:30:00Z"` |
| `instant` | `*string` | `"2024-01-15T10:30:00.000Z"` |
| `time` | `*string` | `"14:30:00"` |
| `uri` | `*string` | `"http://example.org"` |
| `url` | `*string` | `"https://example.org/fhir"` |
| `canonical` | `*string` | `"http://hl7.org/fhir/StructureDefinition/Patient"` |
| `id` | `*string` | `"patient-123"` |
| `code` | Custom type (e.g., `*AdministrativeGender`) | `AdministrativeGenderMale` |
| `base64Binary` | `*string` | Base64-encoded string |
| `positiveInt` | `*int` | `1` |
| `unsignedInt` | `*int` | `0` |

## Why Pointers?

In FHIR, most fields are optional. A missing field has a different meaning from a field set to its zero value. For example, `active: false` is different from the `active` field being absent entirely. Go pointers distinguish these cases:

- `nil` -- field is absent (omitted from JSON output)
- `&value` -- field is present with the given value

```go
// Active is absent (nil)
patient := r4.Patient{Id: ptrTo("1")}

// Active is explicitly false
patient := r4.Patient{Id: ptrTo("1"), Active: ptrTo(false)}
```

Both the builder pattern and functional options handle pointer wrapping automatically, so you only deal with raw values:

```go
// Builder -- no pointer needed
patient := r4.NewPatientBuilder().SetActive(false).Build()

// Functional options -- no pointer needed
patient := r4.NewPatient(r4.WithPatientActive(false))
```

## The Decimal Type

FHIR requires that decimal values preserve their exact textual representation. For example, `1.50` must remain `1.50` in JSON output, not `1.5`. The standard Go `float64` type loses trailing zeros, so the library provides a custom `Decimal` type.

### Creating Decimal Values

There are several ways to create a `Decimal`:

```go
// From a string -- preserves exact representation
d, err := r4.NewDecimalFromString("1.50")
// d.String() == "1.50"
// JSON output: 1.50

// From a float64 -- precision may be lost
d := r4.NewDecimalFromFloat64(1.5)
// d.String() == "1.5"
// JSON output: 1.5

// From a string, panicking on error (for constants only)
d := r4.MustDecimal("98.60")
// d.String() == "98.60"

// From an integer
d := r4.NewDecimalFromInt(100)
// d.String() == "100"

// From an int64
d := r4.NewDecimalFromInt64(9223372036854775807)
```

### Using Decimal in Resources

Decimal values appear in Quantity, Money, and other data types:

```go
obs := r4.Observation{
    Status: ptrTo(r4.ObservationStatusFinal),
    Code:   r4.CodeableConcept{ /* ... */ },
    ValueQuantity: &r4.Quantity{
        Value:  r4.NewDecimalFromFloat64(120.0),
        Unit:   ptrTo("mmHg"),
        System: ptrTo("http://unitsofmeasure.org"),
        Code:   ptrTo("mm[Hg]"),
    },
}
```

For precision-critical values (like lab results or medication dosages), always use `NewDecimalFromString`:

```go
// Preserves "1.50" exactly in JSON output
quantity := r4.Quantity{
    Value: r4.MustDecimal("1.50"),
    Unit:  ptrTo("mg"),
}
```

### Decimal Methods

The `Decimal` type provides these methods:

| Method | Returns | Description |
|--------|---------|-------------|
| `String()` | `string` | Exact textual representation |
| `Float64()` | `float64` | Numeric value (may lose precision) |
| `IsZero()` | `bool` | True if zero or empty |
| `Equal(other)` | `bool` | Numeric equality (ignores trailing zeros) |
| `MarshalJSON()` | `[]byte, error` | Emits bare JSON number preserving precision |
| `UnmarshalJSON(data)` | `error` | Parses bare JSON number preserving precision |

### Precision Preservation in JSON

The `Decimal` type marshals as a bare JSON number (not a quoted string), preserving the exact digits:

```go
d := r4.MustDecimal("1.50")
data, _ := json.Marshal(d)
fmt.Println(string(data)) // 1.50 (not 1.5 or "1.50")
```

When unmarshaling, the exact representation from the JSON input is preserved:

```go
var d r4.Decimal
json.Unmarshal([]byte("1.50"), &d)
fmt.Println(d.String()) // "1.50"
```

## Extension Elements

In FHIR, every primitive element can carry extensions. The JSON representation uses a parallel field prefixed with `_`. For example, a `birthDate` field has a corresponding `_birthDate` field for extensions.

In the Go structs, these appear as `<Field>Ext` fields of type `*Element`:

```go
type Patient struct {
    // ...
    BirthDate    *string  `json:"birthDate,omitempty"`
    BirthDateExt *Element `json:"_birthDate,omitempty"`
    // ...
}
```

The `Element` type holds an optional `Id` and a slice of `Extension` values:

```go
type Element struct {
    Id        *string     `json:"id,omitempty"`
    Extension []Extension `json:"extension,omitempty"`
}
```

### Setting Extensions on Primitives

To add an extension to a primitive field, set both the value and its extension element:

```go
patient := r4.Patient{
    BirthDate: ptrTo("1990-01"),
    BirthDateExt: &r4.Element{
        Extension: []r4.Extension{
            {
                Url:           "http://hl7.org/fhir/StructureDefinition/data-absent-reason",
                ValueCode:     ptrTo("masked"),
            },
        },
    },
}
```

This produces JSON with the `_birthDate` field:

```json
{
  "resourceType": "Patient",
  "birthDate": "1990-01",
  "_birthDate": {
    "extension": [
      {
        "url": "http://hl7.org/fhir/StructureDefinition/data-absent-reason",
        "valueCode": "masked"
      }
    ]
  }
}
```

### Extension-Only Primitives

FHIR allows a primitive to have only an extension with no value. In Go, set the value field to `nil` and populate only the extension field:

```go
patient := r4.Patient{
    BirthDate: nil,
    BirthDateExt: &r4.Element{
        Extension: []r4.Extension{
            {
                Url:       "http://hl7.org/fhir/StructureDefinition/data-absent-reason",
                ValueCode: ptrTo("unknown"),
            },
        },
    },
}
```

## Code System Types

FHIR code elements are represented as typed string constants rather than plain strings. This provides compile-time validation of code values:

```go
// AdministrativeGender is a typed string
type AdministrativeGender string

const (
    AdministrativeGenderMale    AdministrativeGender = "male"
    AdministrativeGenderFemale  AdministrativeGender = "female"
    AdministrativeGenderOther   AdministrativeGender = "other"
    AdministrativeGenderUnknown AdministrativeGender = "unknown"
)
```

Use the predefined constants for type safety:

```go
patient := r4.NewPatient(
    r4.WithPatientGender(r4.AdministrativeGenderMale),
)
```

The compiler will reject invalid values, catching errors that would only surface at runtime with plain strings.

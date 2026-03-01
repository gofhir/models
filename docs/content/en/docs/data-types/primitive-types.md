---
title: "Primitive Types"
linkTitle: "Primitive Types"
description: "Complete FHIR-to-Go type mapping for primitive data types, pointer semantics, and extension elements."
weight: 1
---

FHIR defines a set of primitive data types that serve as the building blocks for all resource fields. The `gofhir/models` library maps each FHIR primitive type to an appropriate Go type, using pointer semantics to represent optionality.

## FHIR-to-Go Type Mapping

The following table shows the complete mapping from FHIR primitive types to their Go representations:

| FHIR Type | Go Type | Notes |
|-----------|---------|-------|
| `string` | `*string` | General-purpose text |
| `uri` | `*string` | URI/IRI reference |
| `url` | `*string` | Absolute URL |
| `canonical` | `*string` | Canonical URL to a FHIR resource definition |
| `id` | `*string` | Resource logical ID (1-64 characters, [A-Za-z0-9\-.]) |
| `oid` | `*string` | OID (urn:oid:...) |
| `uuid` | `*string` | UUID (urn:uuid:...) |
| `markdown` | `*string` | Markdown-formatted text |
| `code` | `*string` or generated enum | e.g., `*AdministrativeGender` for bound value sets |
| `boolean` | `*bool` | true or false |
| `integer` | `*int` | 32-bit signed integer |
| `integer64` | `*int64` | 64-bit signed integer (R5 only) |
| `unsignedInt` | `*uint32` | 32-bit unsigned integer (>= 0) |
| `positiveInt` | `*uint32` | 32-bit unsigned integer (>= 1) |
| `decimal` | `*Decimal` | Custom type, preserves precision |
| `date` | `*string` | ISO 8601 date (YYYY, YYYY-MM, or YYYY-MM-DD) |
| `dateTime` | `*string` | ISO 8601 date/time with optional timezone |
| `instant` | `*string` | ISO 8601 precise timestamp with timezone |
| `time` | `*string` | Time of day (HH:MM:SS) |
| `base64Binary` | `*string` | Base64-encoded binary data |

## Pointer Semantics

All primitive types use pointer semantics (`*T`), where a `nil` pointer means the value is absent from the resource. This is critical for FHIR, where the absence of a field carries meaning distinct from the field being present with a default or zero value.

```go
func ptrTo[T any](v T) *T {
    return &v
}

patient := &r4.Patient{
    ResourceType: "Patient",
    Id:           ptrTo("123"),       // Present: "123"
    Active:       ptrTo(true),        // Present: true
    BirthDate:    ptrTo("1990-01-15"),// Present: "1990-01-15"
    // Gender is nil -- absent from the resource
}
```

When reading fields, always check for `nil` before dereferencing:

```go
if patient.Gender != nil {
    fmt.Println("Gender:", *patient.Gender)
} else {
    fmt.Println("Gender is not specified")
}

if patient.BirthDate != nil {
    fmt.Println("Birth date:", *patient.BirthDate)
}
```

### Why Not Zero Values?

Go's zero values (`""` for strings, `false` for booleans, `0` for integers) cannot distinguish "absent" from "present with zero value." In FHIR:

- A boolean field set to `false` is different from a boolean field that is absent.
- An integer field set to `0` is different from an integer field that is absent.
- A string field set to `""` is different from a string field that is absent.

Pointers solve this cleanly: `nil` means absent, and `ptrTo(false)` means explicitly `false`.

## String-Like Types

FHIR defines many string-like primitive types (`uri`, `url`, `canonical`, `id`, `oid`, `uuid`, `markdown`) that all map to `*string` in Go. The FHIR type information is captured in the struct field comments and JSON tags, but the Go type is the same.

```go
type StructureDefinition struct {
    // ...
    Url         *string `json:"url,omitempty"`       // FHIR type: uri
    Version     *string `json:"version,omitempty"`   // FHIR type: string
    Name        *string `json:"name,omitempty"`      // FHIR type: string
    Description *string `json:"description,omitempty"` // FHIR type: markdown
    // ...
}
```

## Date and Time Types

FHIR has four date/time primitives, all mapped to `*string` in Go. The string values must conform to ISO 8601 format, but the library does not perform date validation at the type level.

```go
patient := &r4.Patient{
    ResourceType: "Patient",
    BirthDate:    ptrTo("1990-01-15"),  // FHIR date: YYYY-MM-DD
}

observation := &r4.Observation{
    ResourceType:     "Observation",
    EffectiveDateTime: ptrTo("2024-03-15T10:30:00Z"), // FHIR dateTime
}
```

| FHIR Type | Format Examples |
|-----------|----------------|
| `date` | `"2024"`, `"2024-03"`, `"2024-03-15"` |
| `dateTime` | `"2024-03-15"`, `"2024-03-15T10:30:00Z"`, `"2024-03-15T10:30:00+01:00"` |
| `instant` | `"2024-03-15T10:30:00.123Z"` (requires full precision with timezone) |
| `time` | `"10:30:00"`, `"10:30:00.123"` |

## Numeric Types

### Integer Types

FHIR defines three integer types with different Go mappings:

```go
// integer -> *int
type Dosage struct {
    Sequence *int `json:"sequence,omitempty"`
}

// unsignedInt -> *uint32
type Attachment struct {
    Size *uint32 `json:"size,omitempty"`
}

// positiveInt -> *uint32
type ContactPoint struct {
    Rank *uint32 `json:"rank,omitempty"`
}
```

### Decimal

The FHIR `decimal` type maps to the custom `*Decimal` type rather than `*float64`, to preserve precision. See the [Decimal Precision](../decimal-precision) page for full details.

```go
type Quantity struct {
    Value *Decimal `json:"value,omitempty"`
    // ...
}

// Create a quantity with precise decimal value
q := r4.Quantity{
    Value: r4.MustDecimal("72.50"),
    Unit:  ptrTo("kg"),
}
```

## Code Types

When a FHIR field is bound to a required value set, the library generates a type-safe enum instead of using `*string`. See the [Code Systems](../code-systems) page for details.

```go
// Gender uses a generated enum type
type Patient struct {
    Gender *AdministrativeGender `json:"gender,omitempty"`
    // ...
}

// Status uses a generated enum type
type Observation struct {
    Status *ObservationStatus `json:"status,omitempty"`
    // ...
}
```

## Extension Elements

Every primitive field in a FHIR resource has a corresponding extension element that can carry the element's `id` and any extensions on the primitive value. In the Go structs, these appear as `*Element` fields with an underscore-prefixed JSON tag.

```go
type Patient struct {
    // The date of birth for the individual
    BirthDate    *string  `json:"birthDate,omitempty"`
    // Extension for BirthDate
    BirthDateExt *Element `json:"_birthDate,omitempty"`

    // Whether this patient's record is in active use
    Active    *bool    `json:"active,omitempty"`
    // Extension for Active
    ActiveExt *Element `json:"_active,omitempty"`
    // ...
}
```

The `Element` struct holds an optional `Id` and a slice of `Extension` values:

```go
type Element struct {
    Id        *string     `json:"id,omitempty"`
    Extension []Extension `json:"extension,omitempty"`
}
```

See the [Extensions](../extensions) page for detailed examples of using primitive extensions.

## Repeating Primitives

When a primitive field can repeat (cardinality `0..*`), it uses a Go slice of the base type rather than a pointer:

```go
type HumanName struct {
    Given     []string  `json:"given,omitempty"`     // Repeating string
    GivenExt  []Element `json:"_given,omitempty"`    // Per-element extensions
    Prefix    []string  `json:"prefix,omitempty"`    // Repeating string
    PrefixExt []Element `json:"_prefix,omitempty"`   // Per-element extensions
    // ...
}
```

An empty or `nil` slice means "no values present." Each element in the extension slice corresponds positionally to the element at the same index in the value slice.

{{< callout type="info" >}}
The library does not perform FHIR validation on primitive values. For example, `*string` fields for `date` types accept any string, and `*uint32` fields for `positiveInt` accept zero even though FHIR requires positive values. Validation should be handled at a higher layer.
{{< /callout >}}

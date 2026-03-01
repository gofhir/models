---
title: "Data Types"
linkTitle: "Data Types"
description: "Overview of the FHIR type system in Go, including primitive types, complex types, code systems, and extensions."
weight: 4
---

The FHIR specification defines a rich type system that underpins all resources. The `gofhir/models` library maps every FHIR data type to an idiomatic Go representation, using pointer semantics for optionality, generated enum types for coded values, and a custom `Decimal` type for precision-sensitive numeric values.

## FHIR Type Categories

FHIR data types fall into several categories, each with a specific Go representation in this library:

### Primitive Types

FHIR primitives are the basic building blocks: strings, booleans, integers, decimals, and dates. In Go, these are represented as pointer types (`*string`, `*bool`, `*int`, `*Decimal`) where `nil` indicates the value is absent. All string-like FHIR types (uri, url, canonical, id, oid, uuid, markdown, date, dateTime, instant, time) map to `*string` in Go.

### Complex Types

Complex types are structured data types composed of multiple elements. Examples include `HumanName`, `Address`, `ContactPoint`, `CodeableConcept`, `Reference`, and `Quantity`. Each is represented as a Go struct with fields for every element defined in the FHIR specification.

### Code Systems

FHIR uses coded values extensively. The library generates type-safe string enums for each FHIR code system, such as `AdministrativeGender`, `ObservationStatus`, and `BundleType`. These prevent invalid code values at compile time.

### The Decimal Type

FHIR requires that decimal values preserve their original precision (e.g., `1.50` must remain `1.50`, not `1.5`). The library provides a custom `Decimal` type that stores the original string representation while supporting numeric operations.

### Extensions

FHIR's extensibility model uses the `Extension` type with a URL and a polymorphic `value[x]` field. The `Element` type carries extensions on primitive values through the `_fieldName` JSON pattern.

## Topics

{{< cards >}}
  {{< card link="primitive-types" title="Primitive Types" subtitle="FHIR-to-Go type mapping, pointer semantics, and nil handling." >}}
  {{< card link="complex-types" title="Complex Types" subtitle="Structured data types: HumanName, CodeableConcept, Reference, and more." >}}
  {{< card link="code-systems" title="Code Systems" subtitle="Generated type-safe enums for FHIR coded values." >}}
  {{< card link="decimal-precision" title="Decimal Precision" subtitle="Custom Decimal type for FHIR-compliant precision preservation." >}}
  {{< card link="extensions" title="Extensions" subtitle="FHIR extensibility with Extension and Element types." >}}
{{< /cards >}}

## Quick Example

Here is a brief example that exercises several data type categories together:

```go
package main

import (
    "fmt"

    "github.com/gofhir/models/r4"
)

func ptrTo[T any](v T) *T {
    return &v
}

func main() {
    obs := &r4.Observation{
        ResourceType: "Observation",
        Id:           ptrTo("bp-reading"),
        Status:       ptrTo(r4.ObservationStatusFinal),  // Code system enum
        Code: &r4.CodeableConcept{                       // Complex type
            Coding: []r4.Coding{
                {
                    System:  ptrTo("http://loinc.org"),
                    Code:    ptrTo("85354-9"),
                    Display: ptrTo("Blood pressure panel"),
                },
            },
        },
        ValueQuantity: &r4.Quantity{                     // Decimal precision
            Value:  r4.MustDecimal("120.0"),
            Unit:   ptrTo("mmHg"),
            System: ptrTo("http://unitsofmeasure.org"),
            Code:   ptrTo("mm[Hg]"),
        },
        Subject: &r4.Reference{                          // Reference type
            Reference: ptrTo("Patient/example"),
        },
    }

    fmt.Printf("Observation %s: status=%s\n", *obs.Id, *obs.Status)
    fmt.Printf("Value: %s %s\n", obs.ValueQuantity.Value.String(), *obs.ValueQuantity.Unit)
}
```

## Type System Design Principles

The Go representations in this library follow several design principles:

1. **Nil means absent.** Pointer types allow distinguishing between "not present" (nil) and "present with default value" (e.g., `*bool` can be nil, true, or false).

2. **Type safety for codes.** Generated string enums catch invalid code values at compile time rather than at runtime.

3. **Precision preservation.** The custom `Decimal` type ensures that numeric precision is never silently lost during serialization round-trips.

4. **Extension support.** Every primitive field has a corresponding `_fieldName` extension field (of type `*Element`) that carries FHIR primitive extensions.

5. **JSON tag consistency.** All struct tags match the FHIR JSON property names exactly, with `omitempty` on optional fields.

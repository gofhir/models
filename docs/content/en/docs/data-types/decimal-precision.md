---
title: "Decimal Precision"
linkTitle: "Decimal Precision"
description: "Custom Decimal type for FHIR-compliant precision preservation in numeric values."
weight: 4
---

The FHIR specification requires that decimal values preserve their original textual precision. For example, a value transmitted as `1.50` must be stored and re-transmitted as `1.50`, not `1.5` or `1.500`. Go's native `float64` type cannot guarantee this, so the `gofhir/models` library provides a custom `Decimal` type that stores the exact string representation while supporting numeric operations.

## The Problem

Go's `float64` type normalizes numeric values, losing trailing zeros and potentially altering precision:

```go
// float64 loses trailing zeros
f := 1.50
fmt.Println(f) // "1.5" -- trailing zero is lost

// JSON marshaling of float64 also loses precision
data, _ := json.Marshal(f)
fmt.Println(string(data)) // "1.5"
```

In FHIR, the precision of a decimal value carries clinical meaning. A lab result reported as `7.0` mg/dL implies single-decimal precision, while `7.00` implies two-decimal precision. Losing this distinction violates the FHIR specification and can cause validation failures with FHIR servers.

## The Decimal Type

The `Decimal` type is defined in `decimal.go`:

```go
type Decimal struct {
    value string
}
```

It stores the numeric value as its original string representation. All construction methods validate that the string is a valid decimal number (not NaN or Infinity), while preserving the exact input text.

## Construction Methods

### NewDecimalFromString

Creates a `Decimal` from a string representation, preserving the exact text. Returns an error if the string is not a valid decimal number.

```go
d, err := r4.NewDecimalFromString("1.50")
if err != nil {
    log.Fatal(err)
}
fmt.Println(d.String()) // "1.50" -- precision preserved
```

This is the preferred constructor when you have a string value from an external source (JSON, database, user input) and want to preserve its exact representation.

### MustDecimal

Creates a `Decimal` from a string, panicking on invalid input. Use only for compile-time constant values that are known to be valid.

```go
d := r4.MustDecimal("1.50")
fmt.Println(d.String()) // "1.50"

// Panics on invalid input
// d := r4.MustDecimal("not-a-number") // panic!
```

### NewDecimalFromFloat64

Creates a `Decimal` from a `float64` value. Note that precision may be lost during the float-to-string conversion.

```go
d := r4.NewDecimalFromFloat64(1.5)
fmt.Println(d.String()) // "1.5" -- trailing zero lost (float64 normalization)

d2 := r4.NewDecimalFromFloat64(72.5)
fmt.Println(d2.String()) // "72.5"
```

{{< callout type="info" >}}
Use `NewDecimalFromString` instead of `NewDecimalFromFloat64` when precision preservation is important. The float64 constructor is a convenience for cases where precision loss is acceptable.
{{< /callout >}}

### NewDecimalFromInt

Creates a `Decimal` from an `int` value:

```go
d := r4.NewDecimalFromInt(100)
fmt.Println(d.String()) // "100"
```

### NewDecimalFromInt64

Creates a `Decimal` from an `int64` value:

```go
d := r4.NewDecimalFromInt64(9999999999)
fmt.Println(d.String()) // "9999999999"
```

## Access Methods

### String

Returns the exact textual representation of the decimal:

```go
d := r4.MustDecimal("3.14159")
fmt.Println(d.String()) // "3.14159"
```

### Float64

Converts the decimal to a `float64` for numeric operations. Precision may be lost in the conversion:

```go
d := r4.MustDecimal("72.50")
f := d.Float64()
fmt.Println(f) // 72.5
```

### IsZero

Returns `true` if the decimal value is zero or empty:

```go
d1 := r4.MustDecimal("0")
d2 := r4.MustDecimal("0.00")
d3 := r4.MustDecimal("1.5")

fmt.Println(d1.IsZero()) // true
fmt.Println(d2.IsZero()) // true
fmt.Println(d3.IsZero()) // false
```

### Equal

Compares two `Decimal` values numerically (not textually). This means `"1.0"` and `"1.00"` are considered equal:

```go
d1 := r4.MustDecimal("1.0")
d2 := r4.MustDecimal("1.00")
d3 := r4.MustDecimal("2.0")

fmt.Println(d1.Equal(*d2)) // true  (same numeric value)
fmt.Println(d1.Equal(*d3)) // false
```

Note that `Equal` compares via `float64` conversion, so it checks numeric equality rather than textual identity.

## JSON Marshaling

The `Decimal` type implements `json.Marshaler` and `json.Unmarshaler` to produce spec-compliant JSON output.

### Marshaling

The `Decimal` marshals as a bare JSON number (not a quoted string), preserving the original textual precision:

```go
d := r4.MustDecimal("1.50")
data, _ := json.Marshal(d)
fmt.Println(string(data)) // 1.50 (bare number, not "1.50")
```

This is critical for FHIR compliance. The output is `1.50`, not `1.5` (which would lose precision) and not `"1.50"` (which would be a string instead of a number).

### Unmarshaling

The `Decimal` accepts a bare JSON number and stores its exact byte representation:

```go
var d r4.Decimal
json.Unmarshal([]byte("1.50"), &d)
fmt.Println(d.String()) // "1.50" -- precision preserved
```

It also handles quoted numbers for compatibility with some FHIR servers that encode decimals as strings:

```go
var d r4.Decimal
json.Unmarshal([]byte(`"1.50"`), &d)
fmt.Println(d.String()) // "1.50"
```

## Usage in Resources

The `Decimal` type is used wherever FHIR defines a `decimal` element. The most common occurrence is in `Quantity` and its specializations (`Age`, `Distance`, `Duration`, `Count`):

```go
observation := &r4.Observation{
    ResourceType: "Observation",
    Id:           ptrTo("weight-1"),
    Status:       ptrTo(r4.ObservationStatusFinal),
    Code: &r4.CodeableConcept{
        Coding: []r4.Coding{
            {
                System:  ptrTo("http://loinc.org"),
                Code:    ptrTo("29463-7"),
                Display: ptrTo("Body weight"),
            },
        },
    },
    ValueQuantity: &r4.Quantity{
        Value:  r4.MustDecimal("72.50"),
        Unit:   ptrTo("kg"),
        System: ptrTo("http://unitsofmeasure.org"),
        Code:   ptrTo("kg"),
    },
}

// Marshal preserves the "72.50" precision
data, _ := r4.Marshal(observation)
fmt.Println(string(data))
// ... "value":72.50 ...
```

## Round-Trip Preservation

The `Decimal` type guarantees round-trip precision preservation through JSON:

```go
// Original value with trailing zero
original := r4.MustDecimal("98.60")

// Marshal to JSON
data, _ := json.Marshal(original)
fmt.Println(string(data)) // 98.60

// Unmarshal back
var restored r4.Decimal
json.Unmarshal(data, &restored)
fmt.Println(restored.String()) // "98.60" -- identical to original
```

## Empty Decimal Behavior

An empty `Decimal` (zero-value struct) marshals as `0`:

```go
var d r4.Decimal
data, _ := json.Marshal(d)
fmt.Println(string(data)) // 0
```

The `IsZero` method returns `true` for both empty decimals and numeric zeros:

```go
var empty r4.Decimal
fmt.Println(empty.IsZero()) // true

zero := r4.MustDecimal("0")
fmt.Println(zero.IsZero()) // true
```

{{< callout type="info" >}}
The `Decimal` type stores values as strings internally, which means arithmetic operations are not directly supported. For calculations, convert to `float64` using the `Float64()` method, perform the arithmetic, then create a new `Decimal` from the result. Be aware that this may affect precision.
{{< /callout >}}

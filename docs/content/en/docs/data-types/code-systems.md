---
title: "Code Systems"
linkTitle: "Code Systems"
description: "Generated type-safe enum types for FHIR coded values, including AdministrativeGender, ObservationStatus, and more."
weight: 3
---

FHIR uses coded values extensively to represent categories, statuses, types, and other enumerated concepts. The `gofhir/models` library generates type-safe string-based enum types for each FHIR code system, providing compile-time safety and IDE autocompletion for coded fields.

## Generated Enum Pattern

Each FHIR code system is represented as a Go type alias of `string`, with constants for each valid code value. All enum types and constants are generated in `codesystems.go` from the FHIR ValueSet definitions.

The naming convention follows this pattern:

- **Type name:** The FHIR code system name in PascalCase (e.g., `AdministrativeGender`)
- **Constant name:** The type name followed by the code value in PascalCase (e.g., `AdministrativeGenderMale`)
- **Constant value:** The FHIR code string (e.g., `"male"`)

## AdministrativeGender

One of the most commonly used code systems:

```go
type AdministrativeGender string

const (
    AdministrativeGenderMale    AdministrativeGender = "male"
    AdministrativeGenderFemale  AdministrativeGender = "female"
    AdministrativeGenderOther   AdministrativeGender = "other"
    AdministrativeGenderUnknown AdministrativeGender = "unknown"
)
```

Used in the `Patient` struct:

```go
type Patient struct {
    // ...
    Gender    *AdministrativeGender `json:"gender,omitempty"`
    GenderExt *Element             `json:"_gender,omitempty"`
    // ...
}
```

Example usage:

```go
func ptrTo[T any](v T) *T {
    return &v
}

patient := &r4.Patient{
    ResourceType: "Patient",
    Id:           ptrTo("example"),
    Gender:       ptrTo(r4.AdministrativeGenderFemale),
}

// Read the value
if patient.Gender != nil {
    switch *patient.Gender {
    case r4.AdministrativeGenderMale:
        fmt.Println("Male")
    case r4.AdministrativeGenderFemale:
        fmt.Println("Female")
    case r4.AdministrativeGenderOther:
        fmt.Println("Other")
    case r4.AdministrativeGenderUnknown:
        fmt.Println("Unknown")
    }
}
```

## ObservationStatus

A complete example showing all values for the `ObservationStatus` code system:

```go
type ObservationStatus string

const (
    ObservationStatusRegistered    ObservationStatus = "registered"
    ObservationStatusPreliminary   ObservationStatus = "preliminary"
    ObservationStatusFinal         ObservationStatus = "final"
    ObservationStatusAmended       ObservationStatus = "amended"
    ObservationStatusCorrected     ObservationStatus = "corrected"
    ObservationStatusCancelled     ObservationStatus = "cancelled"
    ObservationStatusEnteredInError ObservationStatus = "entered-in-error"
    ObservationStatusUnknown       ObservationStatus = "unknown"
)
```

Using it in an Observation resource:

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
        Id:           ptrTo("vitals-1"),
        Status:       ptrTo(r4.ObservationStatusFinal),
        Code: &r4.CodeableConcept{
            Coding: []r4.Coding{
                {
                    System:  ptrTo("http://loinc.org"),
                    Code:    ptrTo("8867-4"),
                    Display: ptrTo("Heart rate"),
                },
            },
        },
        ValueQuantity: &r4.Quantity{
            Value:  r4.MustDecimal("72"),
            Unit:   ptrTo("beats/minute"),
            System: ptrTo("http://unitsofmeasure.org"),
            Code:   ptrTo("/min"),
        },
    }

    fmt.Printf("Observation %s: %s\n", *obs.Id, *obs.Status)
    // Output: Observation vitals-1: final
}
```

## Other Common Code Systems

The library generates enum types for all FHIR code systems. Here are some frequently used ones:

### BundleType

```go
type BundleType string

const (
    BundleTypeDocument    BundleType = "document"
    BundleTypeMessage     BundleType = "message"
    BundleTypeTransaction BundleType = "transaction"
    BundleTypeBatch       BundleType = "batch"
    BundleTypeSearchset   BundleType = "searchset"
    // ...
)
```

### AccountStatus

```go
type AccountStatus string

const (
    AccountStatusActive         AccountStatus = "active"
    AccountStatusInactive       AccountStatus = "inactive"
    AccountStatusEnteredInError AccountStatus = "entered-in-error"
    AccountStatusOnHold         AccountStatus = "on-hold"
    AccountStatusUnknown        AccountStatus = "unknown"
)
```

### NameUse

```go
type NameUse string

const (
    NameUseUsual     NameUse = "usual"
    NameUseOfficial  NameUse = "official"
    NameUseTemp      NameUse = "temp"
    NameUseNickname  NameUse = "nickname"
    NameUseAnonymous NameUse = "anonymous"
    NameUseOld       NameUse = "old"
    NameUseMaiden    NameUse = "maiden"
)
```

### ContactPointSystem

```go
type ContactPointSystem string

const (
    ContactPointSystemPhone ContactPointSystem = "phone"
    ContactPointSystemFax   ContactPointSystem = "fax"
    ContactPointSystemEmail ContactPointSystem = "email"
    ContactPointSystemPager ContactPointSystem = "pager"
    ContactPointSystemUrl   ContactPointSystem = "url"
    ContactPointSystemSms   ContactPointSystem = "sms"
    ContactPointSystemOther ContactPointSystem = "other"
)
```

## Type Safety Benefits

Using generated enum types instead of raw strings provides several advantages:

### Compile-Time Validation

The compiler catches typos and invalid values:

```go
// This compiles -- valid code
patient.Gender = ptrTo(r4.AdministrativeGenderMale)

// This would cause a compile error if you tried to assign an invalid string
// patient.Gender = ptrTo("mael")  // type mismatch
```

### IDE Autocompletion

When you type `r4.AdministrativeGender`, your IDE will suggest all valid values, making it easy to discover available codes without consulting the FHIR specification.

### Self-Documenting Code

Enum constants include a comment with the display name from the FHIR ValueSet:

```go
// AdministrativeGenderMale - Male
AdministrativeGenderMale AdministrativeGender = "male"
// AdministrativeGenderFemale - Female
AdministrativeGenderFemale AdministrativeGender = "female"
```

## Working with String Values

Since the enum types are based on `string`, you can convert between them and raw strings when needed:

```go
// From enum to string
gender := r4.AdministrativeGenderMale
s := string(gender) // "male"

// From string to enum
input := "female"
gender = r4.AdministrativeGender(input)
```

This is useful when reading code values from external sources like configuration files or databases.

## JSON Serialization

Enum types serialize to and from their string values in JSON, exactly as FHIR specifies:

```go
patient := &r4.Patient{
    ResourceType: "Patient",
    Gender:       ptrTo(r4.AdministrativeGenderMale),
}

data, _ := json.Marshal(patient)
// {"resourceType":"Patient","gender":"male"}

var decoded r4.Patient
json.Unmarshal(data, &decoded)
fmt.Println(*decoded.Gender) // "male"
fmt.Println(*decoded.Gender == r4.AdministrativeGenderMale) // true
```

## XML Serialization

In XML, code values are encoded using the standard FHIR primitive encoding with a `value` attribute. The library handles this through the generic `xmlEncodePrimitiveCode` helper:

```xml
<Patient xmlns="http://hl7.org/fhir">
  <gender value="male"/>
</Patient>
```

{{< callout type="info" >}}
The library generates enum types for code systems with required bindings in the FHIR specification. For code systems with extensible or example bindings, the field type is `*string` to allow any code value, since those bindings permit additional codes beyond those defined in the value set.
{{< /callout >}}

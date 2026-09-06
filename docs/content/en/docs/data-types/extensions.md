---
title: "Extensions"
linkTitle: "Extensions"
description: "FHIR extensibility model with Extension and Element types, including the _fieldName JSON pattern."
weight: 5
---

FHIR's extensibility framework allows resources to carry additional data beyond what the base specification defines. The `gofhir/models` library supports extensions through the `Extension` struct (for adding data to resources and complex types) and the `Element` struct (for extending primitive values). Extensions on primitives are preserved at every level — resource, datatype and backbone element alike. Backbone elements used to be the exception, which accounted for every extension loss measured across the corpora; that was fixed and the corpus now round-trips 8757/8757 in JSON and 3653/3653 in XML.

## The Extension Struct

The `Extension` type represents a FHIR extension with a URL that identifies the extension definition and a polymorphic `value[x]` field that can hold any FHIR data type:

```go
type Extension struct {
    Id        *string     `json:"id,omitempty"`
    Extension []Extension `json:"extension,omitempty"` // Nested extensions
    Url       string      `json:"url"`                 // Required: identifies the extension

    // Primitive value types
    ValueBase64Binary *string  `json:"valueBase64Binary,omitempty"`
    ValueBoolean      *bool    `json:"valueBoolean,omitempty"`
    ValueCanonical    *string  `json:"valueCanonical,omitempty"`
    ValueCode         *string  `json:"valueCode,omitempty"`
    ValueDate         *string  `json:"valueDate,omitempty"`
    ValueDateTime     *string  `json:"valueDateTime,omitempty"`
    ValueDecimal      *Decimal `json:"valueDecimal,omitempty"`
    ValueId           *string  `json:"valueId,omitempty"`
    ValueInstant      *string  `json:"valueInstant,omitempty"`
    ValueInteger      *int     `json:"valueInteger,omitempty"`
    ValueMarkdown     *string  `json:"valueMarkdown,omitempty"`
    ValueOid          *string  `json:"valueOid,omitempty"`
    ValuePositiveInt  *uint32  `json:"valuePositiveInt,omitempty"`
    ValueString       *string  `json:"valueString,omitempty"`
    ValueTime         *string  `json:"valueTime,omitempty"`
    ValueUnsignedInt  *uint32  `json:"valueUnsignedInt,omitempty"`
    ValueUri          *string  `json:"valueUri,omitempty"`
    ValueUrl          *string  `json:"valueUrl,omitempty"`
    ValueUuid         *string  `json:"valueUuid,omitempty"`

    // Complex value types
    ValueAddress          *Address          `json:"valueAddress,omitempty"`
    ValueAge              *Age              `json:"valueAge,omitempty"`
    ValueAnnotation       *Annotation       `json:"valueAnnotation,omitempty"`
    ValueAttachment       *Attachment       `json:"valueAttachment,omitempty"`
    ValueCodeableConcept  *CodeableConcept  `json:"valueCodeableConcept,omitempty"`
    ValueCoding           *Coding           `json:"valueCoding,omitempty"`
    ValueContactPoint     *ContactPoint     `json:"valueContactPoint,omitempty"`
    ValueCount            *Count            `json:"valueCount,omitempty"`
    ValueDistance         *Distance         `json:"valueDistance,omitempty"`
    ValueDuration         *Duration         `json:"valueDuration,omitempty"`
    ValueHumanName        *HumanName        `json:"valueHumanName,omitempty"`
    ValueIdentifier       *Identifier       `json:"valueIdentifier,omitempty"`
    ValueMoney            *Money            `json:"valueMoney,omitempty"`
    ValuePeriod           *Period           `json:"valuePeriod,omitempty"`
    ValueQuantity         *Quantity         `json:"valueQuantity,omitempty"`
    ValueRange            *Range            `json:"valueRange,omitempty"`
    ValueRatio            *Ratio            `json:"valueRatio,omitempty"`
    ValueReference        *Reference        `json:"valueReference,omitempty"`
    // ... and more complex types
}
```

Only one `value[x]` field should be set at a time. The FHIR specification enforces that an extension carries at most one value.

## Adding Extensions to Resources

Extensions can be added to any resource or complex type through the `Extension` slice field:

```go
func ptrTo[T any](v T) *T {
    return &v
}

patient := &r4.Patient{
    Id:           ptrTo("with-extensions"),
    Gender:       ptrTo(r4.AdministrativeGenderMale),
    Extension: []r4.Extension{
        {
            Url:         "http://hl7.org/fhir/StructureDefinition/patient-birthPlace",
            ValueAddress: &r4.Address{
                City:    ptrTo("Springfield"),
                State:   ptrTo("IL"),
                Country: ptrTo("US"),
            },
        },
        {
            Url:          "http://example.org/fhir/StructureDefinition/favorite-color",
            ValueString:  ptrTo("blue"),
        },
    },
}
```

This produces the following JSON:

```json
{
  "resourceType": "Patient",
  "id": "with-extensions",
  "extension": [
    {
      "url": "http://hl7.org/fhir/StructureDefinition/patient-birthPlace",
      "valueAddress": {
        "city": "Springfield",
        "state": "IL",
        "country": "US"
      }
    },
    {
      "url": "http://example.org/fhir/StructureDefinition/favorite-color",
      "valueString": "blue"
    }
  ],
  "gender": "male"
}
```

## Modifier Extensions

Modifier extensions change the meaning of the element they are attached to. They are carried in the `ModifierExtension` field, which is available on all `DomainResource` types and `BackboneElement` types:

```go
patient := &r4.Patient{
    Id:           ptrTo("with-modifier"),
    ModifierExtension: []r4.Extension{
        {
            Url:          "http://example.org/fhir/StructureDefinition/confidential",
            ValueBoolean: ptrTo(true),
        },
    },
}
```

{{< callout type="info" >}}
Modifier extensions must be understood by any system processing the resource. If a system encounters a modifier extension it does not recognize, it should reject the resource or handle it with appropriate caution. The FHIR specification requires that modifier extensions are prominently represented in serialized output.
{{< /callout >}}

## The Element Type and Primitive Extensions

FHIR allows extensions on primitive values (strings, booleans, integers, etc.) through a special JSON pattern. In the Go structs, every primitive field has a corresponding `*Element` field with an underscore-prefixed JSON tag.

### The _fieldName Pattern

In FHIR JSON, primitive extensions are represented using a property named `_fieldName` alongside the `fieldName` property:

```json
{
  "resourceType": "Patient",
  "birthDate": "1990-01-15",
  "_birthDate": {
    "id": "birth-date-element",
    "extension": [
      {
        "url": "http://hl7.org/fhir/StructureDefinition/patient-birthTime",
        "valueDateTime": "1990-01-15T08:30:00Z"
      }
    ]
  }
}
```

In the Go struct, this maps to:

```go
patient := &r4.Patient{
    BirthDate:    ptrTo("1990-01-15"),
    BirthDateExt: &r4.Element{
        Id: ptrTo("birth-date-element"),
        Extension: []r4.Extension{
            {
                Url:           "http://hl7.org/fhir/StructureDefinition/patient-birthTime",
                ValueDateTime: ptrTo("1990-01-15T08:30:00Z"),
            },
        },
    },
}
```

### Element Struct

The `Element` struct is minimal, holding just an optional ID and a slice of extensions:

```go
type Element struct {
    Id        *string     `json:"id,omitempty"`
    Extension []Extension `json:"extension,omitempty"`
}
```

### Common Use Cases for Primitive Extensions

**Data absent reason:** When a required primitive value is absent, FHIR allows providing a reason via the `data-absent-reason` extension:

```go
// Patient gender is unknown, but we provide a reason
patient := &r4.Patient{
    // Gender is nil (absent)
    GenderExt: &r4.Element{
        Extension: []r4.Extension{
            {
                Url:       "http://hl7.org/fhir/StructureDefinition/data-absent-reason",
                ValueCode: ptrTo("asked-declined"),
            },
        },
    },
}
```

This produces:

```json
{
  "resourceType": "Patient",
  "_gender": {
    "extension": [
      {
        "url": "http://hl7.org/fhir/StructureDefinition/data-absent-reason",
        "valueCode": "asked-declined"
      }
    ]
  }
}
```

**Element ID:** You can assign an ID to a primitive element for referencing from other parts of the resource:

```go
patient := &r4.Patient{
    BirthDate:    ptrTo("1990-01-15"),
    BirthDateExt: &r4.Element{
        Id: ptrTo("dob"),
    },
}
```

## Nested (Complex) Extensions

Extensions can themselves contain nested extensions instead of a simple value. This is used for complex extension definitions:

```go
patient := &r4.Patient{
    Extension: []r4.Extension{
        {
            Url: "http://hl7.org/fhir/StructureDefinition/patient-nationality",
            Extension: []r4.Extension{
                {
                    Url: "code",
                    ValueCodeableConcept: &r4.CodeableConcept{
                        Coding: []r4.Coding{
                            {
                                System:  ptrTo("urn:iso:std:iso:3166"),
                                Code:    ptrTo("US"),
                                Display: ptrTo("United States of America"),
                            },
                        },
                    },
                },
                {
                    Url:        "period",
                    ValuePeriod: &r4.Period{
                        Start: ptrTo("1990-01-15"),
                    },
                },
            },
        },
    },
}
```

## Extensions on Repeating Primitives

For repeating primitive fields (like `HumanName.given`), the extension array aligns positionally with the value array. Each index in the extension array corresponds to the value at the same index:

```go
name := r4.HumanName{
    Given: r4.PtrSlice("John", "Michael"),
    GivenExt: []r4.Element{
        {}, // No extension for "John" -- empty element
        {   // Extension for "Michael"
            Extension: []r4.Extension{
                {
                    Url:         "http://example.org/fhir/StructureDefinition/name-source",
                    ValueString: ptrTo("middle-name"),
                },
            },
        },
    },
}
```

In JSON, this serializes as:

```json
{
  "given": ["John", "Michael"],
  "_given": [
    {},
    {
      "extension": [
        {
          "url": "http://example.org/fhir/StructureDefinition/name-source",
          "valueString": "middle-name"
        }
      ]
    }
  ]
}
```

## Reading Extensions

Extensions are found by URL — the order of the slice carries no meaning — and every
type that can hold extensions has the lookup built in:

```go
race := patient.GetExtensionByURL("http://hl7.org/fhir/us/core/StructureDefinition/us-core-race")
if race != nil && race.ValueString != nil {
    fmt.Println("Race:", *race.ValueString)
}
```

Three methods, because one answer is not always the right one:

| Method | Returns |
|---|---|
| `GetExtensionByURL(url)` | the first match as `*Extension`, or `nil` |
| `GetExtensionsByURL(url)` | **every** match as `[]*Extension` |
| `HasExtensionByURL(url)` | whether any is present |

`GetExtensionsByURL` matters because a URL repeats where the extension's cardinality
allows it, so taking only the first would drop data. `HasExtensionByURL` matters
because some extensions are flags with no value at all — their presence is the whole
meaning.

Both lookups return **pointers into the resource**, so writing through a result edits
the resource rather than a copy:

```go
if ext := patient.GetExtensionByURL(url); ext != nil {
    ext.ValueString = r4.Ptr("updated")   // patient is updated
}
```

### Every level, including inside another extension

The methods are generated on every type that has an `extension` field — resources,
datatypes, backbone elements, and `Extension` itself. That last one is how a complex
extension is walked, since it holds sub-extensions rather than a value:

```go
ethnicity := patient.GetExtensionByURL(ethnicityURL)
ombCode := ethnicity.GetExtensionByURL("ombCategory")

// datatypes and backbones too
given := patient.Name[0].GetExtensionByURL(url)
phone := patient.Contact[0].GetExtensionByURL(url)
```

Modifier extensions have their own lookup, `GetModifierExtensionByURL`, deliberately
kept separate: a modifier extension changes the meaning of the element it is on, so a
reader that does not recognize one must not process the element at all.

### On a bare slice

The same three exist as functions, for a `[]Extension` that did not come from a
generated type:

```go
r4.ExtensionByURL(exts, url)
r4.ExtensionsByURL(exts, url)
r4.HasExtensionByURL(exts, url)
```

{{< callout type="info" >}}
`Extension.Url` is a `*string`, so an extension with no URL at all is representable —
which is why a hand-written search loop needs a nil check before dereferencing it. The
built-in lookups do that; a loop that forgets panics on the first extension without a
URL.
{{< /callout >}}


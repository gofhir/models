---
title: "Extensiones"
linkTitle: "Extensiones"
description: "Modelo de extensibilidad FHIR con tipos Extension y Element, incluyendo el patrón JSON _fieldName."
weight: 5
---

El framework de extensibilidad de FHIR permite que los recursos lleven datos adicionales más allá de lo que define la especificación base. La biblioteca `gofhir/models` proporciona soporte completo para extensiones a través del struct `Extension` (para agregar datos a recursos y tipos complejos) y el struct `Element` (para extender valores primitivos).

## El Struct Extension

El tipo `Extension` representa una extensión FHIR con una URL que identifica la definición de la extensión y un campo polimórfico `value[x]` que puede contener cualquier tipo de dato FHIR:

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

Solo un campo `value[x]` debe estar establecido a la vez. La especificación FHIR establece que una extensión lleva como máximo un valor.

## Agregar Extensiones a Recursos

Las extensiones pueden agregarse a cualquier recurso o tipo complejo a través del campo slice `Extension`:

```go
func ptrTo[T any](v T) *T {
    return &v
}

patient := &r4.Patient{
    ResourceType: "Patient",
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

Esto produce el siguiente JSON:

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

## Extensiones Modificadoras

Las extensiones modificadoras cambian el significado del elemento al que están adjuntas. Se llevan en el campo `ModifierExtension`, que está disponible en todos los tipos `DomainResource` y `BackboneElement`:

```go
patient := &r4.Patient{
    ResourceType: "Patient",
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
Las extensiones modificadoras deben ser comprendidas por cualquier sistema que procese el recurso. Si un sistema encuentra una extensión modificadora que no reconoce, debe rechazar el recurso o manejarlo con la precaución apropiada. La especificación FHIR requiere que las extensiones modificadoras se representen de forma prominente en la salida serializada.
{{< /callout >}}

## El Tipo Element y Extensiones de Primitivos

FHIR permite extensiones en valores primitivos (cadenas, booleanos, enteros, etc.) a través de un patrón JSON especial. En los structs de Go, cada campo primitivo tiene un campo `*Element` correspondiente con una etiqueta JSON prefijada con guion bajo.

### El Patrón _fieldName

En JSON FHIR, las extensiones de primitivos se representan usando una propiedad llamada `_fieldName` junto a la propiedad `fieldName`:

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

En el struct de Go, esto se mapea a:

```go
patient := &r4.Patient{
    ResourceType: "Patient",
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

### Struct Element

El struct `Element` es mínimo, conteniendo solo un ID opcional y un slice de extensiones:

```go
type Element struct {
    Id        *string     `json:"id,omitempty"`
    Extension []Extension `json:"extension,omitempty"`
}
```

### Casos de Uso Comunes para Extensiones de Primitivos

**Razón de ausencia de datos:** Cuando un valor primitivo requerido está ausente, FHIR permite proporcionar una razón a través de la extensión `data-absent-reason`:

```go
// Patient gender is unknown, but we provide a reason
patient := &r4.Patient{
    ResourceType: "Patient",
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

Esto produce:

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

**ID de elemento:** Puedes asignar un ID a un elemento primitivo para referenciarlo desde otras partes del recurso:

```go
patient := &r4.Patient{
    ResourceType: "Patient",
    BirthDate:    ptrTo("1990-01-15"),
    BirthDateExt: &r4.Element{
        Id: ptrTo("dob"),
    },
}
```

## Extensiones Anidadas (Complejas)

Las extensiones pueden contener a su vez extensiones anidadas en lugar de un valor simple. Esto se usa para definiciones de extensiones complejas:

```go
patient := &r4.Patient{
    ResourceType: "Patient",
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

## Extensiones en Primitivos Repetidos

Para campos primitivos repetidos (como `HumanName.given`), el arreglo de extensiones se alinea posicionalmente con el arreglo de valores. Cada índice en el arreglo de extensiones corresponde al valor en el mismo índice:

```go
name := r4.HumanName{
    Given: []string{"John", "Michael"},
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

En JSON, esto se serializa como:

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

## Lectura de Extensiones

Al procesar datos FHIR entrantes, verifica las extensiones examinando el slice `Extension`:

```go
func findExtension(extensions []r4.Extension, url string) *r4.Extension {
    for i := range extensions {
        if extensions[i].Url == url {
            return &extensions[i]
        }
    }
    return nil
}

// Usage
ext := findExtension(patient.Extension, "http://example.org/fhir/StructureDefinition/favorite-color")
if ext != nil && ext.ValueString != nil {
    fmt.Println("Favorite color:", *ext.ValueString)
}
```

{{< callout type="info" >}}
El campo `Extension.Url` es un `string` requerido (no un puntero), porque toda extensión debe tener una URL que identifique su definición. Este es el único campo que no es puntero en el struct de extensión además de `Id`.
{{< /callout >}}

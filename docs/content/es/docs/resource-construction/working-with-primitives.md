---
title: "Trabajo con Primitivos"
linkTitle: "Trabajo con Primitivos"
description: "Comprende cómo los tipos primitivos FHIR se mapean a tipos Go, incluyendo punteros, el tipo Decimal y elementos de extensión."
weight: 4
---

FHIR define un conjunto de tipos de datos primitivos (string, boolean, integer, decimal, date, dateTime y otros) que aparecen a lo largo de cada recurso. Esta página explica cómo estos primitivos se representan en Go y cómo trabajar con ellos de manera efectiva.

## Mapeo de Tipos Primitivos

Los tipos primitivos FHIR se mapean a tipos Go de la siguiente manera:

| Tipo FHIR | Tipo Go | Valor de Ejemplo |
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
| `code` | Tipo personalizado (ej., `*AdministrativeGender`) | `AdministrativeGenderMale` |
| `base64Binary` | `*string` | Cadena codificada en Base64 |
| `positiveInt` | `*int` | `1` |
| `unsignedInt` | `*int` | `0` |

## Por Qué Punteros

En FHIR, la mayoría de los campos son opcionales. Un campo ausente tiene un significado diferente a un campo establecido en su valor cero. Por ejemplo, `active: false` es diferente de que el campo `active` esté completamente ausente. Los punteros Go distinguen estos casos:

- `nil` -- el campo está ausente (se omite de la salida JSON)
- `&value` -- el campo está presente con el valor dado

```go
// Active is absent (nil)
patient := r4.Patient{Id: ptrTo("1")}

// Active is explicitly false
patient := r4.Patient{Id: ptrTo("1"), Active: ptrTo(false)}
```

Tanto el patrón builder como las opciones funcionales manejan el envolvimiento de punteros automáticamente, por lo que solo tratas con valores sin procesar:

```go
// Builder -- no pointer needed
patient := r4.NewPatientBuilder().SetActive(false).Build()

// Functional options -- no pointer needed
patient := r4.NewPatient(r4.WithPatientActive(false))
```

## El Tipo Decimal

FHIR requiere que los valores decimales preserven su representación textual exacta. Por ejemplo, `1.50` debe permanecer como `1.50` en la salida JSON, no `1.5`. El tipo estándar `float64` de Go pierde los ceros finales, por lo que la biblioteca proporciona un tipo `Decimal` personalizado.

### Crear Valores Decimal

Hay varias formas de crear un `Decimal`:

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

### Usar Decimal en Recursos

Los valores Decimal aparecen en Quantity, Money y otros tipos de datos:

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

Para valores donde la precisión es crítica (como resultados de laboratorio o dosificaciones de medicamentos), siempre usa `NewDecimalFromString`:

```go
// Preserves "1.50" exactly in JSON output
quantity := r4.Quantity{
    Value: r4.MustDecimal("1.50"),
    Unit:  ptrTo("mg"),
}
```

### Métodos de Decimal

El tipo `Decimal` proporciona estos métodos:

| Método | Retorna | Descripción |
|--------|---------|-------------|
| `String()` | `string` | Representación textual exacta |
| `Float64()` | `float64` | Valor numérico (puede perder precisión) |
| `IsZero()` | `bool` | Verdadero si es cero o está vacío |
| `Equal(other)` | `bool` | Igualdad numérica (ignora ceros finales) |
| `MarshalJSON()` | `[]byte, error` | Emite número JSON sin comillas preservando la precisión |
| `UnmarshalJSON(data)` | `error` | Parsea número JSON sin comillas preservando la precisión |

### Preservación de Precisión en JSON

El tipo `Decimal` se serializa como un número JSON sin comillas (no como una cadena entrecomillada), preservando los dígitos exactos:

```go
d := r4.MustDecimal("1.50")
data, _ := json.Marshal(d)
fmt.Println(string(data)) // 1.50 (not 1.5 or "1.50")
```

Al deserializar, la representación exacta de la entrada JSON se preserva:

```go
var d r4.Decimal
json.Unmarshal([]byte("1.50"), &d)
fmt.Println(d.String()) // "1.50"
```

## Elementos de Extensión

En FHIR, cada elemento primitivo puede llevar extensiones. La representación JSON usa un campo paralelo con prefijo `_`. Por ejemplo, un campo `birthDate` tiene un campo `_birthDate` correspondiente para extensiones.

En los structs Go, estos aparecen como campos `<Field>Ext` de tipo `*Element`:

```go
type Patient struct {
    // ...
    BirthDate    *string  `json:"birthDate,omitempty"`
    BirthDateExt *Element `json:"_birthDate,omitempty"`
    // ...
}
```

El tipo `Element` contiene un `Id` opcional y un slice de valores `Extension`:

```go
type Element struct {
    Id        *string     `json:"id,omitempty"`
    Extension []Extension `json:"extension,omitempty"`
}
```

### Establecer Extensiones en Primitivos

Para agregar una extensión a un campo primitivo, establece tanto el valor como su elemento de extensión:

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

Esto produce JSON con el campo `_birthDate`:

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

### Primitivos Solo con Extensión

FHIR permite que un primitivo tenga solo una extensión sin valor. En Go, establece el campo de valor a `nil` y rellena solo el campo de extensión:

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

## Tipos de Sistemas de Códigos

Los elementos code de FHIR se representan como constantes de cadena tipada en lugar de cadenas simples. Esto proporciona validación en tiempo de compilación de los valores de código:

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

Usa las constantes predefinidas para seguridad de tipos:

```go
patient := r4.NewPatient(
    r4.WithPatientGender(r4.AdministrativeGenderMale),
)
```

El compilador rechazará valores inválidos, capturando errores que solo se manifestarían en tiempo de ejecución con cadenas simples.

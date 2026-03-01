---
title: "Tipos Primitivos"
linkTitle: "Tipos Primitivos"
description: "Mapeo completo de tipos primitivos FHIR a Go, semántica de punteros y elementos de extensión."
weight: 1
---

FHIR define un conjunto de tipos de datos primitivos que sirven como bloques de construcción para todos los campos de recursos. La biblioteca `gofhir/models` mapea cada tipo primitivo FHIR a un tipo de Go apropiado, usando semántica de punteros para representar la opcionalidad.

## Mapeo de Tipos FHIR a Go

La siguiente tabla muestra el mapeo completo de tipos primitivos FHIR a sus representaciones en Go:

| Tipo FHIR | Tipo Go | Notas |
|-----------|---------|-------|
| `string` | `*string` | Texto de propósito general |
| `uri` | `*string` | Referencia URI/IRI |
| `url` | `*string` | URL absoluta |
| `canonical` | `*string` | URL canónica a una definición de recurso FHIR |
| `id` | `*string` | ID lógico del recurso (1-64 caracteres, [A-Za-z0-9\-.]) |
| `oid` | `*string` | OID (urn:oid:...) |
| `uuid` | `*string` | UUID (urn:uuid:...) |
| `markdown` | `*string` | Texto con formato Markdown |
| `code` | `*string` o enum generado | p. ej., `*AdministrativeGender` para conjuntos de valores vinculados |
| `boolean` | `*bool` | true o false |
| `integer` | `*int` | Entero con signo de 32 bits |
| `integer64` | `*int64` | Entero con signo de 64 bits (solo R5) |
| `unsignedInt` | `*uint32` | Entero sin signo de 32 bits (>= 0) |
| `positiveInt` | `*uint32` | Entero sin signo de 32 bits (>= 1) |
| `decimal` | `*Decimal` | Tipo personalizado, preserva precisión |
| `date` | `*string` | Fecha ISO 8601 (YYYY, YYYY-MM o YYYY-MM-DD) |
| `dateTime` | `*string` | Fecha/hora ISO 8601 con zona horaria opcional |
| `instant` | `*string` | Marca de tiempo precisa ISO 8601 con zona horaria |
| `time` | `*string` | Hora del día (HH:MM:SS) |
| `base64Binary` | `*string` | Datos binarios codificados en Base64 |

## Semántica de Punteros

Todos los tipos primitivos usan semántica de punteros (`*T`), donde un puntero `nil` significa que el valor está ausente del recurso. Esto es crítico para FHIR, donde la ausencia de un campo tiene un significado distinto al de que el campo esté presente con un valor por defecto o valor cero.

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

Al leer campos, siempre verifica si es `nil` antes de desreferenciar:

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

### Por Qué No Valores Cero?

Los valores cero de Go (`""` para cadenas, `false` para booleanos, `0` para enteros) no pueden distinguir entre "ausente" y "presente con valor cero". En FHIR:

- Un campo booleano establecido en `false` es diferente de un campo booleano ausente.
- Un campo entero establecido en `0` es diferente de un campo entero ausente.
- Un campo cadena establecido en `""` es diferente de un campo cadena ausente.

Los punteros resuelven esto de forma limpia: `nil` significa ausente, y `ptrTo(false)` significa explícitamente `false`.

## Tipos Similares a Cadenas

FHIR define muchos tipos primitivos similares a cadenas (`uri`, `url`, `canonical`, `id`, `oid`, `uuid`, `markdown`) que todos se mapean a `*string` en Go. La información del tipo FHIR se captura en los comentarios de los campos del struct y las etiquetas JSON, pero el tipo de Go es el mismo.

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

## Tipos de Fecha y Hora

FHIR tiene cuatro primitivos de fecha/hora, todos mapeados a `*string` en Go. Los valores de cadena deben conformarse al formato ISO 8601, pero la biblioteca no realiza validación de fechas a nivel de tipo.

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

| Tipo FHIR | Ejemplos de Formato |
|-----------|---------------------|
| `date` | `"2024"`, `"2024-03"`, `"2024-03-15"` |
| `dateTime` | `"2024-03-15"`, `"2024-03-15T10:30:00Z"`, `"2024-03-15T10:30:00+01:00"` |
| `instant` | `"2024-03-15T10:30:00.123Z"` (requiere precisión completa con zona horaria) |
| `time` | `"10:30:00"`, `"10:30:00.123"` |

## Tipos Numéricos

### Tipos Enteros

FHIR define tres tipos enteros con diferentes mapeos en Go:

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

El tipo `decimal` de FHIR se mapea al tipo personalizado `*Decimal` en lugar de `*float64`, para preservar la precisión. Consulta la página de [Precisión Decimal](../decimal-precision) para todos los detalles.

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

## Tipos de Código

Cuando un campo FHIR está vinculado a un conjunto de valores requerido, la biblioteca genera un enum con seguridad de tipos en lugar de usar `*string`. Consulta la página de [Sistemas de Códigos](../code-systems) para más detalles.

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

## Elementos de Extensión

Cada campo primitivo en un recurso FHIR tiene un elemento de extensión correspondiente que puede llevar el `id` del elemento y cualquier extensión sobre el valor primitivo. En los structs de Go, estos aparecen como campos `*Element` con una etiqueta JSON prefijada con guion bajo.

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

El struct `Element` contiene un `Id` opcional y un slice de valores `Extension`:

```go
type Element struct {
    Id        *string     `json:"id,omitempty"`
    Extension []Extension `json:"extension,omitempty"`
}
```

Consulta la página de [Extensiones](../extensions) para ejemplos detallados del uso de extensiones de primitivos.

## Primitivos Repetidos

Cuando un campo primitivo puede repetirse (cardinalidad `0..*`), usa un slice de Go del tipo base en lugar de un puntero:

```go
type HumanName struct {
    Given     []string  `json:"given,omitempty"`     // Repeating string
    GivenExt  []Element `json:"_given,omitempty"`    // Per-element extensions
    Prefix    []string  `json:"prefix,omitempty"`    // Repeating string
    PrefixExt []Element `json:"_prefix,omitempty"`   // Per-element extensions
    // ...
}
```

Un slice vacío o `nil` significa "no hay valores presentes". Cada elemento en el slice de extensiones corresponde posicionalmente al elemento en el mismo índice en el slice de valores.

{{< callout type="info" >}}
La biblioteca no realiza validación FHIR sobre valores primitivos. Por ejemplo, los campos `*string` para tipos `date` aceptan cualquier cadena, y los campos `*uint32` para `positiveInt` aceptan cero aunque FHIR requiere valores positivos. La validación debe manejarse en una capa superior.
{{< /callout >}}

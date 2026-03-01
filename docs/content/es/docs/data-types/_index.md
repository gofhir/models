---
title: "Tipos de Datos"
linkTitle: "Tipos de Datos"
description: "Descripción general del sistema de tipos FHIR en Go, incluyendo tipos primitivos, tipos complejos, sistemas de códigos y extensiones."
weight: 4
---

La especificación FHIR define un sistema de tipos rico que sustenta todos los recursos. La biblioteca `gofhir/models` mapea cada tipo de dato FHIR a una representación idiomática en Go, usando semántica de punteros para opcionalidad, tipos enum generados para valores codificados y un tipo `Decimal` personalizado para valores numéricos sensibles a la precisión.

## Categorías de Tipos FHIR

Los tipos de datos FHIR se dividen en varias categorías, cada una con una representación específica en Go en esta biblioteca:

### Tipos Primitivos

Los primitivos FHIR son los bloques de construcción básicos: cadenas, booleanos, enteros, decimales y fechas. En Go, se representan como tipos puntero (`*string`, `*bool`, `*int`, `*Decimal`) donde `nil` indica que el valor está ausente. Todos los tipos FHIR similares a cadenas (uri, url, canonical, id, oid, uuid, markdown, date, dateTime, instant, time) se mapean a `*string` en Go.

### Tipos Complejos

Los tipos complejos son tipos de datos estructurados compuestos por múltiples elementos. Ejemplos incluyen `HumanName`, `Address`, `ContactPoint`, `CodeableConcept`, `Reference` y `Quantity`. Cada uno se representa como un struct de Go con campos para cada elemento definido en la especificación FHIR.

### Sistemas de Códigos

FHIR usa valores codificados extensivamente. La biblioteca genera enums de cadena con seguridad de tipos para cada sistema de códigos FHIR, como `AdministrativeGender`, `ObservationStatus` y `BundleType`. Estos previenen valores de código inválidos en tiempo de compilación.

### El Tipo Decimal

FHIR requiere que los valores decimales preserven su precisión original (por ejemplo, `1.50` debe permanecer como `1.50`, no `1.5`). La biblioteca proporciona un tipo `Decimal` personalizado que almacena la representación de cadena original mientras soporta operaciones numéricas.

### Extensiones

El modelo de extensibilidad de FHIR usa el tipo `Extension` con una URL y un campo polimórfico `value[x]`. El tipo `Element` lleva extensiones en valores primitivos a través del patrón JSON `_fieldName`.

## Temas

{{< cards >}}
  {{< card link="primitive-types" title="Tipos Primitivos" subtitle="Mapeo de tipos FHIR a Go, semántica de punteros y manejo de nil." >}}
  {{< card link="complex-types" title="Tipos Complejos" subtitle="Tipos de datos estructurados: HumanName, CodeableConcept, Reference y más." >}}
  {{< card link="code-systems" title="Sistemas de Códigos" subtitle="Enums generados con seguridad de tipos para valores codificados FHIR." >}}
  {{< card link="decimal-precision" title="Precisión Decimal" subtitle="Tipo Decimal personalizado para preservación de precisión conforme a FHIR." >}}
  {{< card link="extensions" title="Extensiones" subtitle="Extensibilidad FHIR con tipos Extension y Element." >}}
{{< /cards >}}

## Ejemplo Rápido

Aquí tienes un ejemplo breve que ejercita varias categorías de tipos de datos juntas:

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

## Principios de Diseño del Sistema de Tipos

Las representaciones en Go de esta biblioteca siguen varios principios de diseño:

1. **Nil significa ausente.** Los tipos puntero permiten distinguir entre "no presente" (nil) y "presente con valor por defecto" (por ejemplo, `*bool` puede ser nil, true o false).

2. **Seguridad de tipos para códigos.** Los enums de cadena generados capturan valores de código inválidos en tiempo de compilación en lugar de en tiempo de ejecución.

3. **Preservación de precisión.** El tipo `Decimal` personalizado asegura que la precisión numérica nunca se pierda silenciosamente durante viajes de ida y vuelta de serialización.

4. **Soporte de extensiones.** Cada campo primitivo tiene un campo de extensión correspondiente `_fieldName` (de tipo `*Element`) que lleva extensiones de primitivos FHIR.

5. **Consistencia de etiquetas JSON.** Todas las etiquetas de struct coinciden exactamente con los nombres de propiedades JSON de FHIR, con `omitempty` en campos opcionales.

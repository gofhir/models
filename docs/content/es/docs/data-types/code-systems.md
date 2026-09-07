---
title: "Sistemas de Códigos"
linkTitle: "Sistemas de Códigos"
description: "Tipos enum generados con seguridad de tipos para valores codificados FHIR, incluyendo AdministrativeGender, ObservationStatus y más."
weight: 3
---

FHIR usa valores codificados extensivamente para representar categorías, estados, tipos y otros conceptos enumerados. La biblioteca `gofhir/models` genera tipos enum basados en cadenas con seguridad de tipos para cada sistema de códigos FHIR, proporcionando seguridad en tiempo de compilación y autocompletado del IDE para campos codificados.

## Patrón de Enum Generado

Cada sistema de códigos FHIR se representa como un alias de tipo de `string` en Go, con constantes para cada valor de código válido. Todos los tipos enum y constantes se generan en `codesystems.go` a partir de las definiciones ValueSet de FHIR.

La convención de nomenclatura sigue este patrón:

- **Nombre del tipo:** El nombre del sistema de códigos FHIR en PascalCase (p. ej., `AdministrativeGender`)
- **Nombre de la constante:** El nombre del tipo seguido del valor del código en PascalCase (p. ej., `AdministrativeGenderMale`)
- **Valor de la constante:** La cadena del código FHIR (p. ej., `"male"`)

## AdministrativeGender

Uno de los sistemas de códigos más utilizados:

```go
type AdministrativeGender string

const (
    AdministrativeGenderMale    AdministrativeGender = "male"
    AdministrativeGenderFemale  AdministrativeGender = "female"
    AdministrativeGenderOther   AdministrativeGender = "other"
    AdministrativeGenderUnknown AdministrativeGender = "unknown"
)
```

Usado en el struct `Patient`:

```go
type Patient struct {
    // ...
    Gender    *AdministrativeGender `json:"gender,omitempty"`
    GenderExt *Element             `json:"_gender,omitempty"`
    // ...
}
```

Ejemplo de uso:

```go
func ptrTo[T any](v T) *T {
    return &v
}

patient := &r4.Patient{
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

Un ejemplo completo mostrando todos los valores del sistema de códigos `ObservationStatus`:

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

Uso en un recurso Observation:

```go
package main

import (
    "fmt"

    "github.com/gofhir/models/r4/v2"
)

func ptrTo[T any](v T) *T {
    return &v
}

func main() {
    obs := &r4.Observation{
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

## Otros Sistemas de Códigos Comunes

La biblioteca genera tipos enum para todos los sistemas de códigos FHIR. Aquí tienes algunos de los más utilizados:

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

## Beneficios de la Seguridad de Tipos

Usar tipos enum generados en lugar de cadenas sin procesar proporciona varias ventajas:

### Validación en Tiempo de Compilación

El compilador detecta errores tipográficos y valores inválidos:

```go
// This compiles -- valid code
patient.Gender = ptrTo(r4.AdministrativeGenderMale)

// This would cause a compile error if you tried to assign an invalid string
// patient.Gender = ptrTo("mael")  // type mismatch
```

### Autocompletado del IDE

Cuando escribes `r4.AdministrativeGender`, tu IDE sugerirá todos los valores válidos, facilitando descubrir los códigos disponibles sin consultar la especificación FHIR.

### Código Auto-documentado

Las constantes de los enums llevan el nombre de visualización del ValueSet de FHIR
como comentario, y el mismo texto está disponible en tiempo de ejecución — ver
[Lo que cada código sabe de sí mismo](#lo-que-cada-código-sabe-de-sí-mismo) más
abajo.

```go
// AdministrativeGenderMale - Male
AdministrativeGenderMale AdministrativeGender = "male"
// AdministrativeGenderFemale - Female
AdministrativeGenderFemale AdministrativeGender = "female"
```

## Lo que cada código sabe de sí mismo

Un código por sí solo no es un coding: `"male"` no dice nada hasta que dices de qué
vocabulario viene. Todo enum generado lleva los datos que la especificación da para
cada uno de sus códigos:

```go
g := r4.AdministrativeGenderMale

g.Display()   // "Male"
g.System()    // "http://hl7.org/fhir/administrative-gender"
g.Coding()    // Coding{System: ..., Code: "male", Display: "Male"}
```

`Coding()` es lo que necesita un `CodeableConcept`, y construirlo a mano es donde
la URL del system se suele copiar mal:

```go
cc := r4.NewCodeableConceptBuilder().
    AddCoding(r4.AdministrativeGenderMale.Coding()).
    Build()

// {"coding":[{"system":"http://hl7.org/fhir/administrative-gender","code":"male","display":"Male"}]}
```

### Listar los valores

`<Tipo>Values()` devuelve todos los códigos que el tipo admite, en el orden de la
especificación — para poblar un formulario o comprobar que un `switch` los cubre
todos:

```go
for _, g := range r4.AdministrativeGenderValues() {
    fmt.Printf("%s\t%s\n", g, g.Display())
}
// male    Male
// female  Female
// other   Other
// unknown Unknown
```

### Un código que no está en el ValueSet

El tipo es un string, así que se le puede asignar cualquier cosa, incluido un valor
que llegó por la red y nadie comprobó. Deserializar no lo rechaza:

```go
json.Unmarshal([]byte(`{"resourceType":"Patient","gender":"not-a-gender"}`), &p)
// sin error
```

`Display()` recurre al propio código en lugar de a una cadena vacía, así que el
resultado sigue pudiendo mostrarse a una persona, y `System()` devuelve `""` en
lugar de afirmar un vocabulario que no se sostiene.

Para saber si un código es uno de los que define la especificación, consulta la
tabla directamente:

```go
if _, definido := r4.AdministrativeGenderTable[*p.Gender]; !definido {
    // el código está fuera del ValueSet
}
```

{{< callout type="info" >}}
Esta librería deliberadamente no responde si eso hace que el **documento** sea
inválido. Los enums se generan solo para bindings `required` —un binding
`extensible` o `preferred` deja el campo como `*string`, precisamente porque allí
un código de fuera del ValueSet es válido—, pero la conformidad es una pregunta
para [`gofhir/validator`](https://github.com/gofhir/validator), que trabaja sobre
el documento crudo. Una comprobación de pertenencia aquí se parecería demasiado a
una validación terminológica como para no confundirse con ella.
{{< /callout >}}

## Trabajo con Valores de Cadena

Dado que los tipos enum se basan en `string`, puedes convertir entre ellos y cadenas sin procesar cuando sea necesario:

```go
// From enum to string
gender := r4.AdministrativeGenderMale
s := string(gender) // "male"

// From string to enum
input := "female"
gender = r4.AdministrativeGender(input)
```

Esto es útil al leer valores de código desde fuentes externas como archivos de configuración o bases de datos.

## Serialización JSON

Los tipos enum se serializan hacia y desde sus valores de cadena en JSON, exactamente como especifica FHIR:

```go
patient := &r4.Patient{
    Gender:       ptrTo(r4.AdministrativeGenderMale),
}

data, _ := json.Marshal(patient)
// {"resourceType":"Patient","gender":"male"}

var decoded r4.Patient
json.Unmarshal(data, &decoded)
fmt.Println(*decoded.Gender) // "male"
fmt.Println(*decoded.Gender == r4.AdministrativeGenderMale) // true
```

## Serialización XML

En XML, los valores de código se codifican usando la codificación estándar de primitivos FHIR con un atributo `value`. La biblioteca maneja esto a través del helper genérico `xmlEncodePrimitiveCode`:

```xml
<Patient xmlns="http://hl7.org/fhir">
  <gender value="male"/>
</Patient>
```

{{< callout type="info" >}}
La biblioteca genera tipos enum para sistemas de códigos con vinculaciones requeridas en la especificación FHIR. Para sistemas de códigos con vinculaciones extensibles o de ejemplo, el tipo del campo es `*string` para permitir cualquier valor de código, ya que esas vinculaciones permiten códigos adicionales más allá de los definidos en el conjunto de valores.
{{< /callout >}}

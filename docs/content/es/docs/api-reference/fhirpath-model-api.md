---
title: "API del Modelo FHIRPath"
linkTitle: "API del Modelo FHIRPath"
description: "Referencia completa de API para la struct FHIRPathModelData y sus metodos de acceso."
weight: 4
---

La struct `FHIRPathModelData` proporciona metadatos en tiempo de ejecucion requeridos por los motores de expresiones FHIRPath. Esta pagina documenta cada funcion y metodo exportado con su firma completa, descripcion y ejemplo de uso.

## FHIRPathModel

Devuelve el singleton a nivel de paquete que contiene todos los metadatos de FHIRPath para esta version de FHIR.

### Firma

```go
func FHIRPathModel() *FHIRPathModelData
```

### Retorna

Un puntero a la instancia singleton de `FHIRPathModelData`. Esta instancia se inicializa al momento de cargar el paquete y es segura para acceso de lectura concurrente.

### Ejemplo

```go
import "github.com/gofhir/models/r4"

model := r4.FHIRPathModel()
// model esta listo para usar -- no se necesita inicializacion
```

---

## ChoiceTypes

Devuelve los codigos de tipo FHIR permitidos para una ruta de elemento polimorfico (choice). La ruta debe ser el nombre base sin un sufijo de tipo.

### Firma

```go
func (m *FHIRPathModelData) ChoiceTypes(path string) []string
```

### Parametros

- `path` -- La ruta base del elemento usando notacion de punto, por ejemplo, `"Observation.value"` (no `"Observation.valueQuantity"`).

### Retorna

- Un slice de strings con codigos de tipo que representan los tipos permitidos para este elemento choice.
- `nil` si la ruta no es un choice type o es desconocida.

### Ejemplo

```go
model := r4.FHIRPathModel()

types := model.ChoiceTypes("Observation.value")
// ["Quantity", "CodeableConcept", "string", "boolean", "integer",
//  "Range", "Ratio", "SampledData", "time", "dateTime", "Period"]

types = model.ChoiceTypes("Patient.deceased")
// ["boolean", "dateTime"]

types = model.ChoiceTypes("Patient.name")
// nil -- no es un choice type
```

---

## TypeOf

Devuelve el codigo de tipo FHIR para la ruta de elemento dada. Los tipos primitivos usan minusculas (`"string"`, `"boolean"`, `"dateTime"`), los tipos complejos usan PascalCase (`"HumanName"`, `"CodeableConcept"`).

### Firma

```go
func (m *FHIRPathModelData) TypeOf(path string) string
```

### Parametros

- `path` -- La ruta del elemento usando notacion de punto, por ejemplo, `"Patient.name"`, `"Observation.status"`.

### Retorna

- El codigo de tipo FHIR como string.
- Un string vacio si la ruta es desconocida.

### Ejemplo

```go
model := r4.FHIRPathModel()

model.TypeOf("Patient.name")              // "HumanName"
model.TypeOf("Patient.active")            // "boolean"
model.TypeOf("Observation.status")        // "code"
model.TypeOf("Observation.valueQuantity") // "Quantity"
model.TypeOf("Patient.contact")           // "BackboneElement"
model.TypeOf("Unknown.path")             // ""
```

---

## ReferenceTargets

Devuelve los nombres de tipos de recurso objetivo permitidos para un elemento Reference o canonical en la ruta dada.

### Firma

```go
func (m *FHIRPathModelData) ReferenceTargets(path string) []string
```

### Parametros

- `path` -- La ruta del elemento de un campo de tipo Reference, por ejemplo, `"Observation.subject"`.

### Retorna

- Un slice de nombres de tipos de recurso que son objetivos validos para esta referencia.
- `nil` si la ruta no es una referencia o no tiene objetivos restringidos.

### Ejemplo

```go
model := r4.FHIRPathModel()

targets := model.ReferenceTargets("Observation.subject")
// ["Device", "Group", "Location", "Patient"]

targets = model.ReferenceTargets("Observation.performer")
// ["CareTeam", "Organization", "Patient", "Practitioner", "PractitionerRole", "RelatedPerson"]

targets = model.ReferenceTargets("Patient.name")
// nil -- no es un campo Reference
```

---

## ParentType

Devuelve el nombre del tipo padre inmediato en la jerarquia de tipos FHIR.

### Firma

```go
func (m *FHIRPathModelData) ParentType(typeName string) string
```

### Parametros

- `typeName` -- Un nombre de tipo FHIR, por ejemplo, `"Patient"`, `"Age"`, `"DomainResource"`.

### Retorna

- El nombre del tipo padre como string.
- Un string vacio si el tipo no tiene padre (por ejemplo, el tipo raiz `Element`) o es desconocido.

### Ejemplo

```go
model := r4.FHIRPathModel()

model.ParentType("Patient")        // "DomainResource"
model.ParentType("DomainResource") // "Resource"
model.ParentType("Bundle")         // "Resource"
model.ParentType("Age")            // "Quantity"
model.ParentType("Quantity")       // "Element"
model.ParentType("Element")        // ""
```

---

## IsSubtype

Indica si `child` es el mismo tipo que, o un subtipo de, `parent` en la jerarquia de tipos FHIR. Recorre la cadena `type2Parent` desde `child` hacia arriba hasta que encuentra `parent` o alcanza la raiz.

### Firma

```go
func (m *FHIRPathModelData) IsSubtype(child, parent string) bool
```

### Parametros

- `child` -- El nombre del tipo a verificar.
- `parent` -- El nombre del tipo ancestro potencial.

### Retorna

- `true` si `child == parent` o si `child` es un descendiente de `parent` en la jerarquia de tipos.
- `false` en caso contrario.

### Ejemplo

```go
model := r4.FHIRPathModel()

model.IsSubtype("Patient", "Patient")        // true  (mismo tipo)
model.IsSubtype("Patient", "DomainResource") // true
model.IsSubtype("Patient", "Resource")       // true
model.IsSubtype("Age", "Quantity")           // true
model.IsSubtype("Quantity", "Age")           // false (padre, no hijo)
model.IsSubtype("Patient", "Observation")    // false (hermanos)
```

---

## ResolvePath

Resuelve una ruta de elemento que puede estar definida en otro lugar. Algunas rutas de elementos FHIR estan definidas en un tipo backbone compartido y referenciadas desde multiples recursos. Este metodo mapea dichas rutas a su ubicacion de definicion canonica.

### Firma

```go
func (m *FHIRPathModelData) ResolvePath(path string) string
```

### Parametros

- `path` -- La ruta del elemento a resolver.

### Retorna

- La ruta canonica si el elemento esta definido en otro lugar.
- La `path` original sin cambios si no esta remapeada.

### Ejemplo

```go
model := r4.FHIRPathModel()

// La mayoria de las rutas se resuelven a si mismas
resolved := model.ResolvePath("Patient.name")
// "Patient.name"

// Algunas rutas pueden resolverse a una definicion compartida
resolved = model.ResolvePath("Bundle.entry.resource")
// Devuelve la ruta de definicion canonica
```

---

## IsResource

Indica si el nombre de tipo dado es un tipo de recurso FHIR (a diferencia de un tipo de dato o backbone element).

### Firma

```go
func (m *FHIRPathModelData) IsResource(typeName string) bool
```

### Parametros

- `typeName` -- El nombre del tipo FHIR a verificar.

### Retorna

- `true` si el tipo es un recurso (por ejemplo, `"Patient"`, `"Bundle"`, `"Observation"`).
- `false` si es un tipo de dato (por ejemplo, `"HumanName"`, `"CodeableConcept"`) o desconocido.

### Ejemplo

```go
model := r4.FHIRPathModel()

model.IsResource("Patient")         // true
model.IsResource("Observation")     // true
model.IsResource("Bundle")          // true
model.IsResource("HumanName")       // false
model.IsResource("CodeableConcept") // false
model.IsResource("BackboneElement") // false
```

{{< callout type="info" >}}
Todos los metodos de acceso realizan busquedas simples en mapas y devuelven valores cero para claves desconocidas. Nunca devuelven errores ni generan panic. Esto los hace seguros para llamar con entrada proporcionada por el usuario sin validacion.
{{< /callout >}}

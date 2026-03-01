---
title: "Modelo FHIRPath"
linkTitle: "Modelo FHIRPath"
description: "Metadatos en tiempo de ejecucion para la evaluacion de expresiones FHIRPath, incluyendo informacion de tipos, rutas de choice types, objetivos de referencia y datos de jerarquia de tipos."
weight: 1
---

La struct `FHIRPathModelData` proporciona metadatos en tiempo de ejecucion que un motor FHIRPath necesita para evaluar correctamente expresiones contra recursos FHIR. Cada paquete de version (`r4`, `r4b`, `r5`) expone su propia instancia singleton a traves de la funcion `FHIRPathModel()`.

## El Singleton

Accede a los metadatos del modelo a traves de la funcion a nivel de paquete:

```go
import "github.com/gofhir/models/r4"

model := r4.FHIRPathModel()
```

El `*FHIRPathModelData` devuelto se inicializa una vez al momento de cargar el paquete y es seguro para uso concurrente. Contiene todos los metadatos de tipos extraidos de las StructureDefinitions de FHIR para esa version.

## Mapas de Datos Internos

La struct `FHIRPathModelData` contiene seis mapas internos, cada uno con un proposito especifico durante la evaluacion de FHIRPath:

| Mapa | Tipo | Proposito |
|------|------|-----------|
| `choiceTypePaths` | `map[string][]string` | Mapea las rutas base de elementos polimorficos a sus codigos de tipo permitidos |
| `path2Type` | `map[string]string` | Mapea cada ruta de elemento a su codigo de tipo FHIR |
| `path2RefType` | `map[string][]string` | Mapea las rutas de elementos Reference a los tipos de recurso objetivo permitidos |
| `type2Parent` | `map[string]string` | Mapea los nombres de tipo a su padre en la jerarquia de tipos FHIR |
| `pathsDefinedElsewhere` | `map[string]string` | Resuelve las rutas de elementos que estan definidas en un tipo diferente |
| `resources` | `map[string]bool` | Conjunto de todos los nombres de tipos de recurso |

Estos mapas no se exportan directamente. En su lugar, siete metodos de acceso proporcionan una API limpia para consultar los metadatos.

## Metodos de Acceso

### ChoiceTypes

Devuelve los codigos de tipo permitidos para un elemento polimorfico (choice). La ruta debe usar el nombre base sin el sufijo de tipo.

```go
model := r4.FHIRPathModel()

// Observation.value es un choice type (value[x])
types := model.ChoiceTypes("Observation.value")
// Devuelve: ["Quantity", "CodeableConcept", "string", "boolean", "integer",
//           "Range", "Ratio", "SampledData", "time", "dateTime", "Period"]
```

### TypeOf

Devuelve el codigo de tipo FHIR para una ruta de elemento dada. Los tipos primitivos usan minusculas (`"string"`, `"boolean"`), los tipos complejos usan PascalCase (`"HumanName"`, `"CodeableConcept"`).

```go
model := r4.FHIRPathModel()

t := model.TypeOf("Patient.name")
// Devuelve: "HumanName"

t = model.TypeOf("Patient.active")
// Devuelve: "boolean"
```

### ReferenceTargets

Devuelve los tipos de recurso objetivo permitidos para una ruta de elemento Reference.

```go
model := r4.FHIRPathModel()

targets := model.ReferenceTargets("Observation.subject")
// Devuelve: ["Device", "Group", "Location", "Patient"]
```

### ParentType

Devuelve el nombre del tipo padre inmediato en la jerarquia de tipos FHIR.

```go
model := r4.FHIRPathModel()

parent := model.ParentType("Patient")
// Devuelve: "DomainResource"

parent = model.ParentType("Age")
// Devuelve: "Quantity"
```

### IsSubtype

Indica si `child` es igual o un subtipo de `parent` recorriendo la jerarquia de tipos.

```go
model := r4.FHIRPathModel()

model.IsSubtype("Patient", "Resource")       // true
model.IsSubtype("Patient", "DomainResource") // true
model.IsSubtype("Age", "Quantity")           // true
model.IsSubtype("Quantity", "Age")           // false
```

### ResolvePath

Resuelve rutas de elementos que estan definidas en otro lugar (por ejemplo, backbone elements compartidos entre tipos). Si la ruta no esta definida en otro lugar, devuelve la entrada sin cambios.

```go
model := r4.FHIRPathModel()

resolved := model.ResolvePath("Bundle.entry.resource")
// Puede devolver la ruta canonica si este elemento esta definido en otro tipo
```

### IsResource

Indica si un nombre de tipo dado es un tipo de recurso FHIR.

```go
model := r4.FHIRPathModel()

model.IsResource("Patient")   // true
model.IsResource("HumanName") // false
model.IsResource("Bundle")    // true
```

## Integracion con gofhir/fhirpath

El caso de uso principal de `FHIRPathModelData` es proporcionar informacion de tipos a un motor de evaluacion FHIRPath. La biblioteca complementaria [`gofhir/fhirpath`](https://github.com/gofhir/fhirpath) acepta el modelo a traves de la opcion `WithModel`:

```go
import (
    "fmt"
    "github.com/gofhir/fhirpath"
    "github.com/gofhir/models/r4"
)

// Construir un recurso
patient := r4.NewPatient(
    r4.WithPatientId("example-1"),
    r4.WithPatientBirthDate("1990-01-15"),
)

// Evaluar una expresion FHIRPath contra el recurso
result, err := fhirpath.Evaluate(patient, "Patient.birthDate",
    fhirpath.WithModel(r4.FHIRPathModel()))
if err != nil {
    panic(err)
}
fmt.Println(result) // ["1990-01-15"]
```

El modelo permite al motor FHIRPath:

- **Resolver choice types**: Cuando una expresion hace referencia a `Observation.value`, el motor usa `ChoiceTypes` para saber que campos concretos (`valueQuantity`, `valueString`, etc.) verificar.
- **Validar rutas**: `TypeOf` confirma que una ruta como `Patient.name` existe y devuelve `HumanName`.
- **Navegar la jerarquia de tipos**: `IsSubtype` y `ParentType` soportan los operadores `is` y `as` de FHIRPath.
- **Verificar objetivos de referencia**: `ReferenceTargets` valida las llamadas `.resolve()` contra los tipos de objetivo permitidos.

{{< callout type="info" >}}
El modelo FHIRPath se genera junto con todo el demas codigo. Cuando regeneras los modelos para una nueva version de FHIR, los metadatos del modelo se actualizan automaticamente para reflejar cualquier cambio en la especificacion.
{{< /callout >}}

---
title: "Interfaces de Recursos"
linkTitle: "Interfaces de Recursos"
description: "Las interfaces Resource y DomainResource que permiten el manejo generico de todos los tipos de recursos FHIR."
weight: 1
---

La biblioteca `gofhir/models` define dos interfaces principales que reflejan la jerarquia de tipos FHIR. Estas interfaces permiten el manejo generico de recursos, colecciones polimorficas e infraestructura de servidor type-safe sin recurrir a `interface{}`.

## Interfaz Resource

La interfaz `Resource` es implementada por cada tipo de recurso FHIR. Proporciona acceso a los campos base definidos en el tipo abstracto `Resource` de FHIR.

```go
type Resource interface {
    GetResourceType() string
    GetId() *string
    SetId(string)
    GetMeta() *Meta
    SetMeta(*Meta)
}
```

### Metodos

| Metodo | Tipo de Retorno | Descripcion |
|--------|----------------|-------------|
| `GetResourceType()` | `string` | Devuelve el nombre del tipo de recurso FHIR (por ejemplo, `"Patient"`, `"Observation"`) |
| `GetId()` | `*string` | Devuelve el ID logico del recurso, o `nil` si no esta establecido |
| `SetId(string)` | -- | Establece el ID logico del recurso |
| `GetMeta()` | `*Meta` | Devuelve los metadatos del recurso (version, lastUpdated, perfiles, tags, etiquetas de seguridad) |
| `SetMeta(*Meta)` | -- | Establece los metadatos del recurso |

### Ejemplo

```go
import "github.com/gofhir/models/r4"

func printResourceInfo(res r4.Resource) {
    fmt.Printf("Type: %s\n", res.GetResourceType())
    if id := res.GetId(); id != nil {
        fmt.Printf("ID: %s\n", *id)
    }
    if meta := res.GetMeta(); meta != nil && meta.VersionId != nil {
        fmt.Printf("Version: %s\n", *meta.VersionId)
    }
}

// Funciona con cualquier tipo de recurso
patient := r4.NewPatient(r4.WithPatientId("p-123"))
printResourceInfo(patient) // Type: Patient, ID: p-123

obs := r4.NewObservation(r4.WithObservationId("obs-456"))
printResourceInfo(obs) // Type: Observation, ID: obs-456
```

## Interfaz DomainResource

La interfaz `DomainResource` extiende `Resource` con campos del tipo abstracto `DomainResource` de FHIR. Agrega acceso al texto narrativo, recursos contenidos y extensiones.

```go
type DomainResource interface {
    Resource
    GetText() *Narrative
    SetText(*Narrative)
    GetContained() []Resource
    GetExtension() []Extension
    GetModifierExtension() []Extension
}
```

### Metodos

| Metodo | Tipo de Retorno | Descripcion |
|--------|----------------|-------------|
| `GetText()` | `*Narrative` | Devuelve la narrativa XHTML legible por humanos |
| `SetText(*Narrative)` | -- | Establece la narrativa legible por humanos |
| `GetContained()` | `[]Resource` | Devuelve la lista de recursos contenidos (en linea) |
| `GetExtension()` | `[]Extension` | Devuelve las extensiones estandar |
| `GetModifierExtension()` | `[]Extension` | Devuelve las extensiones modificadoras que cambian el significado del recurso |

### Ejemplo

```go
import "github.com/gofhir/models/r4"

func extractNarrative(res r4.DomainResource) string {
    if text := res.GetText(); text != nil && text.Div != nil {
        return *text.Div
    }
    return ""
}

func listContainedTypes(res r4.DomainResource) []string {
    var types []string
    for _, contained := range res.GetContained() {
        types = append(types, contained.GetResourceType())
    }
    return types
}
```

## Que Recursos Implementan Cada Interfaz

En FHIR R4, los 148 tipos de recurso implementan la interfaz `Resource`. La mayoria de ellos tambien implementan `DomainResource`. Las excepciones son los tres recursos de infraestructura que heredan directamente de `Resource` en lugar de `DomainResource`:

| Recurso | Implementa `Resource` | Implementa `DomainResource` | Razon |
|---------|:---------------------:|:---------------------------:|-------|
| `Bundle` | Si | No | Contenedor de otros recursos, no es un concepto de dominio |
| `Binary` | Si | No | Contenido binario crudo, sin narrativa ni extensiones |
| `Parameters` | Si | No | Contenedor de entrada/salida de operaciones |
| Todos los demas (~145) | Si | Si | Recursos de dominio estandar |

### Type Assertion

Puedes usar type assertions de Go para verificar si un recurso es un `DomainResource`:

```go
func processResource(res r4.Resource) {
    fmt.Printf("Processing %s/%s\n", res.GetResourceType(), safeId(res))

    if dr, ok := res.(r4.DomainResource); ok {
        // Este es un DomainResource -- podemos acceder a text, contained, extensions
        if text := dr.GetText(); text != nil {
            fmt.Println("Has narrative text")
        }
        if exts := dr.GetExtension(); len(exts) > 0 {
            fmt.Printf("Has %d extensions\n", len(exts))
        }
    }
}

func safeId(res r4.Resource) string {
    if id := res.GetId(); id != nil {
        return *id
    }
    return "<no id>"
}
```

## Colecciones Genericas

Las interfaces facilitan el trabajo con colecciones heterogeneas de recursos:

```go
func storeResources(resources []r4.Resource) {
    for _, res := range resources {
        data, err := r4.Marshal(res)
        if err != nil {
            log.Printf("Failed to marshal %s: %v", res.GetResourceType(), err)
            continue
        }
        // Almacenar datos indexados por tipo e ID
        key := fmt.Sprintf("%s/%s", res.GetResourceType(), safeId(res))
        store.Put(key, data)
    }
}
```

{{< callout type="info" >}}
Las interfaces `Resource` y `DomainResource` se generan a partir de la especificacion FHIR y son identicas en R4, R4B y R5 en terminos de sus firmas de metodos. Los tipos concretos que referencian (`Meta`, `Narrative`, `Extension`) son especificos de la version, por lo que `r4.Resource` y `r5.Resource` son tipos de interfaz distintos.
{{< /callout >}}

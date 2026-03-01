---
title: "Funciones del Registro"
linkTitle: "Funciones del Registro"
description: "Funciones de fabrica, deserializacion e introspeccion para el manejo dinamico de recursos desde el registro."
weight: 2
---

El registro proporciona un conjunto de funciones para trabajar con recursos FHIR de forma dinamica -- cuando el tipo de recurso no se conoce en tiempo de compilacion. Estas funciones estan respaldadas por un mapa de fabrica interno que asocia cada nombre de tipo de recurso con una funcion constructora.

Todas las funciones del registro se exportan desde el paquete de version (por ejemplo, `r4.NewResource`, `r4.UnmarshalResource`).

## NewResource

Crea una nueva instancia vacia del tipo de recurso especificado.

### Firma

```go
func NewResource(resourceType string) (Resource, error)
```

### Parametros

- `resourceType` -- El nombre del tipo de recurso FHIR, por ejemplo, `"Patient"`, `"Observation"`. Sensible a mayusculas y minusculas.

### Retorna

- Una interfaz `Resource` apuntando a una struct recien asignada y con valor cero del tipo solicitado.
- Un error si el nombre del tipo de recurso no es reconocido.

### Ejemplo

```go
import "github.com/gofhir/models/r4"

res, err := r4.NewResource("Patient")
if err != nil {
    log.Fatal(err)
}
res.SetId("new-patient-1")

// Type-assert si necesitas el tipo concreto
patient := res.(*r4.Patient)
fmt.Println(patient.GetResourceType()) // "Patient"
```

---

## UnmarshalResource

Deserializa un slice de bytes JSON en la struct de recurso correcta. Primero inspecciona el campo `resourceType` en el JSON para determinar el tipo, luego deserializa el payload completo en la struct Go apropiada.

### Firma

```go
func UnmarshalResource(data []byte) (Resource, error)
```

### Parametros

- `data` -- Un slice de bytes JSON que contiene un recurso FHIR con un campo `resourceType`.

### Retorna

- Una interfaz `Resource` apuntando a la struct completamente poblada.
- Un error si el JSON es invalido, el campo `resourceType` falta, o el tipo no es reconocido.

### Ejemplo

```go
import "github.com/gofhir/models/r4"

jsonData := []byte(`{
    "resourceType": "Patient",
    "id": "example",
    "name": [{"family": "Smith", "given": ["John"]}]
}`)

res, err := r4.UnmarshalResource(jsonData)
if err != nil {
    log.Fatal(err)
}

patient := res.(*r4.Patient)
fmt.Println(*patient.Name[0].Family) // "Smith"
```

---

## GetResourceType

Extrae el campo `resourceType` de un slice de bytes JSON sin deserializar completamente el recurso. Esto es util para enrutamiento o validacion antes de comprometerse con la deserializacion completa.

### Firma

```go
func GetResourceType(data []byte) (string, error)
```

### Parametros

- `data` -- Un slice de bytes JSON que debe contener un campo `resourceType`.

### Retorna

- El nombre del tipo de recurso como string.
- Un error si el JSON es invalido o el campo `resourceType` falta o esta vacio.

### Ejemplo

```go
import "github.com/gofhir/models/r4"

data := []byte(`{"resourceType": "Observation", "id": "123"}`)

rt, err := r4.GetResourceType(data)
if err != nil {
    log.Fatal(err)
}
fmt.Println(rt) // "Observation"

// Usar para enrutamiento antes de la deserializacion completa
switch rt {
case "Patient":
    handlePatient(data)
case "Observation":
    handleObservation(data)
default:
    handleGeneric(data)
}
```

---

## IsKnownResourceType

Verifica si un nombre de tipo de recurso dado es reconocido por el registro.

### Firma

```go
func IsKnownResourceType(resourceType string) bool
```

### Parametros

- `resourceType` -- El nombre del tipo de recurso a verificar. Sensible a mayusculas y minusculas.

### Retorna

- `true` si el tipo es un recurso FHIR conocido para esta version.
- `false` en caso contrario.

### Ejemplo

```go
import "github.com/gofhir/models/r4"

r4.IsKnownResourceType("Patient")      // true
r4.IsKnownResourceType("Observation")  // true
r4.IsKnownResourceType("HumanName")    // false (tipo de dato, no un recurso)
r4.IsKnownResourceType("FakeResource") // false
```

---

## AllResourceTypes

Devuelve un slice que contiene todos los nombres de tipos de recurso conocidos para esta version de FHIR.

### Firma

```go
func AllResourceTypes() []string
```

### Retorna

- Un slice con todos los nombres de tipos de recurso. El orden no esta garantizado.

### Ejemplo

```go
import (
    "fmt"
    "sort"
    "github.com/gofhir/models/r4"
)

types := r4.AllResourceTypes()
sort.Strings(types)

fmt.Printf("R4 defines %d resource types\n", len(types))
// R4 defines 148 resource types

for _, t := range types[:5] {
    fmt.Println(t)
}
// Account
// ActivityDefinition
// AdverseEvent
// AllergyIntolerance
// Appointment
```

{{< callout type="info" >}}
El registro se inicializa al momento de cargar el paquete desde un mapa en tiempo de compilacion. Todas las funciones del registro son seguras para uso concurrente y no requieren ninguna inicializacion mas alla de importar el paquete.
{{< /callout >}}

## Patrones Comunes

### Router de Recursos para Servidor FHIR

```go
import (
    "net/http"
    "github.com/gofhir/models/r4"
)

func handleCreate(w http.ResponseWriter, req *http.Request) {
    data, _ := io.ReadAll(req.Body)

    rt, err := r4.GetResourceType(data)
    if err != nil {
        http.Error(w, "Invalid FHIR resource", http.StatusBadRequest)
        return
    }

    if !r4.IsKnownResourceType(rt) {
        http.Error(w, "Unknown resource type: "+rt, http.StatusBadRequest)
        return
    }

    resource, err := r4.UnmarshalResource(data)
    if err != nil {
        http.Error(w, "Failed to parse resource", http.StatusBadRequest)
        return
    }

    // Procesar el recurso...
    resource.SetId(generateId())
    // Almacenar, validar, etc.
}
```

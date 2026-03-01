---
title: "Deserialización Polimórfica"
linkTitle: "Deserialización Polimórfica"
description: "Funciones del registro de recursos para resolución dinámica de tipos desde datos FHIR JSON y XML sin procesar."
weight: 4
---

Al trabajar con datos FHIR donde el tipo de recurso no se conoce en tiempo de compilación -- como al leer de una base de datos, procesar entradas de Bundle, o recibir cargas FHIR arbitrarias desde una API -- la biblioteca proporciona un registro de recursos que permite el despacho dinámico al struct de Go correcto.

## Funciones del Registro de Recursos

El registro de recursos se define en `registry.go` y proporciona cinco funciones para trabajar con tipos de recursos FHIR en tiempo de ejecución.

### UnmarshalResource

```go
func UnmarshalResource(data []byte) (Resource, error)
```

Deserializa bytes JSON sin procesar en el tipo de recurso concreto correcto. Primero inspecciona el campo `resourceType` para determinar el tipo, crea una instancia vacía a través del registro de fábricas, y luego deserializa el JSON completo en esa instancia.

```go
package main

import (
    "fmt"
    "log"

    "github.com/gofhir/models/r4"
)

func main() {
    jsonData := []byte(`{
        "resourceType": "Patient",
        "id": "example-1",
        "gender": "female",
        "birthDate": "1985-03-22"
    }`)

    resource, err := r4.UnmarshalResource(jsonData)
    if err != nil {
        log.Fatal(err)
    }

    // Use type assertion to access type-specific fields
    if patient, ok := resource.(*r4.Patient); ok {
        fmt.Println(*patient.Id)        // "example-1"
        fmt.Println(*patient.Gender)    // "female"
        fmt.Println(*patient.BirthDate) // "1985-03-22"
    }
}
```

### GetResourceType

```go
func GetResourceType(data []byte) (string, error)
```

Extrae la cadena `resourceType` del JSON sin procesar sin realizar una deserialización completa. Esto es útil para enrutamiento, filtrado o validación antes de comprometerse con una operación de unmarshal completa.

```go
jsonData := []byte(`{"resourceType": "Observation", "id": "obs-1"}`)

resourceType, err := r4.GetResourceType(jsonData)
if err != nil {
    log.Fatal(err)
}
fmt.Println(resourceType) // "Observation"
```

Esta función solo analiza la estructura mínima necesaria para leer el campo `resourceType`, haciéndola eficiente para escenarios de alto rendimiento donde necesitas inspeccionar o enrutar recursos antes de deserializarlos.

### NewResource

```go
func NewResource(resourceType string) (Resource, error)
```

Crea una nueva instancia vacía del tipo de recurso especificado. Retorna un error si el nombre del tipo no es reconocido.

```go
resource, err := r4.NewResource("Patient")
if err != nil {
    log.Fatal(err) // "unknown resource type: ..."
}

patient := resource.(*r4.Patient)
patient.ResourceType = "Patient"
patient.Id = ptrTo("new-patient")
```

Esta función es la base para las otras funciones del registro. Busca el nombre del tipo en el mapa interno `resourceFactories` y llama a la función de fábrica correspondiente.

### IsKnownResourceType

```go
func IsKnownResourceType(resourceType string) bool
```

Retorna `true` si el nombre del tipo de recurso dado está registrado en la fábrica. Útil para validación de entrada antes de intentar la deserialización.

```go
fmt.Println(r4.IsKnownResourceType("Patient"))      // true
fmt.Println(r4.IsKnownResourceType("Observation"))   // true
fmt.Println(r4.IsKnownResourceType("FakeResource"))  // false
```

### AllResourceTypes

```go
func AllResourceTypes() []string
```

Retorna un slice con todos los nombres de tipos de recursos registrados. El orden no está garantizado.

```go
types := r4.AllResourceTypes()
fmt.Println(len(types)) // 146 (for R4)

for _, t := range types {
    fmt.Println(t)
}
```

## Patrones de Aserción de Tipo

Dado que `UnmarshalResource` retorna la interfaz `Resource`, necesitas aserciones de tipo para acceder a campos específicos del recurso. Aquí tienes los patrones más comunes:

### Aserción de Tipo Único

Cuando esperas un tipo específico:

```go
resource, err := r4.UnmarshalResource(jsonData)
if err != nil {
    log.Fatal(err)
}

patient, ok := resource.(*r4.Patient)
if !ok {
    log.Fatalf("expected Patient, got %s", resource.GetResourceType())
}
fmt.Println(*patient.Id)
```

### Switch de Tipo

Al manejar múltiples tipos de recursos:

```go
resource, err := r4.UnmarshalResource(jsonData)
if err != nil {
    log.Fatal(err)
}

switch r := resource.(type) {
case *r4.Patient:
    fmt.Printf("Patient: %s\n", *r.Id)
case *r4.Observation:
    fmt.Printf("Observation: %s\n", *r.Id)
case *r4.Encounter:
    fmt.Printf("Encounter: %s\n", *r.Id)
default:
    fmt.Printf("Other resource: %s\n", r.GetResourceType())
}
```

### Uso de la Interfaz Resource

Para operaciones que aplican a todos los recursos, usa los métodos de la interfaz `Resource` sin aserción de tipo:

```go
resource, _ := r4.UnmarshalResource(jsonData)

// These methods are available on all resources
fmt.Println(resource.GetResourceType()) // e.g., "Patient"
fmt.Println(*resource.GetId())          // e.g., "123"

meta := resource.GetMeta()
if meta != nil {
    fmt.Println(*meta.VersionId)
}
```

La interfaz `Resource` proporciona:

```go
type Resource interface {
    GetResourceType() string
    GetId() *string
    SetId(string)
    GetMeta() *Meta
    SetMeta(*Meta)
}
```

Para recursos de dominio (recursos con narrativa y extensiones), también puedes hacer aserción a `DomainResource`:

```go
if dr, ok := resource.(r4.DomainResource); ok {
    text := dr.GetText()
    extensions := dr.GetExtension()
    contained := dr.GetContained()
    // ...
}
```

## Procesamiento de Entradas de Bundle

Un caso de uso común es procesar entradas de un Bundle FHIR. El campo `resource` de cada entrada puede contener cualquier tipo de recurso:

```go
bundleJSON := []byte(`{
    "resourceType": "Bundle",
    "type": "searchset",
    "entry": [
        {"resource": {"resourceType": "Patient", "id": "p1"}},
        {"resource": {"resourceType": "Observation", "id": "o1"}}
    ]
}`)

var bundle r4.Bundle
if err := json.Unmarshal(bundleJSON, &bundle); err != nil {
    log.Fatal(err)
}

for _, entry := range bundle.Entry {
    if entry.Resource != nil {
        fmt.Printf("Type: %s, ID: %s\n",
            entry.Resource.GetResourceType(),
            *entry.Resource.GetId(),
        )
    }
}
```

## Deserialización Polimórfica XML

El registro también soporta deserialización XML a través de `UnmarshalResourceXML`:

```go
xmlData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<Observation xmlns="http://hl7.org/fhir">
  <id value="obs-1"/>
  <status value="final"/>
</Observation>`)

resource, err := r4.UnmarshalResourceXML(xmlData)
if err != nil {
    log.Fatal(err)
}

obs := resource.(*r4.Observation)
fmt.Println(*obs.Id)     // "obs-1"
fmt.Println(*obs.Status) // "final"
```

`UnmarshalResourceXML` lee el nombre del elemento raíz (por ejemplo, `<Observation>`) para determinar el tipo de recurso, y luego delega a la misma fábrica `NewResource` utilizada por la deserialización JSON.

## Patrón de Enrutamiento

Combina `GetResourceType` con `NewResource` para un enrutamiento eficiente de recursos:

```go
func handleFHIRResource(data []byte) error {
    resourceType, err := r4.GetResourceType(data)
    if err != nil {
        return fmt.Errorf("cannot determine resource type: %w", err)
    }

    if !r4.IsKnownResourceType(resourceType) {
        return fmt.Errorf("unsupported resource type: %s", resourceType)
    }

    resource, err := r4.UnmarshalResource(data)
    if err != nil {
        return fmt.Errorf("failed to unmarshal %s: %w", resourceType, err)
    }

    switch resourceType {
    case "Patient":
        return processPatient(resource.(*r4.Patient))
    case "Observation":
        return processObservation(resource.(*r4.Observation))
    default:
        return processGenericResource(resource)
    }
}
```

{{< callout type="info" >}}
El registro de recursos se genera en tiempo de generación de código e incluye todos los tipos de recursos definidos en la versión FHIR correspondiente. Para R4, esto incluye 146 tipos de recursos. Los paquetes R4B y R5 tienen cada uno sus propios registros con los tipos de recursos definidos en esas versiones.
{{< /callout >}}

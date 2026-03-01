---
title: "Campos de Resumen"
linkTitle: "Campos de Resumen"
description: "Listas de campos de resumen precalculadas para implementar el comportamiento del parametro de busqueda _summary=true de FHIR en un servidor FHIR."
weight: 2
---

La variable `SummaryFields` es un mapa a nivel de paquete que lista, para cada tipo de recurso, los campos que la especificacion FHIR marca con `isSummary=true`. Estos datos son esenciales para implementar el parametro de busqueda `_summary=true` en un servidor FHIR.

## Que son los Campos de Resumen?

La especificacion FHIR define un parametro de busqueda `_summary` que permite a los clientes solicitar versiones abreviadas de los recursos. Cuando un cliente envia `_summary=true`, el servidor debe devolver solo los campos marcados como campos de resumen en la StructureDefinition de ese tipo de recurso, ademas de algunos elementos obligatorios (`id`, `meta`, `resourceType`).

Esto es util para reducir el tamano de las respuestas en resultados de busqueda donde el cliente solo necesita identificadores y atributos clave en lugar del recurso completo.

## El Mapa SummaryFields

Cada paquete de version exporta una variable `SummaryFields`:

```go
var SummaryFields = map[string][]string{
    "Patient": {
        "active",
        "address",
        "birthDate",
        "communication",
        "gender",
        "generalPractitioner",
        "id",
        "identifier",
        "implicitRules",
        "link",
        "managingOrganization",
        "meta",
        "name",
        "telecom",
    },
    // ... todos los demas tipos de recurso
}
```

El mapa esta indexado por nombre de tipo de recurso (por ejemplo, `"Patient"`, `"Observation"`) y los valores son slices ordenados de nombres de campo que deben incluirse en una respuesta de resumen.

## Uso

### Consulta Basica

```go
import "github.com/gofhir/models/r4"

fields := r4.SummaryFields["Patient"]
// Devuelve: ["active", "address", "birthDate", "communication", "gender",
//           "generalPractitioner", "id", "identifier", "implicitRules",
//           "link", "managingOrganization", "meta", "name", "telecom"]
```

### Verificar si un Campo es un Campo de Resumen

```go
import "github.com/gofhir/models/r4"

func isSummaryField(resourceType, fieldName string) bool {
    fields, ok := r4.SummaryFields[resourceType]
    if !ok {
        return false
    }
    for _, f := range fields {
        if f == fieldName {
            return true
        }
    }
    return false
}

isSummaryField("Patient", "name")      // true
isSummaryField("Patient", "photo")     // false
isSummaryField("Observation", "code")  // true
isSummaryField("Observation", "note")  // false
```

### Implementacion en un Servidor FHIR

Un servidor FHIR tipico usa `SummaryFields` para filtrar los campos de un recurso antes de devolver los resultados de busqueda:

```go
import (
    "encoding/json"
    "github.com/gofhir/models/r4"
)

func applySummary(resourceType string, data []byte) ([]byte, error) {
    summaryFields := r4.SummaryFields[resourceType]
    if summaryFields == nil {
        return data, nil // tipo desconocido, devolver tal cual
    }

    // Construir un conjunto para busqueda rapida
    allowed := make(map[string]bool, len(summaryFields))
    for _, f := range summaryFields {
        allowed[f] = true
    }
    // Siempre incluir resourceType
    allowed["resourceType"] = true

    // Parsear, filtrar y re-serializar
    var full map[string]json.RawMessage
    if err := json.Unmarshal(data, &full); err != nil {
        return nil, err
    }

    filtered := make(map[string]json.RawMessage)
    for key, val := range full {
        if allowed[key] {
            filtered[key] = val
        }
    }

    return json.Marshal(filtered)
}
```

## Cobertura

El mapa `SummaryFields` incluye entradas para cada tipo de recurso definido en la version de FHIR. En R4, esto cubre los 148 tipos de recurso. Cada entrada se genera directamente desde el flag `isSummary` en las StructureDefinitions oficiales de FHIR.

{{< callout type="info" >}}
Las listas de campos de resumen se generan a partir de la especificacion FHIR y siempre se mantienen sincronizadas con las definiciones de las structs de recursos. Si se agrega o elimina un campo del conjunto de resumen en una nueva version de FHIR, el codigo regenerado reflejara el cambio automaticamente.
{{< /callout >}}

## Comparacion con _elements

El parametro `_summary=true` es un mecanismo mas general que el parametro `_elements`. Con `_elements`, los clientes especifican exactamente que campos desean. Con `_summary=true`, el servidor devuelve el conjunto de resumen definido en la especificacion. `SummaryFields` soporta este ultimo caso; implementar `_elements` requiere un enfoque diferente (tipicamente filtrado de campos JSON basado en la lista de campos proporcionada por el cliente).

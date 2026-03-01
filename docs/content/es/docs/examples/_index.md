---
title: "Ejemplos"
linkTitle: "Ejemplos"
description: "Ejemplos practicos de codigo para construir recursos FHIR con gofhir/models, desde tipos de datos comunes hasta patrones de produccion."
weight: 7
---

Esta seccion proporciona ejemplos practicos y ejecutables de codigo Go que demuestran como usar la biblioteca `gofhir/models` en escenarios del mundo real. Los ejemplos progresan desde patrones de construccion comunes hasta casos de uso de produccion completos.

## Temas

{{< cards >}}
  {{< card link="common-patterns" title="Patrones Comunes" subtitle="Construccion de patients, observations, bundles y trabajo con CodeableConcept, Coding y el paquete helpers." icon="code" >}}
  {{< card link="real-world-usage" title="Uso en el Mundo Real" subtitle="Patrones de produccion incluyendo handlers HTTP, parseo de respuestas de API, conversion de formatos y enrutamiento de recursos." icon="server" >}}
{{< /cards >}}

## Inicio Rapido

Si eres nuevo en la biblioteca, aqui esta el ejemplo mas simple posible para comenzar:

```go
package main

import (
    "encoding/json"
    "fmt"
    "github.com/gofhir/models/r4"
)

func main() {
    // Crear un Patient usando opciones funcionales
    patient := r4.NewPatient(
        r4.WithPatientId("hello-fhir"),
        r4.WithPatientActive(true),
        r4.WithPatientName(r4.HumanName{
            Family: ptrTo("World"),
            Given:  []string{"Hello"},
        }),
    )

    // Serializar a JSON
    data, _ := r4.Marshal(patient)
    fmt.Println(string(data))
}

func ptrTo[T any](v T) *T {
    return &v
}
```

Salida:

```json
{
  "resourceType": "Patient",
  "id": "hello-fhir",
  "active": true,
  "name": [{"family": "World", "given": ["Hello"]}]
}
```

Todos los ejemplos en esta seccion usan el paquete `r4`. Los mismos patrones aplican a `r4b` y `r5`.

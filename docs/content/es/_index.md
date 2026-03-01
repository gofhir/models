---
title: "Modelos FHIR para Go"
description: "Structs Go con tipado seguro para recursos FHIR R4, R4B y R5 con builders, serialización y metadatos del modelo FHIRPath."
layout: hextra-home
---

<div class="hx:text-center hx:mt-24 hx:mb-6">
{{< hextra/hero-badge >}}
  <span>Open Source</span>
  {{< icon name="github" attributes="height=14" >}}
{{< /hextra/hero-badge >}}
</div>

<div class="hx:mt-6 hx:mb-6">
{{< hextra/hero-headline >}}
  Modelos FHIR para Go
{{< /hextra/hero-headline >}}
</div>

<div class="hx:mb-12">
{{< hextra/hero-subtitle >}}
  Structs Go con tipado seguro para todos los recursos FHIR R4, R4B y R5.&nbsp;<br class="sm:hx:block hx:hidden" />Construye, serializa e integra con builders fluidos y soporte completo de JSON/XML.
{{< /hextra/hero-subtitle >}}
</div>

<div class="hx:mb-6">
{{< hextra/hero-button text="Comenzar" link="docs/getting-started" >}}
{{< hextra/hero-button text="Ver en GitHub" link="https://github.com/gofhir/models" style="alt" >}}
</div>

<div class="hx:mt-6"></div>

{{< hextra/feature-grid >}}
  {{< hextra/feature-card
    title="Todas las Versiones FHIR"
    icon="collection"
    subtitle="Structs Go completos para cada recurso y tipo de dato en FHIR R4 (4.0.1), R4B (4.3.0) y R5 (5.0.0). Cada versión es un módulo Go independiente con su propio ciclo de lanzamiento."
  >}}
  {{< hextra/feature-card
    title="Tres Patrones de Construcción"
    icon="puzzle"
    subtitle="Crea recursos FHIR a tu manera: literales de struct directos, cadenas de builder fluido u opciones funcionales. Cada patrón produce los mismos structs con tipado seguro."
  >}}
  {{< hextra/feature-card
    title="Serialización JSON y XML"
    icon="code"
    subtitle="Serializa y deserializa recursos con total conformidad FHIR. El marshaling JSON seguro para HTML preserva el XHTML narrativo, y los recursos contenidos polimórficos se manejan automáticamente."
  >}}
{{< /hextra/feature-grid >}}

## Inicio Rápido

{{< callout type="info" >}}
  Requiere **Go 1.23** o posterior.
{{< /callout >}}

Instala el paquete para la versión FHIR que necesites:

```shell
go get github.com/gofhir/models/r4
```

Crea un recurso Patient y serialízalo a JSON:

```go
package main

import (
    "fmt"

    "github.com/gofhir/models/r4"
)

func main() {
    patient := r4.NewPatient(
        r4.WithPatientId("example"),
        r4.WithPatientActive(true),
        r4.WithPatientGender(r4.AdministrativeGenderMale),
    )

    data, _ := r4.Marshal(patient)
    fmt.Println(string(data))
}
```

Salida:

```json
{"resourceType":"Patient","id":"example","active":true,"gender":"male"}
```

{{< hextra/hero-button text="Leer la guía completa" link="docs/getting-started" >}}

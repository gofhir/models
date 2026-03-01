---
title: "Instalación"
linkTitle: "Instalación"
description: "Instala gofhir/models para FHIR R4, R4B o R5 y configura tu proyecto Go."
weight: 1
---

## Requisitos Previos

- Se requiere **Go 1.23** o posterior. Puedes verificar tu versión con:

```shell
go version
```

- Tu proyecto debe usar [Go modules](https://go.dev/ref/mod) (`go.mod`). Si aún no tienes uno, inicialízalo:

```shell
go mod init your-module-name
```

## Instalar

Cada versión FHIR se publica como un módulo Go independiente. Instala solo la versión que necesites.

### FHIR R4 (4.0.1)

```shell
go get github.com/gofhir/models/r4
```

### FHIR R4B (4.3.0)

```shell
go get github.com/gofhir/models/r4b
```

### FHIR R5 (5.0.0)

```shell
go get github.com/gofhir/models/r5
```

### R4 Helpers (opcional)

El sub-paquete helpers proporciona valores `CodeableConcept` preconstruidos para categorías comunes de observación, categorías de condición, códigos LOINC y unidades UCUM:

```shell
go get github.com/gofhir/models/r4/helpers
```

## Importar

Después de instalar, importa el paquete en tus archivos fuente Go:

```go
import "github.com/gofhir/models/r4"
```

O para R4B y R5:

```go
import "github.com/gofhir/models/r4b"
import "github.com/gofhir/models/r5"
```

Todos los tipos, builders, opciones funcionales, constantes de sistemas de códigos y funciones de serialización se exportan desde el paquete específico de la versión. No hay un sub-paquete separado para builders o serialización -- todo está en un solo lugar.

## Usar Múltiples Versiones FHIR

Si tu aplicación necesita trabajar con más de una versión FHIR, puedes importar múltiples paquetes en el mismo proyecto. Usa alias de importación de Go para evitar colisiones de nombres:

```go
import (
    r4 "github.com/gofhir/models/r4"
    r5 "github.com/gofhir/models/r5"
)

func main() {
    // R4 Patient
    patientR4 := r4.NewPatient(
        r4.WithPatientId("r4-patient"),
        r4.WithPatientActive(true),
    )

    // R5 Patient
    patientR5 := r5.NewPatient(
        r5.WithPatientId("r5-patient"),
        r5.WithPatientActive(true),
    )

    _, _ = r4.Marshal(patientR4)
    _, _ = r5.Marshal(patientR5)
}
```

Dado que cada versión FHIR es un módulo Go separado con su propio `go.mod`, las versiones de dependencias se resuelven de forma independiente y nunca entran en conflicto.

## Verificar la Instalación

Crea un archivo de prueba simple para confirmar que todo funciona:

```go
package main

import (
    "fmt"

    "github.com/gofhir/models/r4"
)

func main() {
    patient := r4.NewPatient(
        r4.WithPatientId("test"),
    )
    data, err := r4.Marshal(patient)
    if err != nil {
        panic(err)
    }
    fmt.Println(string(data))
}
```

Ejecútalo:

```shell
go run main.go
```

Salida esperada:

```json
{"resourceType":"Patient","id":"test"}
```

## Siguientes Pasos

Continúa con la guía de [Inicio Rápido](../quick-start) para ver ejemplos completos de creación, serialización y deserialización de recursos FHIR.

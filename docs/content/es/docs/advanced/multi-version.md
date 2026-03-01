---
title: "Soporte Multi-Version"
linkTitle: "Multi-Version"
description: "Arquitectura de workspace de Go para importar multiples versiones de FHIR en paralelo con versionado independiente de modulos."
weight: 4
---

El repositorio `gofhir/models` publica cada version de FHIR como un modulo Go independiente. Esta arquitectura te permite importar una sola version, o importar multiples versiones en el mismo proyecto sin conflictos.

## Estructura de Modulos

Cada version de FHIR reside en su propio directorio en la raiz del repositorio con su propio `go.mod`:

```
models/
  go.work             # Archivo de workspace de Go para desarrollo local
  r4/
    go.mod            # module github.com/gofhir/models/r4
    helpers/          # Constantes auxiliares escritas manualmente
  r4b/
    go.mod            # module github.com/gofhir/models/r4b
  r5/
    go.mod            # module github.com/gofhir/models/r5
```

El archivo `go.work` en la raiz conecta estos modulos para el desarrollo local:

```go
go 1.23

use (
    ./cmd/generator
    ./r4
    ./r4b
    ./r5
)
```

## Importando una Sola Version

La mayoria de los proyectos solo necesitan una version de FHIR. Importa el modulo correspondiente directamente:

```go
import "github.com/gofhir/models/r4"

patient := r4.NewPatient(
    r4.WithPatientId("example"),
)
```

Instalar con:

```bash
go get github.com/gofhir/models/r4
```

## Importando Multiples Versiones

Cuando necesitas trabajar con multiples versiones de FHIR (por ejemplo, un convertidor o un servidor que soporta multiples versiones), importalas con alias de paquetes:

```go
import (
    r4 "github.com/gofhir/models/r4"
    r5 "github.com/gofhir/models/r5"
)

func convertPatient(r4Patient *r4.Patient) *r5.Patient {
    r5Patient := r5.NewPatientBuilder().
        SetId(*r4Patient.Id).
        SetBirthDate(*r4Patient.BirthDate).
        Build()

    // Copiar nombres
    for _, name := range r4Patient.Name {
        r5Patient.Name = append(r5Patient.Name, r5.HumanName{
            Family: name.Family,
            Given:  name.Given,
        })
    }

    return r5Patient
}
```

Instalar ambos modulos:

```bash
go get github.com/gofhir/models/r4
go get github.com/gofhir/models/r5
```

{{< callout type="info" >}}
Como cada version de FHIR es un modulo Go completamente separado, no hay conflictos de tipos. Un `r4.Patient` y un `r5.Patient` son tipos distintos, aunque comparten el mismo nombre de struct dentro de sus paquetes.
{{< /callout >}}

## Patrones de Alias de Paquetes

Al importar multiples versiones, el nombre del paquete por defecto coincide con el nombre del directorio (`r4`, `r4b`, `r5`), por lo que los alias solo son necesarios si deseas nombres personalizados:

```go
import (
    // Nombres por defecto -- no se necesita alias
    "github.com/gofhir/models/r4"
    "github.com/gofhir/models/r4b"
    "github.com/gofhir/models/r5"
)

// O usar alias descriptivos
import (
    fhirR4  "github.com/gofhir/models/r4"
    fhirR4B "github.com/gofhir/models/r4b"
    fhirR5  "github.com/gofhir/models/r5"
)
```

## Versionado Independiente

Cada modulo se versiona de forma independiente usando [release-please](https://github.com/googleapis/release-please). Esto significa:

- Una correccion de errores en el modulo R4 no fuerza un incremento de version en R4B o R5
- Los cambios incompatibles en un paquete de version no afectan a los demas
- Cada modulo tiene su propio `CHANGELOG.md` que registra los cambios

Las etiquetas de version siguen el patron `<component>/v<version>`:

```
r4/v0.3.0
r4b/v0.2.1
r5/v0.1.0
```

El archivo `release-please-config.json` en la raiz del repositorio configura esta estrategia de publicacion multi-paquete:

```json
{
  "packages": {
    "r4": {
      "release-type": "go",
      "component": "r4",
      "changelog-path": "CHANGELOG.md"
    },
    "r4b": {
      "release-type": "go",
      "component": "r4b",
      "changelog-path": "CHANGELOG.md"
    },
    "r5": {
      "release-type": "go",
      "component": "r5",
      "changelog-path": "CHANGELOG.md"
    }
  }
}
```

## Funcionalidades Especificas por Version

Aunque los tres paquetes comparten la misma estructura general, cada version de FHIR tiene diferencias en su conjunto de recursos y tipos de datos. Por ejemplo:

- **R4** incluye recursos como `MedicinalProduct` y `EffectEvidenceSynthesis` que fueron eliminados en R5
- **R4B** es en gran medida identico a R4 con actualizaciones incrementales a ciertos recursos
- **R5** introduce nuevos recursos y reestructura otros (por ejemplo, los recursos relacionados con medicamentos fueron reestructurados)

El codigo generado para cada version refleja con precision las StructureDefinitions oficiales para esa version de FHIR, por lo que siempre obtienes las definiciones de tipos correctas para la version que estas utilizando.

## Workspace de Go para Desarrollo

Si estas contribuyendo al proyecto `gofhir/models`, el archivo `go.work` permite que todos los modulos se desarrollen juntos sin publicar. El workspace asegura que los cambios locales en la infraestructura compartida (como el generador) sean inmediatamente visibles en todos los paquetes de version.

```bash
# Ejecutar todas las pruebas de R4
cd r4 && go test ./...

# Ejecutar todas las pruebas de R4B
cd r4b && go test ./...

# Ejecutar todas las pruebas de R5
cd r5 && go test ./...
```

El archivo `go.work` no se incluye en los modulos publicados y no afecta a los consumidores downstream.

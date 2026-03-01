---
title: "Generacion de Codigo"
linkTitle: "Generacion de Codigo"
description: "Como la herramienta cmd/generator lee las StructureDefinitions de FHIR y produce codigo fuente Go para recursos, builders, serializacion y metadatos."
weight: 3
---

Todo el codigo fuente Go en los paquetes `r4`, `r4b` y `r5` se genera automaticamente a partir de las StructureDefinitions oficiales de FHIR. La herramienta `cmd/generator` orquesta este proceso, leyendo archivos JSON de especificacion del directorio `specs/` y produciendo archivos fuente Go en el directorio del paquete destino.

## Ejecutando el Generador

El generador se invoca desde la raiz del repositorio usando `go run`:

```bash
# Generar codigo R4
go run cmd/generator/main.go r4

# Generar codigo R4B
go run cmd/generator/main.go r4b

# Generar codigo R5
go run cmd/generator/main.go r5
```

El generador acepta un unico argumento -- la version de FHIR a generar. Debe ser uno de `r4`, `r4b` o `r5`.

{{< callout type="info" >}}
No necesitas ejecutar el generador para usar la biblioteca. Todo el codigo generado esta committeado en el repositorio y publicado como modulos de Go. El generador solo es necesario cuando se actualiza a una nueva version de la especificacion FHIR o se modifican las plantillas de generacion.
{{< /callout >}}

## Que se Genera

Para cada version de FHIR, el generador produce los siguientes archivos:

| Salida | Descripcion |
|--------|-------------|
| `resource_*.go` | Un archivo por tipo de recurso conteniendo la struct, marshaling JSON/XML, builder y opciones funcionales |
| `datatypes.go` | Todos los tipos de datos FHIR (HumanName, Address, CodeableConcept, Quantity, etc.) |
| `codesystems.go` | Todas las enumeraciones de sistemas de codigos FHIR como tipos `string` de Go con constantes |
| `interfaces.go` | Las interfaces `Resource` y `DomainResource` |
| `registry.go` | El mapa de fabrica de recursos y funciones de deserializacion polimorfica |
| `fhirpath_model.go` | El singleton `FHIRPathModelData` con todos los mapas de metadatos de tipos |
| `summary.go` | El mapa `SummaryFields` con las listas de campos isSummary para cada recurso |
| `marshal.go` | Funciones personalizadas de marshaling JSON (`Marshal`, `MarshalIndent`) |
| `xml_helpers.go` | Funciones auxiliares de serializacion XML y constantes de namespace |

Cada archivo de recurso (por ejemplo, `resource_patient.go`) contiene:

1. **La struct del recurso** con tags JSON y XML
2. **Implementaciones de metodos de interfaz** (`GetResourceType`, `GetId`, `SetId`, `GetMeta`, `SetMeta` y metodos de `DomainResource` cuando corresponda)
3. **`MarshalJSON`/`UnmarshalJSON`** para la inyeccion automatica de `resourceType` y el manejo de recursos contenidos
4. **`MarshalXML`/`UnmarshalXML`** para la codificacion XML conforme a FHIR con manejo adecuado de namespace y atributos primitivos
5. **Structs de backbone elements** (por ejemplo, `PatientContact`, `PatientCommunication`) con sus propios metodos de marshaling
6. **Un builder fluido** (`PatientBuilder` con `NewPatientBuilder`, `Set*`, `Add*`, `Build`)
7. **Opciones funcionales** (tipo `PatientOption`, `NewPatient`, funciones `WithPatient*`)

## Pipeline Interno

El generador sigue un pipeline de tres etapas:

### 1. Parser

El parser lee los archivos JSON de StructureDefinition de FHIR del directorio `specs/<version>/`. Estos archivos son las definiciones oficiales legibles por maquina publicadas por HL7 para cada version de FHIR. El parser extrae:

- Definiciones de recursos y sus jerarquias de elementos
- Definiciones de tipos de datos (primitivos y complejos)
- Conjuntos de valores de sistemas de codigos y sus codigos permitidos
- Metadatos de elementos: cardinalidad, tipos, isSummary, objetivos de referencia, variantes de choice types

### 2. Analizador

El analizador procesa los datos parseados en un modelo interno adecuado para la generacion de codigo. Las transformaciones clave incluyen:

- **Aplanar jerarquias de elementos** en definiciones de campos de struct compatibles con Go
- **Resolver choice types** (por ejemplo, `value[x]`) en campos separados por variante de tipo
- **Construir la jerarquia de tipos** para verificaciones de satisfaccion de interfaces
- **Calcular los limites de backbone elements** para determinar que elementos anidados necesitan su propia struct
- **Recopilar metadatos de FHIRPath** en los seis mapas que poblan `FHIRPathModelData`
- **Extraer flags de resumen** para construir el mapa `SummaryFields`

### 3. Generador

El generador toma el modelo analizado y renderiza archivos fuente Go usando el paquete `text/template` de Go. Las plantillas manejan:

- Generacion de campos de struct con tipos Go correctos, tags JSON y tags XML
- Generacion de metodos de interfaz basandose en si un recurso es un Resource base o DomainResource
- Generacion de builders y opciones funcionales siguiendo patrones de nomenclatura consistentes
- Generacion de marshal/unmarshal XML con reglas de codificacion especificas de FHIR
- Generacion de constantes de sistemas de codigos con nombres compatibles con Go

Despues de la generacion, los archivos de salida se formatean con `gofmt` para asegurar un estilo consistente.

## Configuracion

El generador usa una struct `Config` para determinar su comportamiento:

```go
type Config struct {
    SpecsDir    string // Ruta al directorio specs/<version>/
    OutputDir   string // Ruta al directorio del paquete de salida (por ejemplo, ./r4)
    PackageName string // Nombre del paquete Go (por ejemplo, "r4")
    Version     string // Identificador de version FHIR (por ejemplo, "r4")
}
```

Cuando se invoca via `cmd/generator/main.go`, estas rutas se resuelven relativas a la raiz del repositorio.

## Agregar una Nueva Version de FHIR

Para agregar soporte para una nueva version de FHIR:

1. Descargar los archivos JSON de StructureDefinition de la especificacion FHIR de HL7
2. Colocarlos en `specs/<version>/`
3. Crear el directorio de salida (por ejemplo, `r6/`)
4. Inicializar un modulo Go en el directorio de salida (`go mod init github.com/gofhir/models/r6`)
5. Ejecutar el generador: `go run cmd/generator/main.go r6`
6. Agregar el nuevo modulo a `go.work`
7. Agregar una entrada de release-please en `release-please-config.json`

## Modificar el Codigo Generado

{{< callout type="warning" >}}
Nunca edites los archivos generados directamente. Todos los archivos en `r4/`, `r4b/` y `r5/` (excepto `helpers/`) comienzan con el comentario `// Code generated by gofhir. DO NOT EDIT.` y seran sobrescritos cuando se ejecute el generador.
{{< /callout >}}

Para cambiar la salida generada, modifica las plantillas y la logica de generacion en el paquete `internal/codegen/generator`, luego vuelve a ejecutar el generador para todas las versiones afectadas.

El subdirectorio `helpers/` (por ejemplo, `r4/helpers/`) esta escrito manualmente y no se ve afectado por el generador. Proporciona constantes y funciones de conveniencia que se construyen sobre los tipos generados.

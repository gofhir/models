---
title: "Contribuir"
linkTitle: "Contribuir"
description: "Como contribuir al proyecto gofhir/models: configuracion, flujo de trabajo de desarrollo, pruebas, generacion de codigo y pautas para pull requests."
weight: 99
---

Gracias por tu interes en contribuir a `gofhir/models`. Esta pagina cubre la configuracion del entorno de desarrollo, el flujo de trabajo de pruebas y las pautas para enviar cambios.

## Requisitos Previos

- **Go 1.23 o posterior** -- El proyecto usa Go workspaces y generics.
- **Git** -- Para clonar el repositorio y gestionar ramas.

## Primeros Pasos

### Clonar el Repositorio

```bash
git clone https://github.com/gofhir/models.git
cd models
```

### Entender la Estructura del Proyecto

```
models/
  cmd/generator/     # Herramienta generadora de codigo
  internal/          # Infraestructura compartida de generacion
  specs/             # Archivos JSON de StructureDefinition de FHIR
  r4/                # Modulo R4 generado (github.com/gofhir/models/r4)
    helpers/          # Constantes auxiliares escritas manualmente (no generadas)
  r4b/               # Modulo R4B generado (github.com/gofhir/models/r4b)
  r5/                # Modulo R5 generado (github.com/gofhir/models/r5)
  docs/              # Sitio de documentacion (Hugo)
  go.work            # Archivo de workspace de Go
```

### Configurar el Workspace

El repositorio usa un workspace de Go (`go.work`) para vincular todos los modulos para el desarrollo local. No se requiere configuracion adicional -- Go usara automaticamente el workspace cuando compiles o ejecutes pruebas desde dentro del repositorio.

```bash
# Verificar que el workspace es funcional
go work sync
```

## Ejecutando Pruebas

Cada modulo de version FHIR tiene su propia suite de pruebas. Ejecuta las pruebas desde el directorio del modulo:

```bash
# Probar R4
cd r4 && go test ./...

# Probar R4B
cd r4b && go test ./...

# Probar R5
cd r5 && go test ./...
```

O ejecuta todas las pruebas desde la raiz del repositorio:

```bash
cd r4 && go test ./... && cd ../r4b && go test ./... && cd ../r5 && go test ./...
```

Las suites de pruebas cubren:

- Construccion de structs de recursos y acceso a campos
- Marshaling y unmarshaling JSON (incluyendo inyeccion de `resourceType` y recursos contenidos)
- Marshaling y unmarshaling XML (manejo de namespaces, codificacion de primitivos, backbone elements)
- Correccion del builder fluido
- Funciones del registro (NewResource, UnmarshalResource, GetResourceType, etc.)
- Precision de los metadatos del modelo FHIRPath
- Validacion de constantes de sistemas de codigos
- Preservacion de precision del tipo Decimal

## Obteniendo las especificaciones FHIR

El generador lee tres bundles JSON por versión desde `specs/`. Suman unos 143 MB y
**no** están versionados, así que un clon nuevo debe descargarlos primero:

```bash
# Descargar y verificar todas las versiones
./scripts/fetch-specs.sh

# O solo la que necesites
./scripts/fetch-specs.sh r4

# Verificar lo que ya está en disco, sin descargar
./scripts/fetch-specs.sh --verify
```

Cada archivo se verifica contra el SHA-256 fijado en `specs.lock`, que también
registra la URL de origen y la versión FHIR. Una discrepancia es un fallo duro: si
la especificación publicada cambia, actualiza `specs.lock` de forma deliberada en
lugar de dejar que el código generado se desvíe.

Sin estos archivos el generador se detiene con
`failed to read required value sets`. Es intencional: antes continuaba y emitía en
silencio `*string` en lugar de cada tipo generado de code system.

## Suite de conformidad

`conformance/` hace round-trip de todos los ejemplos oficiales de FHIR a través de
la librería —unos 12.400 archivos entre las tres versiones— y compara el resultado
con las listas de fallos conocidos en `conformance/testdata/known_failures/`.

```bash
./scripts/fetch-examples.sh        # descargar los corpus (~200 MB por versión)
cd conformance && go test ./...
```

La suite se salta a sí misma cuando el corpus no está presente, así que nunca
fuerza la descarga en un `go test ./...` normal.

Las listas funcionan como un trinquete, no como una foto: la suite falla tanto
cuando un archivo que pasaba empieza a fallar (una regresión) como cuando un
archivo listado empieza a pasar (progreso que hay que registrar). Tras un arreglo,
regenéralas y lee el diff: cada línea eliminada es un bug corregido:

```bash
cd conformance && go test . -update-known
```

## Regenerando Modelos

Si modificas el generador de código o actualizas los archivos de especificación FHIR, regenera el código.
`cmd/generator` es un módulo Go propio, así que ejecútalo desde ese directorio:

```bash
cd cmd/generator

go run . r4    # Regenerar código R4
go run . r4b   # Regenerar código R4B
go run . r5    # Regenerar código R5
```

CI regenera las tres versiones en cada pull request y falla si el resultado difiere
de lo versionado, así que incluye la salida regenerada en tu PR.

Despues de la regeneracion, ejecuta la suite completa de pruebas para verificar la correccion:

```bash
cd r4 && go test ./... && cd ../r4b && go test ./... && cd ../r5 && go test ./...
```

{{< callout type="info" >}}
Nunca edites los archivos generados en `r4/`, `r4b/` o `r5/` directamente (excepto el subdirectorio `helpers/`). Todos los archivos generados comienzan con `// Code generated by gofhir. DO NOT EDIT.` y seran sobrescritos cuando se ejecute el generador. Modifica el codigo del generador en `internal/codegen/generator` en su lugar.
{{< /callout >}}

## Tipos de Contribuciones

### Correcciones de Errores

1. Abre un issue describiendo el error con un caso de reproduccion minimo.
2. Crea una rama: `git checkout -b fix/description`.
3. Agrega una prueba que falle y demuestre el error.
4. Corrige el problema -- si involucra codigo generado, corrige el generador y regenera.
5. Verifica que todas las pruebas pasen.
6. Envia un pull request.

### Nuevas Funcionalidades

1. Abre un issue para discutir la funcionalidad antes de comenzar a trabajar.
2. Crea una rama: `git checkout -b feat/description`.
3. Implementa la funcionalidad con pruebas.
4. Si la funcionalidad involucra codigo generado, actualiza las plantillas del generador y regenera todas las versiones.
5. Envia un pull request.

### Documentacion

1. La documentacion reside en `docs/content/en/docs/`.
2. El sitio usa Hugo con el tema Hextra.
3. Para previsualizar localmente:

```bash
cd docs
hugo server
```

### Constantes Auxiliares

El directorio `r4/helpers/` contiene constantes de conveniencia escritas manualmente. Las contribuciones de nuevas constantes auxiliares son bienvenidas -- por ejemplo:

- Codigos LOINC adicionales para observaciones clinicas comunes
- Funciones auxiliares de unidades UCUM adicionales
- Valores `CodeableConcept` preconstruidos para conjuntos de valores comunes

## Convencion de Mensajes de Commit

Este proyecto usa [Conventional Commits](https://www.conventionalcommits.org/) para los mensajes de commit. El bot release-please los usa para generar changelogs y determinar incrementos de version.

| Prefijo | Proposito | Incremento de Version |
|---------|-----------|:---------------------:|
| `feat:` | Nueva funcionalidad | Minor |
| `fix:` | Correccion de error | Patch |
| `refactor:` | Reestructuracion de codigo | -- |
| `docs:` | Solo documentacion | -- |
| `test:` | Adiciones/cambios en pruebas | -- |
| `chore:` | Tareas de mantenimiento | -- |

Ejemplos:

```
feat: add Encounter resource builder validation
fix: correct XML namespace for contained resources
refactor: consolidate generated files and add XML deserialization
docs: add examples for Bundle construction
```

## Pautas para Pull Requests

1. **Un tema por PR** -- Manten los pull requests enfocados en un solo cambio.
2. **Incluye pruebas** -- Todos los cambios de codigo deben tener pruebas correspondientes.
3. **Ejecuta la suite completa** -- Asegurate de que todas las pruebas pasen en las tres versiones de FHIR antes de enviar.
4. **Regenera si es necesario** -- Si tu cambio afecta al generador, incluye la salida regenerada en el PR.
5. **Sigue el estilo existente** -- Respeta el estilo de codificacion y los patrones usados en el codigo existente.
6. **Actualiza la documentacion** -- Si tu cambio agrega o modifica la API publica, actualiza las paginas de documentacion relevantes.

## Codigo de Conducta

Por favor se respetuoso y constructivo en todas las interacciones. Estamos construyendo herramientas para la interoperabilidad en salud, y una comunidad acogedora nos ayuda a hacerlo mejor.

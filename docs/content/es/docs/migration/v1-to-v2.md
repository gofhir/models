---
title: "De v1 a v2"
linkTitle: "De v1 a v2"
description: "Qué cambia en la v2, qué hacer al respecto y qué está todavía por decidir."
weight: 1
---

{{< callout type="info" >}}
**La v2 aún no está publicada.** Esta página existe para que los avisos de deprecación de la v1.6.0 apunten a algo real. La sección sobre opciones funcionales es definitiva y puedes actuar sobre ella hoy. El resto es una previsión, marcada como tal, y se completará cuando se corte la v2.
{{< /callout >}}

## Empieza por aquí: activa los avisos

La v1.6.0 marca con la convención `Deprecated:` de Go toda la API que la v2 elimina. No necesitas leer esta página para encontrar tus llamadas — tus herramientas te las listan:

```shell
go vet ./...              # no reporta deprecaciones
staticcheck ./...         # sí: SA1019
golangci-lint run ./...   # sí, a través de staticcheck
```

Los editores que usan `gopls` muestran esos mismos símbolos tachados.

Cada aviso nombra su propio reemplazo, así que la mayor parte de la migración es mecánica:

```text
main.go:11:3: r4.WithPatientActive is deprecated: use PatientBuilder.SetActive
              instead; removed in v2. (SA1019)
```

Si no ves ningún aviso, no estás usando nada que la v2 elimine.

## Las opciones funcionales pasan a ser builders

**Estado: definitivo.** Es el cambio más grande por volumen —11.952 funciones entre los tres módulos— y el más mecánico.

Cada opción `With<Recurso><Campo>` tiene un método de builder con el mismo tipo de parámetro y el mismo efecto. La correspondencia se verificó par por par, no se dio por supuesta: un test parsea el código generado y compara el tipo del parámetro y la asignación de las 11.952.

| v1 | v2 |
|---|---|
| `NewPatient(opts...)` | `NewPatientBuilder()` … `.Build()` |
| `WithPatientActive(true)` | `.SetActive(true)` |
| `WithPatientIdentifier(id)` | `.AddIdentifier(id)` |
| `PatientOption` (tipo) | eliminado — el builder no necesita un tipo de opción |

La regla para el nombre del método: **`Add` para los campos repetibles, `Set` para todo lo demás.** Refleja lo que la opción ya hacía — los campos repetibles añadían al final, los simples asignaban.

Antes:

```go
p := r4.NewPatient(
    r4.WithPatientId("p1"),
    r4.WithPatientActive(true),
    r4.WithPatientIdentifier(r4.Identifier{System: r4.Ptr("urn:mrn")}),
    r4.WithPatientName(r4.HumanName{Family: r4.Ptr("Smith")}),
)
```

Después:

```go
p := r4.NewPatientBuilder().
    SetId("p1").
    SetActive(true).
    AddIdentifier(r4.Identifier{System: r4.Ptr("urn:mrn")}).
    AddName(r4.HumanName{Family: r4.Ptr("Smith")}).
    Build()
```

Los literales de struct no se ven afectados y siguen siendo la forma más corta para un recurso que ya tienes completo:

```go
p := &r4.Patient{ResourceType: "Patient", Id: r4.Ptr("p1"), Active: r4.Ptr(true)}
```

### Si pasas opciones de un lado a otro

El valor de una opción es un `func(*Patient)`, y el código que las almacena o las pasa necesita reestructurarse, no renombrarse:

```go
// v1
func configure() []r4.PatientOption {
    return []r4.PatientOption{r4.WithPatientActive(true)}
}

// v2 — recibe el builder en su lugar
func configure(b *r4.PatientBuilder) *r4.PatientBuilder {
    return b.SetActive(true)
}
```

Es la única parte de la migración de opciones que no es un renombrado. Si tu código no menciona `PatientOption` en ningún sitio, no te afecta.

### Encontrar tus llamadas

No hay codemod, y es deliberado, no algo a medias: la reescritura es estructural, no un renombrado. `NewPatient(a, b, c)` tiene que convertirse en una cadena de métodos terminada en `Build()`, y una expresión regular que renombrara `WithPatientActive` a `SetActive` en el sitio dejaría código que no compila. La parte mecánica es elegir el nombre nuevo, y el mensaje de deprecación ya lo hizo por cada llamada.

Para obtener la lista completa, cada sitio emparejado con su reemplazo exacto:

```shell
staticcheck ./... 2>&1 \
  | sed -nE 's/^(.+): ([A-Za-z0-9_.]+) is deprecated: use ([A-Za-z0-9_.]+) instead.*/\1\n      \2  ->  \3/p'
```

```text
main.go:23:3
      r4.WithPatientId  ->  PatientBuilder.SetId
main.go:24:3
      r4.WithPatientActive  ->  PatientBuilder.SetActive
main.go:25:3
      r4.WithPatientIdentifier  ->  PatientBuilder.AddIdentifier
```

O, para dimensionar el trabajo antes de empezarlo:

```shell
staticcheck ./... 2>&1 | grep -c SA1019
```

## Todo lo que sigue es una previsión

**Estado: no definitivo.** Estos cambios están planificados para la v2 y se describen aquí para que puedas valorar el impacto con antelación. Los detalles, y en algunos casos si llegan a publicarse, pueden cambiar. Ninguno se puede señalar con un aviso de deprecación, porque los campos conservan su nombre y solo cambian de tipo — que es precisamente por lo que vale leerlos antes.

### `ResourceType` deja de ser un `string`

Para dejar de pagar un `MarshalJSON` por recurso, se prevé que `ResourceType` pase a ser un tipo marcador de tamaño cero. Leerlo como string, o asignarlo en un literal de struct, dejaría de compilar.

`GetResourceType()` sigue funcionando y devuelve un `string`. **Si usas el método en lugar del campo, este cambio no te cuesta nada** — esa es la migración, y funciona hoy en la v1.

Dos consecuencias que conviene revisar ya:

- **Cambia el orden de las claves JSON.** Si comparas payloads byte a byte, o los hasheas, esa comparación se mueve.
- `json.Marshal(patient)` deja de ser equivalente a `r4.Marshal(patient)`.

### Los primitivos repetibles pasan a ser slices de punteros

`[]string` pasa a `[]*string`, para que un `null` en medio de un array FHIR sobreviva:

```json
"given": ["A", null, "C"]
```

Hoy ese `null` intermedio se lee como `""` —un valor vacío que clínicamente es distinto de ausente—. Al volver a escribir ese array se re-emiten datos inventados. El slice de punteros es lo que hace representable la ausencia.

Se prevé que los helpers de conversión (`PtrSlice` en un sentido, `Vals` en el otro) lleguen con el cambio. No están en la v1.6.0 — hoy los helpers disponibles son `Ptr`, `Val` y `First`, y solo cubren escalares.

### Los campos complejos requeridos pasan a ser punteros

Unos 99 campos por versión son structs sin puntero, así que se serializan incluso vacíos:

```go
r4.Observation{Id: r4.Ptr("o1")}
// {"resourceType":"Observation","id":"o1","code":{}}   ← code:{} es FHIR inválido
```

Convertirlos en punteros corrige la salida y cambia el tipo de esos campos.

### `Contained` pasa a ser un tipo de slice con nombre

`[]Resource` pasa a `ContainedList`. La asignación desde `[]Resource`, el `range`, el `append` y pasarlo a un `func([]Resource)` siguen funcionando; lo que cambia es que `%T` imprime `r4.ContainedList`.

### Cambian los nombres de los tipos de sistemas de códigos

Se prevé derivar los nombres de tipos y constantes de enums de la extensión FHIR `bindingName`, lo que renombra unos 657 tipos y 3.613 constantes. La publicación irá acompañada de una tabla completa de correspondencias viejo→nuevo. Es el cambio con más probabilidades de separarse o escalonarse, porque el volumen es grande y el beneficio es coherencia de nombres, no corrección.

## La ruta de importación

Los módulos v2 llevan el sufijo `/v2` que Go exige:

```shell
go get github.com/gofhir/models/r4/v2
```

```go
import "github.com/gofhir/models/r4/v2"

var p r4.Patient   // el paquete sigue llamándose r4
```

No hay ningún directorio `v2` en el repositorio — el sufijo pertenece a la ruta del módulo, no a la estructura de carpetas. Como la ruta cambió, **la v1 y la v2 pueden convivir en la misma compilación**, y eso es lo que hace posible una migración incremental: mover un paquete a la vez mientras alguna dependencia siga arrastrando la v1.

## Quedarse en la v1

La v1 sigue recibiendo los avisos de deprecación y cualquier arreglo que no exija un cambio rompiente. Es un sitio válido donde quedarse mientras migras; los avisos son informativos y no cambian ningún comportamiento.

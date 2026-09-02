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

**Estado: hecho.** Es el cambio más grande por volumen —11.952 funciones entre los tres módulos— y el más mecánico. Ya no están en la v2, que es lo que prometían los mensajes de deprecación de la v1.6.0.

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

## Cambios de tipo

Ninguno se puede señalar con un aviso de deprecación, porque los campos conservan su nombre y solo cambian de tipo — que es precisamente por lo que vale leerlos antes. Cada uno indica si ya aterrizó o sigue siendo una previsión.

### `ResourceType` deja de ser un `string`

**Estado: hecho, y el mayor cambio individual para la mayoría del código.** `ResourceType` es ahora un tipo marcador de tamaño cero, así que asignarlo no compila:

```go
p := &r4.Patient{ResourceType: "Patient", Id: r4.Ptr("p1")}  // v1
p := &r4.Patient{Id: r4.Ptr("p1")}                           // v2 — misma salida
```

Borrar la línea es toda la migración. La salida es idéntica byte a byte, con `resourceType` incluido y sigue siendo la primera clave.

Para leerlo, usa `GetResourceType()`, que también funciona en la v1 — así que el código que ya llama al método no necesita ningún cambio:

```go
p.GetResourceType()   // "Patient", en ambas versiones
```

Dos correcciones a lo que aquí se preveía, ahora que está medido:

- **El orden de las claves JSON no cambia.** Se esperaba que lo hiciera; no lo hace.
- **`json.Marshal` ya no era equivalente a `r4.Marshal`,** y sigue sin serlo. El método eliminado llamaba a `SetEscapeHTML(false)`, pero eso nunca tuvo efecto: `encoding/json` recompacta el resultado de `MarshalJSON` aplicando la configuración del encoder externo. `Marshal` es lo que preserva `<` en la narrativa XHTML — eso no cambia, y es la razón de que la documentación te diga que lo prefieras.

Lo que ganas: serializar un Bundle de 50 entradas es **3,4× más rápido y reserva 4,7× menos memoria** (145850 → 43200 ns/op, 71900 → 15216 B/op), porque el método eliminado construía un segundo `bytes.Buffer` y un `json.Encoder` por cada recurso.

Y se corrige un bug silencioso. Un método promovido desde un struct incrustado satisface `json.Marshaler` para el tipo exterior, así que esto se serializaba como un Patient a secas, descartando `tenant` y `score` sin ningún error:

```go
type WithMetadata struct {
    r4.Patient
    Tenant string `json:"tenant"`
    Score  int    `json:"score"`
}
```

### Los primitivos repetibles pasan a ser slices de punteros

**Estado: hecho.** `[]string` pasa a `[]*string`, para que un `null` en medio de un array FHIR sobreviva:

```json
"given": ["A", null, "C"]
```

Hoy ese `null` intermedio se lee como `""` —un valor vacío que clínicamente es distinto de ausente—. Al volver a escribir ese array se re-emiten datos inventados. El slice de punteros es lo que hace representable la ausencia.

`PtrSlice` convierte una lista de valores en el slice de punteros, y `Vals` hace el camino inverso:

```go
Given: r4.PtrSlice("John", "Q")   // []*string
nombres := r4.Vals(hn.Given)      // []string, las entradas nil pasan a ""
```

### El `integer64` de R5 viaja como string

**Estado: hecho.** El `integer64` de R5 se generaba como `int64` y se serializaba como un número JSON a secas. La especificación exige un string, porque los números JSON solo llevan 53 bits de precisión entera e `integer64` son 64 — los valores por encima de 2^53 se redondeaban en silencio.

El tipo del campo pasa a `Integer64`, que se serializa como string y acepta ambas formas al leer, así que los documentos escritos por versiones anteriores siguen leyéndose:

```go
var att r5.Attachment
json.Unmarshal([]byte(`{"size":"9007199254740993"}`), &att)

att.Size.Int64()   // 9007199254740993 — exacto, por encima de 2^53
att.Size.String()  // "9007199254740993"
```

La lectura sigue aceptando la forma numérica antigua, así que los documentos escritos por versiones anteriores no se rechazan.

### Los campos complejos requeridos pasan a ser punteros

**Estado: hecho.** 245 campos en r4, 250 en r4b y 315 en r5 eran structs por valor, así que se serializaban aunque no se hubiera puesto nada:

```go
r4.Observation{Id: r4.Ptr("o1")}
// v1: {"resourceType":"Observation","id":"o1","code":{}}   ← inválido según ele-1
// v2: {"resourceType":"Observation","id":"o1"}
```

El razonamiento para usar campos por valor era que el tipo expresara la obligatoriedad. No podía: Go no tiene forma de forzar que un campo se rellene, así que lo único que conseguía el tipo por valor era emitir un objeto vacío en lugar de nada. La obligatoriedad es tarea de un validador FHIR; lo que el tipo sí tiene que poder decir es «ausente».

`Extension.url` siguió el mismo camino, así que una extensión sin rellenar es ahora `{}` en vez de `{"url":""}`.

**Para migrar: añade `&` al literal de struct.**

```go
Code: r4.CodeableConcept{Coding: ...}    // v1
Code: &r4.CodeableConcept{Coding: ...}   // v2
```

La lectura no se ve afectada —Go desreferencia automáticamente, así que `obs.Code.Coding` sigue compilando—. Lo que cambia es que ahora puede ser nil, así que protégelo donde el valor pudiera no estar puesto.

**Las llamadas al builder no cambian en absoluto.** `SetCode` sigue recibiendo un `CodeableConcept` por valor y tomando su dirección internamente:

```go
r4.NewObservationBuilder().SetCode(r4.CodeableConcept{...}).Build()   // sin cambios
```

### Los helpers de r4 pasan a ser funciones

**Estado: hecho.** `helpers.BodyWeight` y las otras 58 constantes LOINC y de categoría eran variables de paquete; ahora son funciones que devuelven un `*r4.CodeableConcept` nuevo, y los 33 helpers UCUM `Quantity*` devuelven `*r4.Quantity`.

```go
Code: helpers.BodyWeight,      // v1
Code: helpers.BodyWeight(),    // v2 — y como el campo ya es puntero, encaja directo
```

Con un builder, `Set*` sigue recibiendo un valor, así que desreferencia:

```go
SetCode(*helpers.BodyWeight()).
AddCategory(*helpers.ObservationCategoryVitalSigns()).
```

Esto es una corrección, no una limpieza. Una variable entregaba el mismo valor a todos los usos y, como `CodeableConcept` lleva un slice `Coding`, incluso copiando el struct se compartía el array subyacente: ajustar un `display` en un recurso lo cambiaba en todos los demás construidos con ese helper, y en el helper mismo. Ya pasaba en la v1; convertir los campos requeridos en punteros lo amplió de los slices al struct entero.

### `Contained` pasa a ser un tipo de slice con nombre

**Estado: hecho, y casi con seguridad no afecta a tu código.** `[]Resource` pasa a `ContainedList`, un slice con nombre y el mismo tipo subyacente:

```go
p.Contained = []r4.Resource{org}        // sigue compilando
p.Contained = append(p.Contained, x)    // sigue compilando
for _, c := range p.Contained { }       // sigue compilando
recibeSlice(p.Contained)                // func([]Resource) — sigue compilando
p.GetContained()                        // sigue devolviendo []Resource
```

**La única diferencia observable es `%T`**, que ahora imprime `r4.ContainedList`. Si formateas el tipo del campo en logs o lo comparas, esa cadena cambia.

El tipo con nombre es lo que lleva el `UnmarshalJSON` que despacha cada elemento. Antes cada recurso generaba el suyo, y por eso esto elimina **437 métodos y 16.130 líneas** de código generado. Decodificar un Patient con recursos contenidos es 1,35× más rápido, y un Bundle de 50 entradas 1,16× — real, pero menos que el 1,51× / 1,26× que proyectaba el plan.

Los mensajes de error conservan su índice, incluso anidados:

```
failed to unmarshal resource: failed to unmarshal Patient: failed to unmarshal contained[0]: unknown resource type: Nope
```

### Cambian los nombres de los tipos de sistemas de códigos

**Estado: previsto, y el menos seguro de todos.** Se prevé derivar los nombres de tipos y constantes de enums de la extensión FHIR `bindingName`, lo que renombra unos 657 tipos y 3.613 constantes. La publicación irá acompañada de una tabla completa de correspondencias viejo→nuevo. Es el cambio con más probabilidades de separarse o escalonarse, porque el volumen es grande y el beneficio es coherencia de nombres, no corrección.

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

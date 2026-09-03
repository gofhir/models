# Plan de remediación — gofhir/models

> Segunda revisión, tras review adversarial.
> Derivado de siete auditorías sobre `main` (4d39e4d). Las referencias a archivo y línea corresponden a ese commit.
> Cada defecto y cada solución citada aquí fue **ejecutada**, no inferida.

**Ocho fases · cuatro releases · 21–28 días-persona** (sin fase 8).
Lo explotable se cierra en las fases 0–2: **3 días**.

---

## Índice

- [Qué cambió respecto de la primera versión](#qué-cambió-respecto-de-la-primera-versión)
- [Fase 0 — Harness y control de release](#fase-0--harness-y-control-de-release)
- [Fase 1 — Corpus de conformidad como inventario](#fase-1--corpus-de-conformidad-como-inventario)
- [Fase 2 — Guarda de profundidad (cerrar el DoS)](#fase-2--guarda-de-profundidad-cerrar-el-dos)
- [Fase 3 — Higiene del generador y primera impresión](#fase-3--higiene-del-generador-y-primera-impresión)
- [Fase 4 — XML conforme, de una vez](#fase-4--xml-conforme-de-una-vez)
- [Fase 5 — Migración de la ruta de módulo](#fase-5--migración-de-la-ruta-de-módulo)
- [Fase 6 — Rendimiento y modelo](#fase-6--rendimiento-y-modelo)
- [Fase 7 — Cerrar el corpus](#fase-7--cerrar-el-corpus)
- [Fase 8 — Capa diferenciadora](#fase-8--capa-diferenciadora)
- [Releases y compatibilidad](#releases-y-compatibilidad)
- [Riesgos y capacidad](#riesgos-y-capacidad)

---

## Qué cambió respecto de la primera versión

El review atacó cinco soluciones prototipándolas. **Cuatro estaban mal.** Este delta explica por qué el orden y las estimaciones se movieron.

| Propuesta original | Veredicto medido | Sustituto |
|---|---|---|
| Propagar profundidad vía `unmarshalResourceDepth(data, depth)` | **Inviable** — `UnmarshalJSON([]byte) error` es firma fija; un contador de paquete produce data race y hasta 14.265 rechazos espurios de 19.200 | Guarda de escaneo de bytes O(n) en el dispatcher polimórfico |
| Borrar los `MarshalJSON` generados | **Rompe la salida** — es esa función la que fabrica el `resourceType`; sin ella se emite `"resourceType":""` | Marcador de tamaño cero: 2,6× medido, round-trip correcto |
| `ContainedList` da 1,6× | **Optimista** — 1,51× en Patient suelto pero solo **1,26× en Bundles**, el camino caliente real | Se mantiene con expectativa corregida; los 3 campos `Resource` singulares no admiten la técnica |
| `_ext` en backbones: «unas diez líneas» | **×5** — 6 sitios de plantilla × 2 archivos, más el `id` de backbone que va como atributo XML | Alcance real documentado; los choice types son una segunda clase de bug, no listada antes |
| Suite de conformidad al final | **Orden invertido** — cobertura real de 1,1 % en r4 y 0,8 % en r5, sin fixtures | Se parte: inventario de deuda al principio, gate al final |
| Commitear `go.work` | **Contraproducente** — degrada el job `verify-modules`, que dejaría de detectar un `go.sum` roto | `go get testify` en la raíz + `GOWORK=off` en CI |
| Sufijo `/v2` como nota al pie | **Faltaba fase** — release-please etiquetaría `r4/v2.0.0` sobre un `go.mod` sin sufijo y el proxy lo rechaza | Fase 5 propia, previa a cualquier commit rompiente |
| «3.765 funciones `With*`» | **Subconteo ×3** — ese número es solo r4; el total es **11.952** | Tarea con estimación propia en la fase 6 |

### Una refutación que se rechazó

El review sostuvo que la colisión de `MedicationStatusCodes` no se reproduce. La especificación lo zanja: **dos ValueSets con URLs distintas comparten el mismo `name`**.

```
.../ValueSet/medication-statement-status   NAME=Medication Status Codes
.../ValueSet/medication-status             NAME=Medication Status Codes
```

El review comparó contra `MedicationRequest` en lugar de `MedicationStatement`, y buscó tipos duplicados en `codesystems.go` — pero la ausencia de duplicados es el **síntoma**, no la refutación: el segundo ValueSet nunca llega a generarse. La tarea se mantiene.

### Dos premisas que cayeron

- El binario `generator` de 5 MB **no está commiteado** — está en `.gitignore` y es basura local.
- **No hay drift** entre los specs y el código generado: verificado regenerando las tres versiones byte a byte.

Ninguna de las dos necesita trabajo.

---

## Fase 0 — Harness y control de release

**1 día · sin release · prerrequisito de todo**

> **Objetivo:** que cualquiera pueda clonar, regenerar y comprobar el resultado — y que el generador no pueda volver a borrar la mitad de la API sin decirlo.

### 0.1 · Obtención de specs con hash fijado

`specs/` está gitignorado (~143 MB en las tres versiones) y nada en el repo registra su procedencia ni su versión FHIR. Añadir `scripts/fetch-specs.sh` más un `specs.lock` con URL y SHA-256 por archivo, verificado en la descarga.

Sin el hash fijado el check de drift no es determinista: depende de que HL7 sirva bytes estables.

### 0.2 · Arreglar el módulo raíz

El `go.mod` raíz son tres líneas sin un solo `require`, y los tests de `internal/` importan testify. Hoy solo compilan porque `go.work` lo toma prestado del `go.sum` de r4.

> **⚠ Cambio sobre la versión anterior.** El plan decía «commitear `go.work`». Es contraproducente: con el workspace presente, el job `verify-modules` resuelve dependencias vía workspace y deja de detectar un `go.sum` roto, que es justo su razón de existir. Además contradice una decisión explícita — `go.work` está en `.gitignore:53`.
>
> Correcto: `go get github.com/stretchr/testify` en la raíz y `GOWORK=off` en el job nuevo.

### 0.3 · `valuesets.json` ausente debe ser fatal

Verificado ejecutándolo: quitando ese archivo el generador imprime *«Successfully generated r4»* y sale con **código 0**, cambiando 122 archivos.

- Los 205 enums degradan a `*string`.
- `codesystems.go` sobrevive intacto, con sus tipos ya huérfanos.
- `go build ./...` compila limpio → el job `verify-modules` no ve nada.
- Solo fallan los tests, y por casualidad: dos de ellos usan un enum al azar.

> **⚠ Cambio sobre la versión anterior.** Estaba en la fase 3 como higiene del generador. Se sube aquí y se reclasifica como **control de release**: lo que previene no es un generador incómodo, es publicar un *patch* que borra 657 tipos exportados.

Incluir también la limpieza del directorio de salida, sin la cual el estado queda a medias y compilando. Hoy `Generate()` solo hace `MkdirAll`, así que un recurso eliminado del spec deja su `resource_*.go` huérfano para siempre, y los artefactos `.unformatted` que `template_loader.go:135` deja al fallar no están en `.gitignore`.

### 0.4 · Test golden y CI del generador

Bundle sintético mínimo con un backbone, un choice, un array de primitivos y **dos ValueSets que colisionen de nombre** → generar a directorio temporal → comparar contra fixture commiteado.

Los cuatro jobs actuales usan `matrix.module: [r4, r4b, r5]`, así que `internal/codegen/**` y `cmd/generator/` no se compilan, testean ni lintean nunca. Añadir job para el módulo raíz.

Actualizar `.golangci.yml` a formato v2 —el linter actual lo rechaza de plano— y borrar sus ocho `exclude-rules`, que apuntan a directorios inexistentes (`pkg/fhir`, `pkg/fhirpath`, `pkg/validator`) copiados de otro repositorio. `goimports.local-prefixes` también apunta a un módulo equivocado.

**Aceptación:** añadir al drift-check una aserción *positiva* (recuento mínimo de tipos enum), para que falle aunque nadie mire el diff.

---

## Fase 1 — Corpus de conformidad como inventario

**1–2 días · sin release · calibra la fase 2**

> **Objetivo:** construir el oráculo antes de reescribir lo que debe juzgar.

Cobertura real: **1,1 % en r4 y 0,8 % en r5**, sobre 709.000 líneas generadas, sin un solo directorio `testdata/` ni archivo golden. El plan original proponía reescribir la capa de serialización completa sobre esa base y construir la suite al final.

Descargar los ejemplos oficiales (JSON y XML) a `testdata/` y escribir un runner que solo haga *parse → serialize → re-parse → comparar semánticamente*, marcando los fallos actuales como `t.Skip` con su issue asociado.

No es un gate: es un **inventario de deuda**, y sale barato porque no exige arreglar nada todavía.

**Por qué va aquí y no al final:** además de dar señal a las fases 4 y 6, mide la profundidad de anidamiento real del corpus — el número que la fase 2 necesita para elegir su límite sin inventarlo.

### Resultado (ejecutado)

Implementada en el módulo `conformance/`, sobre **12.411 ejemplos oficiales**:

| | archivos | fallan | pasan |
|---|---|---|---|
| r4 JSON | 2.912 | 3 | **99,9 %** |
| r4b JSON | 3.022 | 11 | 99,6 % |
| r5 JSON | 2.824 | 61 | 97,8 % |
| r4 XML | 1.138 | **1.094** | 3,9 % |
| r4b XML | 1.156 | **1.075** | 7,0 % |
| r5 XML | 1.359 | **1.287** | 5,3 % |

**Un solo defecto causa el 96 % de los fallos de XML.** Todos los mismatches muestran el mismo primer byte de divergencia: la primera serialización emite `<rawInner><div…>` y al re-parsear el narrative desaparece. Arreglar `rawInner` (tarea 4.3) desbloquea ~3.400 archivos de una vez, lo que confirma la decisión de fusionar el writer de bytes en la fase XML en lugar de dejarlo para la v2.

Los tres fallos de r4 JSON son **exactamente** los bugs ya identificados, ahora con respaldo de ejemplos oficiales:

| Archivo | Diagnóstico | Tarea |
|---|---|---|
| `json-edge-cases.json` | `contact[0].name._given[0]: <nil> became {object}` | 6.4 (arrays de primitivos) |
| `structuredefinition-example-composition.json` | `differential.element[2].type[0]._profile: dropped` | 4.4 (`_ext` en backbones) |
| `package-min-ver.json` | sin `resourceType` | ruido del corpus, no es un recurso |

### Dos hallazgos nuevos

**`integer64` de R5 debe serializarse como string.** Ninguna de las siete auditorías lo vio. La spec exige que `integer64` viaje como string en JSON porque JSON solo garantiza 53 bits de precisión; la librería tipó `Attachment.size` como `int64` y falla al leer ejemplos oficiales:

```
failed to unmarshal Communication: json: cannot unmarshal string into
Go struct field Attachment.Alias.payload.contentAttachment.size of type int64
```

Es un cambio de tipo, así que **va a la v2** (fase 6). Afecta a los 20 fallos `parse` de r5.

**La profundidad real es 28, no 14.** El máximo está en `structuremap-questionnaire.json` de r4b — un Questionnaire con árboles de `item` muy anidados. Los bundles de definiciones llegan solo a 14, así que dimensionar el límite desde los specs habría sido optimista: el margen del límite de 64 propuesto es **2,3×**, no 4×. Y `Questionnaire.item`, `Parameters.parameter.part` y `GraphDefinition.link.target` no tienen tope en la especificación, así que documentos legítimos pueden ir más hondo. Dado que el riesgo es asimétrico, la tarea 2.1 debe elegir un límite **holgadamente** por encima de 28, no apenas por encima.

---

## Fase 2 — Guarda de profundidad (cerrar el DoS)

**½ día · v1.4.1 · aviso de seguridad**

> **Objetivo:** lo único de todo el plan que un atacante puede disparar hoy sin autenticarse.

> **⚠ La propuesta original era inviable.** `UnmarshalJSON([]byte) error` es la firma fija de `json.Unmarshaler`: cuando `json.Unmarshal` invoca el método de un tipo anidado, no hay forma de pasarle un contador. Las alternativas obvias se midieron y fallan:
>
> | variante | rechazos espurios | estado | `-race` |
> |---|---|---|---|
> | `var depth int` | **3.230 / 19.200** | contador corrupto de forma permanente | **data race confirmada** |
> | `atomic.Int64` | **14.265 / 19.200** | vuelve a 0 | limpio pero **inútil** |
>
> El atómico es peor: cuenta concurrencia global, no profundidad. Bajo carga rechaza ~74 % del tráfico legítimo.

### 2.1 · Escaneo de bytes O(n) en el dispatcher polimórfico ✅ HECHO

Implementado en `templates/registry.go.tmpl` como `resourceNestingDepth`, sin estado y seguro para concurrencia, dentro de `UnmarshalResource`. Todo anidamiento polimórfico pasa obligatoriamente por ahí, así que protege también el `json.Unmarshal(data, &p)` idiomático.

> **Corrección sobre lo planificado: cuenta recursos, no llaves JSON.**
> El plan proponía contar llaves con un tope de 64. Medido sobre el corpus, son **dos métricas distintas** que no deben confundirse:
>
> | métrica | máximo real | ¿causa el coste? |
> |---|---|---|
> | anidamiento estructural (llaves) | **28** (`Questionnaire.item`) | no — no re-lee nada |
> | recursos anidados | **3** (`bundle-response-simplesummary.json`) | **sí** — cada nivel re-parsea su subárbol |
>
> Un tope sobre llaves tendría que superar 28 para no rechazar Questionnaires legítimos, lo que dejaría el caso caro casi sin acotar. Contando recursos, el límite queda en **32** —más de diez veces el máximo real de 3— y un Questionnaire con 100 niveles de `item` pasa sin problema, porque su profundidad de recursos es 1.
>
> FHIR además prohíbe que un recurso contenido tenga sus propios contenidos, así que cualquier profundidad legítima más allá de un Bundle de Bundles ya está fuera de la especificación.

Medido:

```
ataque original (160 KB, depth 4000)   3.85 s / 670 MB  ->  0 ms / 2 MB   RECHAZADO
Bundle en Bundle en Patient (depth 3)  aceptado
Questionnaire con 100 niveles de item  aceptado
rechazo de 42 KB hostiles              14 us
```

Overhead en tráfico normal (Bundle de 50 Patients): **0,6 %** —dentro del ruido de medición— y **cero allocaciones** añadidas.

> **Dos bypasses que un code review adversarial encontró en la primera versión.**
> Ambos reproducidos antes de arreglarlos:
>
> | evasión | resultado | por qué |
> |---|---|---|
> | `contained` antes de `resourceType` | **aceptado, 7,55 s de CPU** | el escáner marcaba el objeto al *leer* la clave, y JSON no ordena los miembros |
> | clave escrita como `"resourceType"` | aceptado | `encoding/json` la desescapa; una comparación de bytes crudos no |
>
> El primero anulaba la guarda por completo: basta reordenar el payload. El arreglo calcula la profundidad **al cerrar** cada objeto —bottom-up, y por tanto independiente del orden— y decodifica la clave cuando lleva escapes.
>
> Los tests originales no los detectaban porque todos los generadores de payload emitían `resourceType` primero y sin escapar: la forma que produce un encoder correcto, no la que elige un atacante. `TestDepthGuardResistsEvasion` cubre ahora las tres combinaciones en las tres versiones.

`MaxResourceDepth` es configurable (0 la desactiva) y `ErrMaxResourceDepth` permite `errors.Is`. Los 12.411 ejemplos del corpus siguen pasando: la guarda no rechaza ningún documento publicado.

### 2.2 · Límite de profundidad en el decoder XML

La recursión mutua `UnmarshalXML → xmlDecodeContainedResource → xmlDecodeInlineResource → UnmarshalXML` consume 5 frames y ~1,4 KB por nivel. A ~200.000 niveles (8,2 MB de entrada) produce `fatal error: stack overflow`, que **`recover()` no atrapa**: cae el proceso con todas sus conexiones en vuelo.

Aplicar también a `Extension.UnmarshalXML`, que se recursa sin límite.

Dato relevante: la protección propia de `encoding/xml` (tope de 10.000) **se reinicia en cada `UnmarshalJSON`/`UnmarshalXML`**, así que 27.000 niveles reales se aceptan sin error. Actualizar el toolchain por `GO-2026-6088` no arregla esto — verificado: el path de la stdlib aguanta 300.000 niveles y el stack trace del overflow muestra solo frames de `r4`.

### 2.3 · Validar el namespace FHIR

Hoy no hay *ninguna* comprobación: `xmlns="http://evil.example.org/notfhir"` y la ausencia total de namespace parsean igual, y un elemento de un namespace ajeno se consume como campo FHIR.

### ⚠ Dos cosas que NO van en esta fase

**El bump de toolchain sale de aquí.** Subir la directiva `go` no cambia la API, pero para un consumidor con `GOTOOLCHAIN=local` es un fallo de compilación duro — y un patch que no compila donde compilaba el anterior es exactamente lo que semver promete no hacer. Va a la fase 6. Aprovechar para unificar `ci.yml`, `docs.yml` y los `go.mod`, que hoy declaran tres versiones distintas de Go.

**Linealizar el coste no cabe en medio día.** Medido: eliminar la doble pasada de `GetResourceType` da un 2× constante y **la curva sigue siendo cuadrática**.

```
depth   input     doble pasada   una pasada     mejora
1000    40KB      186ms          78ms           2.4x
2000    80KB      597ms          304ms          2.0x
4000    160KB     2.388s         1.207s         2.0x
```

El coste real es **re-escaneo**, no copia: un byte a profundidad *d* se toca ~2(*d*+1) veces. La copia sí explica la memoria — 34,8 MB frente a 1,27 MB con alias de bytes. Ese trabajo son 1,5–3 días, exige fuzzing contra `encoding/json` para probar equivalencia en entrada malformada, y va en la fase 6 como PR separado.

---

## Fase 3 — Higiene del generador y primera impresión

**1,5 días · v1.5.0 · riesgo casi nulo**

> **Objetivo:** entregar valor visible pronto, sin tocar la serialización. Es el release que demuestra movimiento.

### 3.1 · Colisión de nombres de ValueSet

Que la colisión **falle la generación** en vez de continuar con el `continue` silencioso de `template_loader.go:205-208`, con el caso de `Medication.status` parcheado a mano.

Adoptar `binding.extension[bindingName]` como fuente de nombres se aplaza a la fase 6, donde el renombrado masivo ya está anunciado.

Contar además los otros silencios del mismo archivo:
- El descarte de ValueSets con más de 100 códigos (`analyzer.go:653-656`), que degrada bindings *required* grandes a `*string`.
- El `continue` ante un `Analyze` fallido (`codegen.go:104-110`), que puede borrar un recurso entero sin diagnóstico.
- `defMap['Quantity']` resuelve a `SimpleQuantity` porque se indexa por `sd.Type` y tres SDs lo comparten.

Todos deben ser advertencias contadas con resumen final, no silencio.

### 3.2 · Exportar `Ptr`, `Val` y `First`

```go
func Ptr[T any](v T) *T { return &v }
func Val[T any](p *T) T { var z T; if p == nil { return z }; return *p }
func First[T any](s []T) *T { if len(s) == 0 { return nil }; return &s[0] }
```

> **⚠ No es «re-exportar».** No existe nada que exportar: los 23 helpers del repo están todos en `_test.go` o sin exportar. Es API nueva.
>
> **Mina:** `r4/xml_test.go:11` declara `func ptr[T any]` en el propio paquete `r4`, idéntico en r4b y r5 — si el helper se añade en **minúscula**, los tres paquetes dejan de compilar. En mayúscula es seguro (verificado: cero colisiones).

Es además prerrequisito de retirar las 11.952 funciones `With*`, que hoy son el único mecanismo público para obtener un `*string`.

### 3.3 · README, `doc.go` y ejemplos compilables

El ejemplo del README **no compila**:

```
./main.go:13:20: undefined: r4.String
./main.go:14:20: undefined: r4.Boolean
./main.go:16:16: undefined: r4.String
```

- `doc.go` importa un paquete `common` inexistente y usa `ID` donde el campo es `Id` — es una copia de otra librería.
- r4b y r5 **no tienen `doc.go`**.
- En la documentación Hugo hay dos símbolos más que no existen, en ambos idiomas: `r4.HTTPVerbPOST` (es `HTTPVerbPost`) y `helpers.QuantityCel` (es `QuantityCelsius`).

**Aceptación:** sustituir el snippet del README por un test `Example*` compilable, para que la CI impida que vuelva a pudrirse. Añadir la nota de que los structs no validan y que eso vive en `gofhir/validator`.

### 3.4 · `ResourceType` en constructores — **bloquea 6.1**

Ni `NewPatient` ni `NewPatientBuilder` lo inicializan — son 905 constructores en esa situación. Hoy lo tapa `MarshalJSON`, así que es invisible; en cuanto la fase 6 toque ese método, deja de estarlo.

### 3.5 · Marcar deprecaciones y publicar la guía de migración

No hay una sola marca `// Deprecated:` en el repo. Marcar en esta release todo lo que muere en la v2 (`MarshalJSON`, `With*`, `Contained`) es **la mitigación más barata del plan**: hace que los IDEs y `staticcheck` avisen a los usuarios meses antes.

La guía de migración se publica **aquí**, no con la v2. Incluir tabla viejo→nuevo y, para los renombrados mecánicos, un script de codemod.

### 3.6 · Sacar `marshal.go` y `doc.go` del limbo

No existe `marshal.go.tmpl`: `marshal.go`, `doc.go` y `helpers/` están escritos a mano y sobreviven porque alguien los triplicó. **El riesgo ya se materializó** — `doc.go` y `helpers/` existen solo en r4, nunca se crearon para r4b y r5. Una futura R6 saldría sin `Marshal`.

Meterlos en el generador, o documentar explícitamente el paso manual.

> Nota sobre `helpers/`: portarlo a r4b y r5 son 1.027 líneas + 1.184 de test × 2 módulos, con 217 referencias `r4.` a reescribir — **1 día**, no media jornada compartida. Y LOINC/UCUM son independientes de la versión FHIR, así que triplicar el catálogo es deuda pura: conviene discutirlo como módulo compartido con genéricos, o esperar a que 8.x lo genere.

---

## Fase 4 — XML conforme, de una vez

**6–9 días estimados · ~1 día real · v1.7.0 · ✅ el defecto grande, cerrado**

> **Lo que costó de menos, y por qué.** La estimación asumía que `xml.Encoder` no permite escribir bytes crudos, así que preservar el XHTML exigiría reemplazar el encoder. **Esa premisa era falsa:** `,innerxml` sí escribe verbatim al codificar. El defecto era que `encoding/xml` nombra el elemento según el tipo Go, así que un struct anónimo salía como `<rawInner>`, y el decodificador —que discrimina por nombre— nunca lo reconocía. Comprobar la premisa antes de empezar es lo que redujo nueve días a uno.
>
> **XML: de 74/3653 a 3653/3653.** Con dos defectos menores del mismo tema: `xmlEscapeAttr` no escapaba CR/LF/TAB (que XML normaliza a espacio dentro de un atributo, y varios profiles publicados llevan `&#xD;` ahí), y `collapseEmptyElements` reescribía el XHTML del autor.
>
> **El primer test que escribí no servía:** pasaba con el bug reintroducido a propósito, porque una sola pasada deja el div dentro del wrapper y la pérdida solo aparece al releer. Con dos ciclos, la mutación falla con 1.125 narrativas perdidas. La suite de XML solo mide auto-estabilidad, y un documento que pierda siempre el mismo campo converge igual de limpio — que es exactamente cómo esto sobrevivió sin detectarse. Por eso la narrativa se compara ahora contra el archivo publicado.
>
> Siguen vivos: el namespace no se valida al leer, y los `_ext` sobre primitivos dentro de backbones no son representables (afecta también a JSON).

> **Objetivo:** el XML no está roto en varios puntos: está roto de forma **sistémica**, y parchearlo por partes multiplica el coste de regenerar y de rehacer tests.

> **⚠ Fusión respecto de la versión anterior.** El plan separaba el envoltorio `<resource>` (fase 2) del writer de bytes (fase 6), con la pérdida de la narrativa esperando en medio. Pero el `<rawInner>` no se puede arreglar sin control de bytes —se intentó re-emitir tokens y `encoding/xml` produce `xmlns` duplicado que libxml2 rechaza— y ambos defectos son pérdida total de datos.

### 4.1 · Envoltorio `<resource>`, en ambas direcciones — **crítico**

Es un bug de **entrada** además de salida. Un Bundle conforme —lo que devuelve cualquier servidor FHIR— entra y el recurso se pierde con `err=nil`:

```
unmarshal err=<nil>
entries=1
  entry[0].Resource = <nil>     ← el recurso se evaporó

salida: <Bundle><entry><Patient>…</Patient></entry></Bundle>
                        ↑ falta el wrapper <resource>
```

Medido sobre ejemplos oficiales: `diagnosticreport-example.xml` entra con 131.045 bytes y sale con 1.455 — **el 98,9 % de los datos desaparece**. Lo mismo en `Parameters.parameter.resource`. El camino JSON hace esto correctamente.

Añadir `xmlEncodeWrappedResource(e, "resource", res)` y `xmlDecodeWrappedResource(d, start)` — esencialmente los helpers de `contained` con otro nombre de elemento. El prototipo confirma que funciona y mantiene el decoder en sincronía con hermanos posteriores.

**El decoder debe aceptar ambas formas** para no romper el XML ya persistido por la librería.

**Aceptación:** los tests actuales *bendicen* el bug y —esto es lo importante— `xml_test.go:271-278` usa `assert.Contains`, así que **seguiría pasando con el arreglo aplicado**. Hay que invertirlos, no añadir encima. Y existen idénticos en r4b y r5: son **tres copias**.

### 4.2 · Desincronización del decoder — **crítico**

El bucle «consumir hasta cualquier `EndElement`» (`xml_helpers.go:568-577`) sale en el cierre equivocado si hay un hermano suelto:

```
baseline                    contained=1 gender=female birthDate=1980-01-01
con <junk><a/></junk>       contained=1 gender=<nil>  birthDate=<nil>
                            err=<nil>
```

Diferencial de parser explotable: insertando un hermano trivial dentro de `<contained>` se hace desaparecer sin error cualquier campo posterior. Arreglo: `d.Skip()` en los hermanos extraños. Afecta a `contained` hoy y afectaría al envoltorio nuevo si se copia el patrón.

### 4.3 · Writer de bytes

API de cuatro funciones — `Open`, `Close`, `Empty`, `XHTML` — sustituyendo `*xml.Encoder` en escritura y manteniendo `encoding/xml` en lectura. Cierra de una vez:

| Defecto | Hoy |
|---|---|
| Elemento inventado `<rawInner>` | El narrative se pierde al re-parsear, en las tres versiones |
| `collapseEmptyElements` | Una regex reescribe el XHTML del usuario: `<a></a>` → `<a/>`, y de forma inconsistente (`<br></br>` se deja intacto) |
| `xml:lang` en el narrative | Se usa la URI como prefijo y el XML no se puede ni re-parsear |
| Inyección vía `Div` | Se emite XML desbalanceado con `err == nil` |
| Elementos requeridos vacíos | `<code></code>`, inválido contra el schema |
| `Div` sin namespace XHTML | Hereda el de FHIR y queda inválido |

Fijar el alcance en **sustitución mecánica**: cualquier rediseño queda fuera. Un prototipo de ~90 líneas validó la API completa.

Hallazgo colateral a incluir: `xml.Marshal(patient)` directo emite `<Patient>` **sin `xmlns`**, porque la plantilla solo pone el namespace si el nombre llega vacío. Solo `MarshalResourceXML` sale conforme.

### 4.4 · Campos `_ext` escalares en backbones

El bug: los datatypes reciben el campo compañero, los backbones no.

```go
type Coding struct {          // datatype → SÍ
    Display    *string  `json:"display,omitempty"`
    DisplayExt *Element `json:"_display,omitempty"`   ✓
}
type PatientContact struct {  // backbone → NO
    Gender *AdministrativeGender `json:"gender,omitempty"`
    // …no hay GenderExt                                ✗
}
```

820 campos en r4, 981 en r5. El analyzer ya expone el dato (`analyzer.go:575` marca `HasExtension`); son las plantillas de backbone las que lo ignoran.

> **⚠ Alcance corregido: ~6 sitios × 2 plantillas, no diez líneas.** Además del bloque de struct:
> - las dos ramas de marshal XML (escalar y array),
> - la de unmarshal,
> - el `omitempty` hardcodeado que divergió de la rama de recurso,
> - el `id` de backbone —que va como **atributo** XML y colisionaría con un `IdExt` generado—,
> - los helpers `extFieldRef` / `isExtField` en `template_loader.go:406-424`.
>
> Verificado: añadir el campo *sin* tocar las plantillas de XML hace que el XML lo descarte en silencio.

Lo bueno, medido: con el campo a `nil` el `omitempty` lo omite y la salida JSON es **byte-idéntica** a la actual. La suite del repo pasa. Cero colisiones de nombre en 550 structs.

Dos notas:
- `Binary` es uno de los dos únicos structs **comparables** del paquete, así que añadirle un campo *slice* rompería cualquier `==` o uso como clave de mapa. Usar `*Element` es seguro.
- **Los choice types son una segunda clase de bug** no listada antes: `extFieldRef` devuelve `"nil"` cuando `IsChoice`, así que sus extensiones se pierden en ambas direcciones de XML.

**Los `_ext` de arrays NO van aquí** — ver 6.4.

### 4.5 · Gramática de `decimal`, en dos mitades

> **⚠ Se parte.** La gramática estricta de dígitos es un fix limpio: casi todo lo que rechazaría (`0x1p-2`, `1_000.5`, `+1.5`, `.5`, `007`) ya produce hoy un error tardío de marshal, así que endurecer la entrada convierte un fallo confuso en uno claro.
>
> **Pero** el rechazo de decimales *entre comillas* es distinto: hay servidores FHIR reales que los emiten, el código los acepta deliberadamente (`decimal.go:113-115`) y un test lo consagra (`decimal_test.go:44`). Eso rompe a usuarios con datos almacenados → **v2**, o tras un flag de compatibilidad.

El bug de fondo: `strconv.ParseFloat` acepta sintaxis de literal **Go**, y `MarshalJSON` devuelve el texto verbatim. Resultado: envenenamiento persistente — el `POST` devuelve 201 y todo `GET` posterior falla con un error de serialización.

Añadir dos cosas que faltaban:
- El tope FHIR de **18 dígitos significativos**.
- `Decimal.Equal` compara vía `Float64()`, devolviendo `true` para valores que difieren más allá de la precisión de float64 — lo que anula el propósito declarado del tipo.

---

## Fase 5 — Migración de la ruta de módulo

**½ día · prep-v2 · ✅ hecha (código) · doc pendiente**

> **Objetivo:** sin esta fase, la v2 se publica y **nadie puede instalarla**.

Go exige que la ruta del módulo termine en `/v2`, y release-please con `release-type: go` **no edita rutas de módulo**: crearía el tag `r4/v2.0.0` sobre un `go.mod` que sigue declarando `.../r4`. Como el módulo tiene `go.mod`, no aplica `+incompatible`: el proxy rechaza la versión y el tag queda quemado.

Hecho, **antes** de cualquier commit rompiente:

1. ✅ `module github.com/gofhir/models/{r4,r4b,r5}/v2` en los tres `go.mod`.
2. ✅ Los cuatro archivos de `r4/helpers/` que importan el paquete padre (no solo `ucum.go`), los catorce paquetes `_test.go` externos, y los pares `require`/`replace` de `conformance`.
3. ✅ `header.go.tmpl` no contenía la ruta —el plan estaba desactualizado en ese punto—; **ningún** template la menciona, así que regenerar los tres no produce drift. Verificado.
4. ✅ Verificado empíricamente, no por lectura de la documentación de release-please: un clon tageado `r4/v2.0.0`, `r4b/v2.0.0` y `r5/v2.0.0` tal como los emitiría, y un módulo consumidor virgen resolviendo los tres por su ruta `/v2` más `r4/v2/helpers`, y ejecutando contra ellos.

El sufijo va en el `go.mod`, **no** se repite en el tag, y el directorio `r4/` se mantiene sin subcarpeta `v2/`.

El módulo raíz conserva su ruta v1: solo contiene `internal/`, que nadie de fuera puede importar.

### Dos trampas encontradas al ejecutarla

**El commit no puede llevar marca rompiente.** Con `build!:`, release-please cortaría **v2.0.0 con solo el cambio de ruta**, dejando los rompientes reales necesitando una v3. El tipo es `build:` sin `!`, y `build` no está en `changelog-sections`, así que el commit no dispara nada por sí mismo. El major lo cortan los `fix(json)!` que vienen detrás.

**La documentación no se migra aquí.** Sus líneas `go get` apuntarían a un `/v2` que ningún tag satisface todavía, y un push a `main` que toque `docs/**` despliega el sitio de inmediato: la web anunciaría una ruta no instalable durante todo lo que tarde la v2. La migración de la doc es un PR aparte, **retenido hasta que el tag exista** — es un reemplazo mecánico, así que el coste de retenerlo es cero.

**Aceptación:** smoke test post-release — `go get github.com/gofhir/models/r4/v2@v2.0.0` desde un módulo vacío contra proxy.golang.org. Pendiente hasta que el tag se publique; el equivalente contra un repo tageado localmente ya pasa.

### El tag suelto `v1.0.0`

Contamina el listado de versiones del módulo raíz. **Recomendación: dejarlo.** Borrar un tag ya publicado no lo quita de `proxy.golang.org`, que lo sirve desde su caché de forma indefinida, y sí rompe a cualquiera que lo tenga fijado en un `go.sum`. El coste real es cosmético y solo afecta a un módulo que nadie puede importar.

---

## Fase 6 — Rendimiento y modelo

**8–12 días · v2.0.0 · siete cambios rompientes**

### 6.1 · Marcador de tamaño cero para `resourceType`

> **⚠ Sustituye a «borrar los MarshalJSON».** Esa función hace **dos** cosas, y el plan solo contaba una:
>
> ```go
> func (r Patient) MarshalJSON() ([]byte, error) {
> 	r.ResourceType = "Patient"        // ← esto fabrica el campo
> 	type Alias Patient
> 	enc.SetEscapeHTML(false)          // ← solo esto es lo inútil
> 	…
> }
> ```
>
> Sin ella, builders, options y literales emiten `"resourceType":""` —el tag no tiene `omitempty`— y los `contained` quedan sin tipo, así que el dispatcher no puede releerlos nunca. Los tests de round-trip **no lo detectarían**, porque un objeto reparseado sí trae el campo poblado.

```go
type patientTypeMarker struct{}
func (patientTypeMarker) MarshalJSON() ([]byte, error) { return []byte(`"Patient"`), nil }
func (*patientTypeMarker) UnmarshalJSON([]byte) error  { return nil }
// ResourceType patientTypeMarker `json:"resourceType"`
```

Medido: **4810 ns/op frente a 12438 (2,6×)**, contra los 4489 ns del borrado completo. Recupera casi toda la ganancia y mantiene la corrección.

Coste: `p.ResourceType` deja de ser `string` — ruptura de API, pero `GetResourceType()` sigue funcionando.

Sobre el escape: la crítica era correcta pero irrelevante para la decisión. `r4.Marshal` preserva `<` gracias a su *propio* encoder externo, no al interno del método. Documentar en el CHANGELOG **cuatro** cambios, no uno:

1. `ResourceType` deja de auto-rellenarse (de ahí la dependencia con 3.4).
2. `json.Marshal` directo ya no es equivalente a `r4.Marshal`.
3. **Cambia el orden de las claves JSON** — relevante para quien compare payloads byte a byte o los hashee.
4. Cambia el comportamiento de la promoción por embebido para quien incruste `r4.Patient` en su propio struct.

### 6.2 · `ContainedList`

```go
type ContainedList []Resource
func (c *ContainedList) UnmarshalJSON(data []byte) error { /* dispatch */ }
type Patient struct { Contained ContainedList `json:"contained,omitempty"` }
```

Prototipado sobre una copia real: 143 campos cambiados, 143 métodos eliminados, compila. La compatibilidad de fuente aguanta mejor de lo previsto — asignación desde `[]Resource`, `range`, paso a `func f([]Resource)` y `append` siguen funcionando. Los errores con índice se conservan, incluso anidados.

> **⚠ Expectativa corregida.** 1,51× en un Patient suelto pero solo **1,26× en Bundles**, que es el camino caliente real, porque tres campos `Resource` *singulares* conservan su `UnmarshalJSON`: `BundleEntry.Resource`, `BundleEntryResponse.Outcome` y `ParametersParameter.Resource`.
>
> Y la técnica **no se puede extender** a un campo no-slice: se prototipó y un `struct` envolvente ignora `omitempty`, emitiendo `{"resource":null}` —FHIR inválido— en toda entry vacía; la variante con puntero arregla el JSON pero obliga a desreferenciar en todas partes.

Única ruptura observable más allá de lo anterior: `%T` pasa de `[]r4.Resource` a `r4.ContainedList`.

### 6.3 · Linealizar el coste de deserialización

PR **separado**, con fuzzing contra `encoding/json` para probar equivalencia en entrada malformada.

- Eliminar la doble pasada de `GetResourceType` con un scanner de bytes: **306 µs → 0,79 µs** en un bundle de 57 KB. Ojo: usar scanner de bytes, no `json.Decoder`, que aumenta las allocs de 8 a 18.
- Sustituir `[]json.RawMessage` por un alias sin copia: **34,8 MB → 1,27 MB**.

No elimina el carácter cuadrático —eso exigiría decodificadores streaming escritos a mano, otro proyecto— pero baja la constante y la memoria.

### 6.4 · Arrays de primitivos, cambio atómico

> **⚠ Corrección de secuencia.** Los arrays no «pierden el null»: **fabrican datos**.
>
> ```
> in : "given":["A",null,"C"], "_given":[null,{ext},null]
> out: "given":["A","","C"],   "_given":[{},{ext},{}]
> ```
>
> Un `null` posicional se convierte en `""` —un valor vacío inventado, clínicamente distinto de ausente— y el `_given` paralelo re-emite `{}`, que los validadores HL7 rechazan por `ele-1`.
>
> Y el arreglo simétrico obligatorio, `[]Element → []*Element`, no estaba en ninguna fase: si la fase 4 añadiera los `_ext` de arrays en v1.6.0, se publicaría en **minor** una API que se rompe en el siguiente **major**.

Por eso la fase 4 añade solo los escalares, y aquí van los de array junto con `[]string → []*string`, como un **único cambio**. Añadir test de conformidad específico para `_given: [null, {...}]`, que es la razón de ser del cambio.

### 6.5 · Renombrado de enums vía `bindingName`

**✅ Hecha, y una décima parte de lo estimado.** El plan decía **657 tipos y 3.613 constantes**. Lo medido: **48 renombrados (11 r4, 13 r4b, 24 r5) y 365 constantes**. La razón es que los nombres ya coincidían: **186 de los 206 enums de r4** llevaban exactamente su `bindingName`. Lo que faltaba era la minoría, y ahí está todo el valor —`MedicationStatusCodes` → `MedicationStatus`, `Currencies` → `CurrencyCode`.

La extensión no se parseaba: se añadieron `Binding.Extension` y `Binding.Name()` al parser.

**El `bindingName` es dato de la especificación, no un nombre pensado para Go.** Tomarlo a ciegas rompía o empeoraba cosas, y cada regla de rechazo salió de un fallo concreto:

| Se rechaza cuando | Caso que lo obligó |
|---|---|
| varios bindings no coinciden | `request-priority`, enlazado desde 5 puntos en R4 con 5 nombres |
| nombra un recurso | `subscription-status` → `SubscriptionStatus` en R4B/R5: **el paquete no compilaba** |
| dice menos que el título | `verificationresult-status` → `Status`; artifact-assessment → `Disposition`, `InformationType`, `WorkflowStatus` |
| colisiona con otro ValueSet | `Specimen.combined` (grouped\|pooled) anotado `bindingName="PublicationStatus"` |
| no es identificador, o es errata | `appointment-type`, `LOINC LL379-9 answerlist`, `ConceptMapmapAttributeType` |

Las dos últimas son erratas de HL7, no ambigüedades. La regla de «dice menos» es estructural —el `bindingName` es sufijo del nombre derivado del título— no una lista de nombres a mano.

**El override manual de colisiones quedó redundante.** `medication-status` se resuelve ahora por su `bindingName`, y vaciar `valueSetCollisionOverrides` produce salida byte a byte idéntica en las tres versiones. Se conserva porque no cuesta nada y una colisión sin resolver hace fallar la generación: el `bindingName` que hoy la resuelve es dato upstream y puede cambiar.

**Un renombrado no es igual en todas las versiones.** `MedicationStatusCodes` → `MedicationStatementStatus` en r4, pero → `MedicationStatus` en r4b/r5, porque en R4 dos ValueSets comparten el nombre «Medication Status Codes» y el tipo contenía el de MedicationStatement. Un `sed` global sobre un código multi-versión se equivoca aquí; la guía lo dice explícitamente.

Tabla de mapeo viejo→nuevo en la guía de migración, en dos idiomas — y **verificada por test** (`conformance/migration_table_test.go`) contra los paquetes generados y una instantánea de los tipos de v1, para que no se pudra: cada fila debe describir un renombrado real, y ningún tipo puede desaparecer sin fila que diga dónde fue.

`code-systems.md` no necesitó cambios: no nombraba ninguno de los 48.

### 6.6 · Consolidar en builders y retirar `With*`

**✅ Hecha.** 11.952 `With*`, 445 `New<Res>(opts...)` y 445 tipos `<Res>Option` eliminados: **−90.784 líneas** de código generado. No era opcional: la v1.6.0 publicó 12.845 marcas `Deprecated` que dicen «removed in v2», así que sin esto la v2 dejaba esa promesa incumplida y los avisos de staticcheck señalando a algo que no había desaparecido.

La doc costó más que el código: 34 bloques de ejemplo transformados a builder, la página `functional-options.md` eliminada en dos idiomas, la sección equivalente de `builder-api.md` borrada, ocho enlaces a la página muerta arreglados, y «tres patrones» pasado a dos en 20 sitios. Cada ejemplo transformado se compiló y ejecutó.

**11.952 funciones**, no 3.765 —ese número era solo r4— repartidas en 26 páginas de documentación con 154 apariciones. `functional-options.md` desaparece entera.

Extender los builders a datatypes y backbones **primero**; retirar después. Hoy cubren 146/146 recursos y **0** datatypes/backbones, así que la cadena fluida se rompe donde vive la mayor parte de los datos.

Dato del análisis competitivo: **ninguna de las siete librerías Go del ecosistema ofrece builders**, así que completarlos es diferenciador, no paridad.

### 6.7 · Campos complejos requeridos — **nuevo**

No estaba en el plan. Los campos no-puntero se serializan siempre:

```
r4.Observation{Id:"o1"} → {"resourceType":"Observation","id":"o1","code":{}}
r4.Extension{}          → {"url":""}
XML                     → <code></code>
```

Ambos inválidos según `ele-1`. Son 245 campos en r4 (~99 reales), 230 en r4b, 256 en r5. Convertirlos en punteros cambia 99 tipos de campo → v2.

Prueba de que es un descuido: los backbones hardcodean `,omitempty` para *todos* los campos mientras recursos y datatypes lo condicionan a `or .IsArray .IsPointer`.

### 6.8 · Coste de importación y toolchain

**✅ Hecha, y tres de sus cuatro puntos se cayeron al medirlos.**

- ✅ **`sync.OnceValue` para el modelo FHIRPath.** El heap tras cargar `r4` pasa de **1011 KB a 243 KB**. Puesto en contexto: un paquete Go vacío ya usa **215 KB**, así que importar `r4` pasa de costar **+796 KB a +28 KB** sobre el suelo del runtime. Ahí se acaba el margen — `SummaryFields` y `resourceFactories`, los otros dos mapas de nivel superior, caben enteros en esos 28 KB. La primera llamada cuesta 550 µs y las siguientes 1,6 ns sin reservar memoria.
- ❌ **Registro por `init()` para que el linker pode.** La premisa era «+1,8 MB». **Medido: 0,25 MB.** Un binario que solo usa structs pesa 4,81 MB y uno que usa `UnmarshalResource` 5,06 MB, así que el linker ya poda bien y el techo de la tarea es siete veces menor que lo estimado. El grueso del binario no es el registro: son los métodos XML/JSON de 445 tipos. No compensa el riesgo de tocar el dispatch.
- ❌ **Bump de Go y unificar versiones.** Ya estaban unificadas: los seis módulos Go declaran 1.23 (`docs` es un módulo Hugo sin código Go). Subir la mínima de una librería solo excluye usuarios, sin ganancia.
- ❌ **testify solo de test.** Go no tiene dependencias solo-de-test en `go.mod`. El impacto real en un consumidor son **8 líneas de `go.sum`** —su `go.mod` no las lista y el build no las compila— frente a reescribir ~700 llamadas en 11 archivos. No compensa.

> Contexto que relativiza la urgencia: **nadie gana esta partida**. google reparte un paquete por recurso y paga 11 MB igualmente; DAMEDIC paga 21 MB. Los 5,3 MB actuales ya son competitivos.

---

## Fase 7 — Cerrar el corpus

**1–2 días · v2.0.0 · gate no negociable**

Convertir en verdes los `t.Skip` de la fase 1 y activar el gate en CI.

Añadir los fuzz targets como regresión —9,4 millones de ejecuciones sin un panic recuperable merecen conservarse— y el reproductor del DoS **escrito contra profundidad de anidamiento, no contra arrays anchos**:

```
flat n=128000  → 182 ms      ← lineal: un test así pasaría contra código sin arreglar
nested d=2000  → 999 ms      ← aquí está el blowup
```

Cerrar también la brecha entre paquetes:

| | r4 | r4b | r5 |
|---|---|---|---|
| Funciones de test | 103 | 75 | 75 |
| Líneas de test / 1k de producción | 21,0 | 11,1 | **9,2** |

Los nombres de test son idénticos entre r4b y r5: se clonaron, no se escribieron. Los 11 tests de `decimal.go` existen solo en r4 aunque el archivo es **byte-idéntico** en los tres. `fhirpath_model.go` tiene 13 tests en r4 y 4 en r5, que es el más grande (12.204 líneas). Los tests son manuales —no hay `*_test.go.tmpl`— así que regenerar nunca lo arreglará.

---

## Fase 8 — Capa diferenciadora

**variable · v2.1+ · todo aditivo**

Aquí se ocupa terreno que en todo el ecosistema Go está vacío. Ampliada con los hallazgos huérfanos que el review recuperó.

| Tarea | Por qué |
|---|---|
| **`UnknownResource` de reserva** *(nuevo)* | Un solo `resourceType` desconocido destruye el Bundle entero. Sin esto la librería no puede leer respuestas de un servidor de versión más nueva, que es el caso normal. Conservando el JSON crudo, el round-trip incluso mejora. |
| **Validación en el límite** *(reencuadrado)* | Un `IsValid()` suelto no arregla nada: `"gender":"not-a-gender"` se acepta y se re-emite en JSON y XML. La validación de enums tiene que estar **cableada al unmarshal**, con modo estricto opcional, o no existe. |
| **Enums con identidad completa** | `System()`, `Display()`, `Coding()`, `Values()`. Vuelve innecesaria buena parte del `helpers/categories.go` escrito a mano — que además solo existe en r4. Requiere añadir `System`/`Definition` a `parser.ParsedCode`. |
| **Accessors de choice type** | El hueco más claro del ecosistema: 40 campos en `Observation`, 69 en `Extension`, y hoy se pueden setear tres a la vez sin queja. **Aviso:** `ChoiceBaseName` existe pero es un `string` sin discriminador, así que el generador necesitará también `ChoiceTypes` y el `FHIRType` de cada variante — 11 para `Observation.value`. |
| **Extensiones por URL** | No existe en ninguna librería Go. Las utilidades de google están solo en C++, Java y Python. |
| **`Equal()` y `DeepCopy()`** | Solo google los tiene, de rebote por protobuf. Hoy clonar exige round-trip JSON (5,3 µs, 33 allocs) y `reflect.DeepEqual` es semánticamente incorrecto aquí: falla con `Decimal("1.5")` vs `("1.50")` y con `nil` vs slice vacío. |
| **Errores con path FHIR** | Hoy filtran `.Alias` —detalle del truco anti-recursión— y no dicen índice ni subcampo, así que no se puede poblar `OperationOutcome.issue.expression`. El manejo de `contained` ya lleva índice: es el patrón a extender. |
| **Metadata del spec** | 1.697 SearchParameters ya en el repo, 908 `isModifier`, 240 invariantes FHIRPath, 46 OperationDefinition, 5 compartimentos. **Aviso de estimación:** quitar el filtro de `structuredefinition.go:186` son 30 líneas, pero eso solo entrega blobs sin parser, analyzer ni plantilla. La cifra cubría el borrado, no la feature. |
| **Higiene menor** *(nuevo)* | `SummaryFields` es un mapa exportado y mutable (data race demostrada); `_id` de recurso se descarta; `Contained{nil}` emite `[null]`; el desajuste de `resourceType` al deserializar no se valida (un Practitioner se reescribe como Patient); claves `resourceType` duplicadas dan diferencial de parser; `AllResourceTypes()` devuelve orden aleatorio; XML con valores como texto de elemento se acepta y devuelve un recurso vacío. |
| **`Validate()` generado** | Cardinalidad (730 campos `min≥1`), bindings required (369), regex de primitivos (19 de 20 vienen en el spec), rango int32, exclusividad de choice. Coordinar el reparto con `gofhir/validator` **antes** de escribir código. |

---

## Releases y compatibilidad

Lo publicado, y lo que queda. Las tres primeras salieron de `main`; las dos de la línea v1, de `maintenance-v1`.

| Release | Contenido | Estado |
|---|---|---|
| **v1.4.1** | Fase 2 | ✅ Guarda de profundidad, aviso de seguridad. |
| **v1.5.0** | Fase 3 | ✅ Aditivo: `Ptr`/`Val`/`First`, colisión de ValueSets, enums. |
| **v1.5.1** | — | ✅ `Bundle.issues` legible y escribible; `null` explícito tolerado. |
| **v1.6.0** | Gate 2 | ✅ 12.845 marcas `// Deprecated:` y la guía de migración. Tags a mano: release-please no pudo cerrarla con la rama llamada `v1`. |
| **v1.7.0** | Fase 4 | ✅ La narrativa XML. XML pasa a 3653/3653. Publicada sola, lo que valida el renombrado de la rama. |
| **v2.0.0** | Fases 5, 6 y 7 | Siete cambios rompientes de una vez, con ruta `/v2`. La fase 5 aterriza en `main` **antes**, sin marca rompiente, para que no corte el major por sí sola. |
| **v2.1+** | Fase 8 | Aditivo, incremento a incremento. |

### Los siete cambios rompientes

1. Arrays de primitivos a `[]*string` **más** `[]*Element` (uno solo, atómico).
2. Marcador de tamaño cero para `resourceType` (`p.ResourceType` deja de ser `string`).
3. `Contained` de `[]Resource` a `ContainedList`.
4. Renombrado de enums vía `bindingName` (657 tipos, 3.613 constantes).
5. Retirada de las 11.952 funciones `With*`.
6. Campos complejos requeridos a puntero (~99 campos reales en r4).
7. Rechazo de decimales entre comillas.

### Dos gates antes de cortar la v2

Agrupar los rompientes es correcto —el coste real de un major en Go es el sufijo y la migración de imports, que se paga igual con uno que con siete— pero con la cobertura actual una v2 de siete cambios es una apuesta. **No cortarla sin:**

1. ✅ El corpus de la fase 1 en verde.
2. ✅ Todo lo que muere en v2 marcado `// Deprecated:` — en la **v1.6.0**, no en la v1.5.0: esa ventana se cerró al publicar v1.5.0 y v1.5.1 sin las marcas.

**El gate 2 obligó a crear la rama `v1`, no por preferencia sino porque no había alternativa.** Con `main` ya en `/v2`, Go rechaza cualquier tag v1.x sobre él:

```
invalid version: r4/go.mod has post-v1 module path
"github.com/gofhir/models/r4/v2" at revision r4/v1.6.0
```

Verificado, no deducido. Y las marcas no sirven de nada dentro de la propia v2, donde los símbolos ya no existen. La rama sale de `r4/v1.5.1` (6c843b2), antes de que `main` moviera las rutas.

Lo marcado son solo los símbolos que **desaparecen**, no los que cambian de tipo: 11.952 `With*`, 445 `New<Res>(opts...)` y 445 tipos `<Res>Option`. Un `// Deprecated:` sobre `ResourceType` o `Contained` sería engañoso —el campo sigue existiendo con el mismo nombre— así que esos van a la guía de migración.

Dos comprobaciones que valía hacer antes de emitir 12.845 mensajes:

- **El aviso llega al usuario y no rompe al repo.** `staticcheck` omite SA1019 dentro del mismo módulo, así que los tests propios siguen en verde, mientras un consumidor externo sí recibe el aviso con el reemplazo nombrado. Comprobado en ambas direcciones con un módulo aparte.
- **La equivalencia que el mensaje promete es real.** Un test parsea el código generado y compara, para las 11.952, el tipo del parámetro y la operación del cuerpo (`append` / `&v` / asignación). Cero discrepancias. Sin eso, el mensaje sería una promesa sin respaldo repetida doce mil veces.

### Política de la rama `maintenance-v1`

- **Recibe:** solo arreglos de seguridad. **El desarrollo pasa a `main` en exclusiva.**
- **No recibe:** funcionalidad, arreglos de conformidad ni nada que cambie una firma o un tipo.
- **Cuánto:** hasta que la v2 lleve un ciclo publicada.
- **Mecánica:** el mismo workflow de release-please sirve a las dos ramas con `target-branch: ${{ github.ref_name }}`. Cada rama lleva su propio manifest. El CI también corre en `maintenance-v1`; sin eso, sus commits de release se publicarían sin verificar.

Lo publicado en ella: **v1.6.0** (12.845 marcas de deprecación) y **v1.7.0** (la narrativa XML). Ese es su alcance final salvo emergencia.

### El nombre de la rama no es cosmético

Se llama `maintenance-v1` y no `v1` porque con el nombre `v1` **release-please no puede terminar un release**. El título por defecto de un manifest multi-paquete es `chore: release ${branch}`, así que al mergear parsea ese `v1` final como versión y muere:

```
Error: unable to parse version string: 1
```

Reproducido localmente contra el repo real, lo que además mostró que el fallo se limita al paso `github-release`: `release-pr` funciona. Por eso la v1.6.0 quedó a medias —manifest y CHANGELOG dentro, tags fuera— y hubo que etiquetarla a mano. `main` recorre esa misma ruta en cada release y funciona precisamente porque «main» no parece una versión. La v1.7.0 se publicó sola, lo que confirma el arreglo de punta a punta.

### Ningún release de `main` por debajo de 2.0.0 es instalable

Consecuencia directa de la fase 5, y hay que tenerla presente en cada merge hasta que la v2 salga. `main` declara `/v2`, así que Go rechaza cualquier tag v1.x sobre él, **y el intento quema ese número para siempre**.

El manifest de `main` decía 1.5.1 mientras `r4/v1.6.0` y `r4/v1.7.0` ya existían en la otra rama: un `feat` habría propuesto 1.6.0, un tag ya ocupado. Sincronizado a 1.7.0 —lo realmente publicado— y con `Release-As: 2.0.0` en el commit del port para que el siguiente release sea 2.0.0 pase lo que pase antes. Verificado con release-please contra la rama, no supuesto.

---

## Riesgos y capacidad

> ### ⚠ Las estimaciones son días-persona, no calendario
>
> El historial del repositorio son **21 commits humanos en 7 meses** de un solo mantenedor, con un hueco de cuatro meses. 21–28 días-persona a ese ritmo no son «cinco semanas»: son varios trimestres. El plan tampoco reserva tiempo para revisión de PRs, issues entrantes ni el ciclo de release.
>
> Si el objetivo es cerrar lo crítico rápido, **las fases 0 a 2 son tres días** y son separables del resto.

| Riesgo | Mitigación |
|---|---|
| Los PRs de plantilla traen diffs de decenas de miles de líneas generadas | Test golden de la fase 0, y separar cada PR en «cambio de plantilla» y «regeneración» como commits distintos |
| Tests que bendicen el bug seguirían pasando tras el arreglo | Confirmado en `xml_test.go:271-278`, que usa `assert.Contains`. Invertirlos, no añadir encima — y en los tres paquetes |
| La documentación queda mintiendo sin que nada avise | 72 páginas bilingües; `docs.yml` solo se dispara con cambios en `docs/**`. Blast radius por cambio: `MarshalJSON` 10 páginas · `Contained` 8 · arrays 22 · enums 40 · `With*` **26 páginas / 154 hits**. Una tarea de docs **por fase**, más un check que falle si `r4/` cambia sin que cambie `docs/` |
| La fase 4 se desborda | El prototipo del writer ya validó la API de cuatro funciones. Alcance fijado en sustitución mecánica |
| La fase 8 duplica el alcance de `gofhir/validator` | Repartir antes de escribir: lo generable sin red aquí, las invariantes FHIRPath allá |
| Fatiga de v2: nadie migra y quedan dos líneas vivas | Un solo major, anunciado con una minor de antelación y con deprecaciones marcadas |
| Marcar el XML como experimental llega tarde | Hacerlo **ya**, antes de la fase 4: hoy el README anuncia un round-trip XML que no se cumple |

---

## Dos cosas que hacer hoy, antes de tocar código

1. **Marcar el soporte XML como experimental** en `docs/content/{en,es}/docs/serialization/xml-marshaling.md` y quitar del README la promesa de round-trip XML. Es gratis y evita que más gente construya sobre un comportamiento que va a cambiar.
2. **Coordinar el alcance de `Validate()` con `gofhir/validator`.** Si no se decide antes, se duplica trabajo.

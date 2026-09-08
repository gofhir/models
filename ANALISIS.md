# Análisis de gofhir/models — brechas, fortalezas y debilidades

> Fecha: 2026-09-07. Árbol analizado: rama `fix/align-ext-slot-arrays`, HEAD `22ecd7f`, módulos en 2.7.0.
> Cada hallazgo relevante se verificó por reproducción o medición, no solo por lectura. Después del análisis se hizo una pasada adversarial contra las propias conclusiones; las severidades que aparecen aquí son las que sobrevivieron a esa pasada, y la sección final registra qué cambió.

## Índice

- [Veredicto](#veredicto)
- [Método](#método)
- [Fortalezas](#fortalezas)
- [Defectos verificados](#defectos-verificados)
- [Cómo se consume el StructureDefinition](#cómo-se-consume-el-structuredefinition)
- [Bindings y ValueSets](#bindings-y-valuesets)
- [API generada: limitaciones de diseño](#api-generada-limitaciones-de-diseño)
- [Proceso y medición](#proceso-y-medición)
- [El consumidor principal no está en v2](#el-consumidor-principal-no-está-en-v2)
- [Otros archivos base del spec](#otros-archivos-base-del-spec)
- [Prioridades](#prioridades)
- [Pasada adversarial: qué cambió](#pasada-adversarial-qué-cambió)

---

## Veredicto

La librería es sólida donde el spec es estructural: tipos, cardinalidad, JSON, guardas de seguridad, builders. Es frágil donde el spec es terminológico o donde el corpus de conformidad no puede mirar: el XML comparado solo contra sí mismo, los ValueSets compuestos y el `isSummary` de los choice types.

Del zip oficial de definiciones el generador consume 3 de 12 archivos en R4 y 3 de 9 en R5. Hay valor sin explotar, pero menos del que parece a primera vista, porque el resto de la organización ya resuelve varias de esas necesidades por su cuenta.

El hallazgo más consecuente no está en este repo: `gofhir/server`, el consumidor principal, sigue fijado en `models` v1.6.0 y nada de lo entregado en la v2 le llega.

---

## Método

1. Tres auditorías paralelas sobre el árbol: pipeline del generador (`internal/codegen`), API pública generada (`r4/`, `r4b/`, `r5/`), e infraestructura (conformance, CI, docs, releases).
2. Descarga de `definitions.json.zip` de R4 y R5 para inventariar los archivos que el generador no lee, y `jq`/Python sobre los specs para cuantificar bindings, constraints, `isModifier`, `isSummary`, etc.
3. Reproducciones dirigidas: un programa externo que consume los módulos vía `replace` para reproducir cada defecto.
4. Medición del camino XML sobre los 3653 archivos del corpus: firma de tokens del archivo publicado contra la del re-emitido, elemento por elemento y atributo por atributo, ignorando `xmlns`. Es la comparación contra la fuente que la suite hoy no hace.
5. Búsqueda de código en la organización `gofhir` para saber qué partes de la API se usan realmente fuera de este repo.

---

## Fortalezas

**Corpus de conformidad real y con trinquete.** 8757 JSON y 3653 XML publicados por HL7 pasan en las tres versiones, con las seis listas de fallos conocidos en cero. El JSON se compara semánticamente contra el archivo original, con `UseNumber` para distinguir `2.00` de `2.0` y comparación del tipo Go para atrapar el caso `integer64`. El trinquete falla en ambas direcciones: regresión y progreso no registrado.

**Fidelidad representacional poco común en Go.**
- `Decimal` textual; `integer64` como string en JSON.
- `null` posicional en primitivos repetidos con `[]*T` y `[]*Element` alineados.
- `resourceType` por marcador de tamaño cero: el cero-valor de cualquier recurso ya serializa bien.
- Campos complejos requeridos como puntero para no emitir `{}`, que viola `ele-1`.
- `UnknownResource` que preserva bytes en JSON y XML: un servidor R5 no rompe a un cliente R4 por una entrada.

**Seguridad tratada en serio.** Guarda de profundidad O(n) que cuenta recursos anidados y no llaves, con las evasiones reales reproducidas y cubiertas: `contained` antes de `resourceType`, claves escapadas, casing. SECURITY.md documenta el incidente y la superficie.

**Builders completos.** 663 en r4, 681 en r4b, 834 en r5. Cubren recursos, datatypes, backbones y los compañeros `_field`, con exclusividad de choice types. Ninguna otra librería Go del ecosistema lo ofrece.

**Ingeniería del generador.**
- Specs fijados por sha256 (`specs.lock`), con verificación en la descarga.
- Job de drift que regenera las tres versiones y falla ante cualquier diferencia o archivo sin trackear, más una aserción positiva de recuento de enums.
- Golden tests con fixtures deliberadamente adversariales.
- Determinismo explícito en cada iteración de mapa.
- Reglas de `bindingName` derivadas de fallos concretos, no de listas a mano.

**Higiene de módulos.** Cero dependencias en runtime, tres módulos independientes en Go 1.26, `conformance/` como módulo aparte para que lo publicado no cargue el corpus. Docs bilingües con la tabla de migración v1→v2 verificada por test.

---

## Defectos verificados

### 1. XML pierde las extensiones de primitivos en choice types

**Qué pasa.** Para toda variante `[x]` de tipo primitivo, el encode XML pasa `nil` en vez del campo `<Campo>Ext` y el decode descarta la extensión con `_ = ext`. El struct sí tiene el campo y el JSON lo preserva, así que es una asimetría JSON/XML no documentada.

```go
// r4/resource_patient.go:269
xmlEncodePrimitiveBool(e, "deceasedBoolean", r.DeceasedBoolean, nil)
// r4/resource_patient.go:452
v, ext, err := xmlDecodePrimitiveBool(d, t)
r.MultipleBirthBoolean = v
_ = ext
```

**Reproducción.** Un `Patient` con `_deceasedDateTime` extendido entra por JSON, sale de `MarshalResourceXML` sin la extensión, y al volver a JSON ya no está.

**Alcance en el código generado.**

| Sitios con `nil` en encode | r4 | r5 |
|---|---|---|
| Recursos | 263 | 417 |
| Datatypes (incluye `Extension.value[x]`) | 117 | 126 |

**Alcance medido en el corpus.** Comparando firmas de tokens del original contra el re-emitido:

| Versión | Archivos | Con pérdida | Qué se perdió |
|---|---|---|---|
| r4 | 1138 | 1 | `v2-tables.xml`: 1294 extensiones `translation` sobre `valueString`, 12 946 tokens |
| r4b | 1156 | 0 | nada |
| r5 | 1359 | 1 | `valuesets.xml`: 2 extensiones sobre `valueMarkdown` |

El resto del corpus está limpio a nivel de token. La suite no lo detecta porque el XML solo se compara contra su propia salida (`conformance/roundtrip_test.go:226` lo admite); el gemelo `v2-tables.json` pasa porque el JSON sí se compara contra la fuente.

**Severidad.** Real, en uso y acotada. En el corpus publicado lo que se pierde son traducciones sobre metadatos terminológicos, no datos clínicos. Lo que la eleva es que `gofhir/server` usa `MarshalResourceXML` y `UnmarshalResourceXML` en producción, así que el camino afectado no es teórico.

**Origen.** El analyzer añade las variantes de choice como propiedades reales con su `<Name>Ext`, pero las plantillas XML solo cablean el compañero cuando lo emiten ellas mismas desde `HasExtension`. `template_loader.go` tiene un `extFieldRef` que devuelve `"nil"` cuando `IsChoice`.

### 2. `FHIRTypes` en r5 es inservible

El parser de ValueSet ignora `compose.include[].valueSet`. `version-independent-all-resource-types` compone `include[0].valueSet=[all-resource-types]` más `include[1].system=fhir-old-types`; solo se resuelve el segundo. El enum queda con 41 constantes, todas de tipos obsoletos (`FHIRTypesBodysite`, `FHIRTypesConformance`), y sin `Patient`. Afecta a `GraphDefinition.node.type`, tipado `*FHIRTypes` en `r5/resource_graphdefinition.go:978`.

**Sobre el arreglo.** Resolver `include.valueSet` no repara el enum: `all-resource-types` tiene unos 162 códigos y el tope `maxEnumCodes` es 100, así que el campo caería a `*string`. Eso es más honesto que un enum incorrecto, pero un enum útil exige subir el tope o una regla dedicada. Es el único ValueSet con `include.valueSet` ligado como `required` sobre `code` en las tres versiones.

### 3. `summary.go` omite todos los choice types

`analyzeChoiceType` no propaga `isSummary`, así que `Observation` aparece sin `value` ni `effective` aunque el spec marca `Observation.value[x]`, `Observation.effective[x]` y `Observation.component.value[x]` con `isSummary=true`. La tabla además solo cubre propiedades de primer nivel, no backbones. Ver `r4/summary.go:1383`.

**Severidad.** Baja en la práctica: ningún repo de la organización llama a `GetSummaryFields`, y el servidor resuelve `_summary` por su cuenta. Sigue siendo una tabla incorrecta publicada bajo un comentario que promete ser el conjunto `_summary`.

### 4. Documentación de paquete obsoleta

`r4/doc.go:27` (y sus copias en r4b y r5) sigue anunciando functional options que se retiraron en la v2, y en la línea 41 describe el bug del wrapper de `Narrative.Div` como vigente cuando se corrigió en el commit `cd5f841`. Sale de `doc.go.tmpl:27` y `:42`.

### 5. Mensaje de error de colisión con variable inexistente

`template_loader.go:386` remite a `valueSetTypeNameOverrides`; la variable real es `valueSetCollisionOverrides` en `analyzer.go:872`. El golden test fija la cadena equivocada, con lo que el error queda blindado en su forma incorrecta.

---

## Cómo se consume el StructureDefinition

El generador lee `profiles-types.json`, `profiles-resources.json` y `valuesets.json`. Un solo `Analyzer`, sin ramas por versión.

### Campos de ElementDefinition

| Campo | Estado | Nota |
|---|---|---|
| `path`, `max`, `type[0].code`, `type[].targetProfile`, `contentReference`, `short`, `isSummary`, `binding` (`strength`, `valueSet`, `bindingName`) | Consumidos | `targetProfile` solo para `Reference` y `canonical` |
| `min` | Parseado, ningún template lo lee | Punteros a propósito; es dato muerto, no brecha |
| `type[1..]` fuera de `[x]` | Se pierde | Solo se mira `Type[0]`; no hay casos en el core |
| `constraint[]` | Solo del elemento raíz, sin consumidor | Los invariantes por elemento ni se recogen |
| `isModifier`, `mustSupport`, `maxLength`, `fixed[x]`, `pattern[x]`, `condition`, `mapping`, `definition`, `comment`, `base`, `type[].profile` | Parseados y descartados | Los tags JSON de `fixed`/`pattern`/`example.value` nunca coincidirían con el JSON real (`fixedUri`, `patternCodeableConcept`) |
| `slicing`, `representation`, `defaultValue[x]`, `meaningWhenMissing`, `orderMeaning`, `binding.additional` | Ni parseados | Los slices se descartan por `sliceName != ""` |

**`representation` está hardcodeado.** El `id` como atributo, `Extension.url` como atributo, primitivos como `value=` y `xhtml` como markup crudo se deciden en las plantillas. Verificado contra el spec: son exactamente los casos que el core declara, así que la cobertura es completa, pero es coincidencia mantenida a mano.

**Primitivos sin tipo.** Todo lo temporal, `uri`, `base64Binary` y `code` sin binding `required` es `*string`. No hay validación de formato en ningún punto, por decisión declarada en `doc.go`: la validación vive en `gofhir/validator`. Un `date` tipado sería representación y no validación, pero cambiarlo es un major.

**Código muerto en el generador.** `analyzer.TypeHierarchy()`, `templates/header.go.tmpl`, `AnalyzedProperty.Binding`, `AnalyzedType.{Constraints,URL,IsAbstract,ParentResource}`, `typemap.GoTypeRequiresPointer`, `typemap.PrimitiveTypesNeedingExtension`.

**Riesgo latente.** `resolveContentReference` desreferencia `sd.Snapshot.Element` sin comprobar nil; un SD con solo differential haría panic. No ocurre con los specs core.

---

## Bindings y ValueSets

Enum solo si se cumple todo: `type[0].code == "code"`, `strength == "required"`, el ValueSet resuelve, y tiene entre 1 y 100 códigos. Resultado: 206 enums en r4, 215 en r4b, 237 en r5.

| Situación | Efecto | Valoración |
|---|---|---|
| `extensible`, `preferred`, `example` | `*string` | Correcto: conjuntos abiertos |
| `required` sobre `CodeableConcept` (14 en R4) | sin enum | Correcto; a lo sumo, constantes |
| ValueSet con más de 100 códigos (`all-types`, `resource-types`, `spdx-license`…) | `*string` | Deliberado; discutible para `resource-types` |
| CodeSystem con `content=not-present` (`currencies`, `mimetypes`) | `*string` por lista vacía | Resultado correcto por accidente: `content` se parsea y nunca se consulta |
| `compose.include[].filter` (67 ValueSets en R4) | se expande el CodeSystem entero | Sobre-inclusión latente; cero casos vivos con `required`+`code` |
| `compose.include[].valueSet` | se ignora | Un caso vivo: `FHIRTypes` en r5 |
| `compose.exclude` | se ignora | Cero casos vivos |
| `ValueSet.expansion` | no se lee | Los bundles core casi no traen expansiones |

Sistemas externos con binding `required` en r5 que legítimamente quedan como `*string`: BCP-47 (`Resource.language`), ISO 4217, BCP-13, IANA time zones, UCUM, ISO 3166, y varios `terminology.hl7.org` que ya no viajan en el zip de R5 (`condition-clinical`, `allergyintolerance-verification`, `research-subject-state`…).

---

## API generada: limitaciones de diseño

- **Choice types aplanados.** Una propiedad por variante, sin tipo suma ni accesor de lectura. `Extension` tiene unos 50 campos `Value*`. Solo los builders garantizan exclusividad; un struct literal puede llevar dos variantes.
- **`json.Marshal` frente a `r4.Marshal`.** El estándar escapa `<`, `>` y `&` y corrompe la narrativa; nada lo impide más que la documentación.
- **Sin `Equal()` ni `DeepCopy()`.** Clonar exige round-trip JSON; `reflect.DeepEqual` es semánticamente incorrecto con `Decimal("1.5")` y `("1.50")`.
- **Extensiones.** Hay lookup por URL en 660 tipos, pero no `AddExtension` por URL, ni acceso tipado al `value[x]`, ni lookup plural de `modifierExtension`. Cero usos de `GetExtensionByURL` fuera de este repo: la demanda no está probada.
- **Modelo FHIRPath.** Es solo proveedor de tipos: sin cardinalidad, sin invariantes, sin navegación reflexiva. `gofhir/fhirpath` tiene su propio `model.go`, evalúa sobre JSON crudo con `jsonparser` y no depende de `models`. El archivo más grande de cada paquete no tiene consumidor real en la organización; al ser perezoso, su coste en runtime es pequeño.
- **Errores.** Nombran tipo y campo, pero no el índice dentro de arrays de datatypes (`Patient.name[1].family`), así que no sirven para poblar `OperationOutcome.issue.expression`.
- **Lenient por defecto.** Campos JSON y elementos XML desconocidos se ignoran en silencio, sin opción estricta.

---

## Proceso y medición

- **El XML del corpus solo se compara consigo mismo.** La medición de firmas de tokens muestra que ese hueco escondió exactamente una clase de defecto (el 1) y ninguna otra. La comparación corre en segundos y podría entrar en la suite como oráculo contra la fuente.
- **Un solo benchmark y un solo fuzz target.** Los números de rendimiento citados en PLAN.md y en comentarios del código se midieron fuera del árbol y no son reproducibles desde el repo. El corpus completo corre en 89 s, así que el rendimiento no es un problema abierto.
- **`conformance/` no pasa por lint.** Ningún job de `govulncheck`. `testify` viaja en el `go.mod` publicado.
- **Cobertura reportada del 1,1 % en r4.** Es aritmética sobre 294 000 líneas generadas; la métrica real es el corpus, y Codecov confunde a quien lo mira.
- **Un solo mantenedor**, 17 ramas locales sin limpiar tras merge, y `v2.0.0` ininstalable por el sumdb (documentado en README).

---

## El consumidor principal no está en v2

`gofhir/server` fija en su `go.mod`:

```
github.com/gofhir/models/r4  v1.6.0
github.com/gofhir/models/r4b v1.6.0
github.com/gofhir/models/r5  v1.6.0
```

No está en la v2 ni siquiera en la v1.7.0 que corrigió la narrativa XML. Nada de lo entregado en la v2 le llega: extensiones en backbones, `integer64` como string, campos requeridos como puntero, `UnknownResource`, `ContainedList`. El riesgo que PLAN.md llamó "fatiga de v2" está materializado dentro de la propia organización, y el servidor usa el camino XML, que es donde vive el defecto 1.

También relevante para las decisiones de alcance: el servidor descarga y embebe `search-parameters.json` por su cuenta (`pkg/fhirdefs/embed.go`, `scripts/download-fhir-defs.sh`), y `gofhir/fhirpath` y `gofhir/validator` trabajan sobre JSON crudo sin depender de `models`.

---

## Otros archivos base del spec

Contenido de `definitions.json.zip` y artefactos vecinos, con lo que aportaría cada uno. El valor está ajustado por lo que la organización ya resuelve por otra vía.

| Archivo | R4 | R5 | Qué habilita | Valor |
|---|---|---|---|---|
| `search-parameters.json` | 1375 SP | 1239 SP | Constantes tipadas de búsqueda por recurso con `type`, `expression` y targets. Aditivo. | Solo si `server` lo adopta; hoy lo embebe él mismo, y el README de `models` declara que no hace búsqueda |
| `extension-definitions.json` | 393 SD, 36 complejas | no viene (paquete `hl7.fhir.uv.extensions`) | Accesores tipados para extensiones core; cierra la brecha del `value[x]` aplanado | Demanda no probada |
| `expansions.json` (fuera del zip) | sí | sí | Expansiones oficiales: resuelve `filter`, `include.valueSet` y el tope de 100 sin expandir CodeSystems a mano | Alto para los bindings; arregla `FHIRTypes` de raíz |
| `fhir-all-xsd.zip` (fuera del zip) | sí | sí | Oráculo externo para el XML: validar la salida con `xmllint` | Alto para tests; complementa la firma de tokens |
| `profiles-others.json` | 44 perfiles | 62 perfiles | Vitalsigns, bp, `Shareable*`, los ocho tipos de Bundle en R5. Exige slicing, `fixed` y `pattern` | Medio, largo plazo |
| `fhir.schema.json.zip` | sí | sí | JSON Schema como segundo oráculo de forma en golden tests | Medio |
| `v2-tables.json`, `v3-codesystems.json` | 424 + 143 CS | no vienen (paquete `hl7.terminology`) | Resuelve `v3-ConfidentialityClassification`, hoy ausente | Bajo |
| `conceptmaps.json` | 19 | 22 | Mapas core, casi todos hacia v2/v3 | Bajo |
| `choice-elements.json`, `backbone-elements.json` (fuera del zip) | sí | sí | Oráculo barato para el analyzer | Bajo, útil en tests |
| `dataelements.json` | 6781 SD | 308 SD | Un SD por elemento; redundante con los snapshots | Nulo |
| `version.info` | sí | sí | Fija la versión FHIR sin escanear SDs | Cosmético |

Para paridad de R5 con R4 en extensiones y terminología hay que descargar los paquetes NPM `hl7.fhir.uv.extensions` y `hl7.terminology`, con su propio hash en `specs.lock`.

---

## Prioridades

1. **Defecto 1**, con la comparación de firmas de tokens incorporada a la suite como oráculo del XML contra la fuente. Es pérdida de datos en un camino que el servidor usa.
2. **Migrar `gofhir/server` a `models` v2.** No es trabajo de este repo, pero es donde el valor ya construido no está llegando.
3. **`FHIRTypes` y el `isSummary` de choices**, con el arreglo bien planteado: resolver `include.valueSet` y decidir el tope, o traer `expansions.json`.
4. **Refrescar `doc.go.tmpl`** y el mensaje de colisión.
5. **Search parameters y extensiones tipadas** solo si el servidor las va a consumir. Si no, son el mismo argumento con que se descartó `Validate()`.

---

## Pasada adversarial: qué cambió

| Afirmación original | Ataque | Resultado |
|---|---|---|
| XML pierde extensiones de choice: "pérdida de datos clínicos" | Medir los 3653 archivos, no un ejemplo propio | Confirmado; severidad corregida: 2 archivos afectados, metadatos terminológicos. Sube por el uso del XML en `server` |
| "El hueco de compararse contra sí mismo cuesta datos" | Cuantificar cuánto | El hueco es real pero escondió exactamente una clase de defecto; el resto del corpus está limpio a nivel de token |
| `FHIRTypes`: resolver `include.valueSet` lo arregla | Contar los códigos resultantes | Refutado el arreglo: 162 códigos, tope 100, caería a `*string` |
| `summary.go` sin choices | Buscar consumidores | Confirmado contra el spec; severidad baja, nadie en la organización lo usa |
| "Extensible y preferred se ignoran" como debilidad | Revisar la semántica | Retirado: es correcto por diseño |
| `content=not-present` no consultado | Ver el resultado producido | No es bug: el resultado es correcto |
| `min` ignorado como brecha | Buscar quién lo leería | Reclasificado como dato muerto |
| Modelo FHIRPath incompleto | Ver quién lo consume | Rebajado: `gofhir/fhirpath` tiene su propio modelo y no depende de `models` |
| Search parameters "valor alto" | Ver qué hace `server` | Debilitado: el servidor ya embebe el archivo; el README de `models` excluye búsqueda |
| Extensiones tipadas "valor alto" | Buscar usos de `GetExtensionByURL` | Cero usos externos; demanda no probada |
| Sin benchmarks | Ver si el rendimiento es un problema abierto | Se sostiene como observación; el corpus corre en 89 s |
| — | Revisar dependencias de los repos hermanos | Nuevo: `server` en v1.6.0, sin nada de la v2 |

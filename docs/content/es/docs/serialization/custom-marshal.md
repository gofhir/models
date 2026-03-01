---
title: "Marshal Personalizado"
linkTitle: "Marshal Personalizado"
description: "Serialización JSON segura para HTML con r4.Marshal() para preservar el contenido XHTML narrativo de FHIR."
weight: 3
---

La biblioteca `gofhir/models` proporciona funciones de marshaling JSON personalizadas que resuelven un problema crítico con el `json.Marshal` estándar de Go: el escape de entidades HTML en el contenido narrativo de FHIR.

## El Problema del Escape HTML

La función estándar `json.Marshal` de Go escapa los caracteres HTML (`<`, `>`, `&`) en cadenas a sus secuencias de escape Unicode (`\u003c`, `\u003e`, `\u0026`). Esta es una medida de seguridad para prevenir ataques XSS cuando JSON se embebe en HTML, pero rompe el contenido narrativo de FHIR.

Los recursos FHIR pueden incluir un campo `text.div` que contiene XHTML válido. La especificación FHIR requiere que este contenido se preserve exactamente como se proporcionó. Cuando el marshaler estándar de Go escapa las entidades HTML, el JSON resultante es técnicamente válido pero viola la especificación FHIR y puede causar problemas con servidores y clientes FHIR que esperan XHTML sin escapar.

```go
package main

import (
    "encoding/json"
    "fmt"

    "github.com/gofhir/models/r4"
)

func ptrTo[T any](v T) *T {
    return &v
}

func main() {
    patient := &r4.Patient{
        ResourceType: "Patient",
        Id:           ptrTo("narrative-example"),
        Text: &r4.Narrative{
            Status: ptrTo(r4.NarrativeStatusGenerated),
            Div:    ptrTo(`<div xmlns="http://www.w3.org/1999/xhtml"><p>John Smith, Male, DOB: 1990-01-15</p></div>`),
        },
    }

    // Standard json.Marshal escapes HTML entities
    data, _ := json.Marshal(patient)
    fmt.Println(string(data))
    // Output contains escaped HTML:
    // "div":"\u003cdiv xmlns=\"http://www.w3.org/1999/xhtml\"\u003e\u003cp\u003eJohn Smith, Male, DOB: 1990-01-15\u003c/p\u003e\u003c/div\u003e"
}
```

La salida escapada `\u003cdiv\u003e` es JSON válido, pero muchas implementaciones FHIR esperan los caracteres HTML literales. Esto puede llevar a errores de validación, problemas de visualización o problemas de interoperabilidad.

## La Solución: r4.Marshal()

La función `r4.Marshal()` usa `json.NewEncoder` con `SetEscapeHTML(false)` para producir una salida JSON que preserva los caracteres HTML exactamente como aparecen en las cadenas de Go:

```go
package main

import (
    "encoding/json"
    "fmt"

    "github.com/gofhir/models/r4"
)

func ptrTo[T any](v T) *T {
    return &v
}

func main() {
    patient := &r4.Patient{
        ResourceType: "Patient",
        Id:           ptrTo("narrative-example"),
        Text: &r4.Narrative{
            Status: ptrTo(r4.NarrativeStatusGenerated),
            Div:    ptrTo(`<div xmlns="http://www.w3.org/1999/xhtml"><p>Hello</p></div>`),
        },
    }

    // Standard json.Marshal escapes HTML entities
    data1, _ := json.Marshal(patient)
    fmt.Println(string(data1))
    // {"resourceType":"Patient","id":"narrative-example","text":{"status":"generated","div":"\u003cdiv xmlns=\"http://www.w3.org/1999/xhtml\"\u003e\u003cp\u003eHello\u003c/p\u003e\u003c/div\u003e"}}

    // r4.Marshal preserves HTML as required by FHIR
    data2, _ := r4.Marshal(patient)
    fmt.Println(string(data2))
    // {"resourceType":"Patient","id":"narrative-example","text":{"status":"generated","div":"<div xmlns=\"http://www.w3.org/1999/xhtml\"><p>Hello</p></div>"}}
}
```

## Firmas de Funciones

La biblioteca proporciona dos funciones de marshal personalizadas, definidas en `marshal.go`:

### Marshal

```go
func Marshal(v interface{}) ([]byte, error)
```

Serializa cualquier valor a JSON sin escape HTML. Internamente, crea un `bytes.Buffer`, configura un `json.NewEncoder` con `SetEscapeHTML(false)`, y elimina el salto de línea final que `Encode` agrega.

### MarshalIndent

```go
func MarshalIndent(v interface{}, prefix, indent string) ([]byte, error)
```

Produce la misma salida segura para HTML que `Marshal` pero aplica indentación para una salida legible por humanos. Primero llama a `Marshal` para obtener el JSON compacto, luego aplica `json.Indent` con las cadenas de prefijo e indentación especificadas.

```go
data, err := r4.MarshalIndent(patient, "", "  ")
if err != nil {
    log.Fatal(err)
}
fmt.Println(string(data))
```

Salida:

```json
{
  "resourceType": "Patient",
  "id": "narrative-example",
  "text": {
    "status": "generated",
    "div": "<div xmlns=\"http://www.w3.org/1999/xhtml\"><p>Hello</p></div>"
  }
}
```

## Detalles de Implementación

La función `Marshal` es directa pero maneja un caso especial importante. El método `json.Encoder.Encode()` agrega un carácter de salto de línea al final, lo cual no es deseable para salida de bytes sin procesar. La función elimina este salto de línea:

```go
func Marshal(v interface{}) ([]byte, error) {
    var buf bytes.Buffer
    enc := json.NewEncoder(&buf)
    enc.SetEscapeHTML(false)
    if err := enc.Encode(v); err != nil {
        return nil, err
    }
    b := buf.Bytes()
    if len(b) > 0 && b[len(b)-1] == '\n' {
        b = b[:len(b)-1]
    }
    return b, nil
}
```

## Cuándo Usar Cada Función

| Escenario | Función Recomendada |
|-----------|---------------------|
| Recurso FHIR con contenido narrativo | `r4.Marshal()` |
| Recurso FHIR para respuesta de API | `r4.Marshal()` |
| Depuración rápida / registro | `json.Marshal()` o `r4.MarshalIndent()` |
| Almacenamiento de recursos FHIR | `r4.Marshal()` |
| Tipos no FHIR (structs generales de Go) | `json.Marshal()` |

{{< callout type="info" >}}
La función `r4.Marshal()` funciona con cualquier valor de Go, no solo con recursos FHIR. Sin embargo, su propósito principal es manejar el problema del escape HTML en contenido narrativo FHIR. Si tus recursos nunca contienen XHTML narrativo, `json.Marshal()` producirá una salida funcionalmente equivalente.
{{< /callout >}}

## Uso con Handlers HTTP

Al servir recursos FHIR por HTTP, usa `r4.Marshal()` para asegurar que el cuerpo de la respuesta contenga HTML sin escapar:

```go
func handlePatient(w http.ResponseWriter, r *http.Request) {
    patient := buildPatientWithNarrative()

    data, err := r4.Marshal(patient)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/fhir+json")
    w.Write(data)
}
```

Alternativamente, si prefieres transmitir la respuesta en streaming, puedes configurar el codificador directamente:

```go
func handlePatient(w http.ResponseWriter, r *http.Request) {
    patient := buildPatientWithNarrative()

    w.Header().Set("Content-Type", "application/fhir+json")
    enc := json.NewEncoder(w)
    enc.SetEscapeHTML(false)
    enc.Encode(patient)
}
```

## Disponibilidad por Versión

Las funciones `Marshal` y `MarshalIndent` están disponibles en todos los paquetes de versión:

- `r4.Marshal()` / `r4.MarshalIndent()` para FHIR R4
- `r4b.Marshal()` / `r4b.MarshalIndent()` para FHIR R4B
- `r5.Marshal()` / `r5.MarshalIndent()` para FHIR R5

Todas las implementaciones tienen un comportamiento idéntico.

---
title: "Serialización JSON"
linkTitle: "Serialización JSON"
description: "Compatibilidad estándar con encoding/json para la serialización y deserialización de recursos FHIR."
weight: 1
---

Todos los structs de recursos y tipos de datos en la biblioteca `gofhir/models` implementan las interfaces estándar `json.Marshaler` y `json.Unmarshaler` de Go. Esto significa que puedes usar `encoding/json` directamente con cualquier tipo FHIR, y la biblioteca se integra perfectamente con código Go existente, handlers HTTP y bibliotecas de terceros que esperan el comportamiento estándar de marshaling JSON.

## MarshalJSON y UnmarshalJSON

Cada struct de recurso generado (como `Patient`, `Observation`, `Bundle`) implementa métodos personalizados `MarshalJSON()` y `UnmarshalJSON()`. Estos métodos manejan aspectos específicos de FHIR como:

- Serialización del campo discriminador `resourceType`
- Marshaling de recursos contenidos polimórficos
- Manejo del patrón de tipo choice `value[x]`
- Codificación de elementos de extensión de primitivos (`_fieldName`)

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"

    "github.com/gofhir/models/r4"
)

func ptrTo[T any](v T) *T {
    return &v
}

func main() {
    // Create a Patient resource
    patient := &r4.Patient{
        ResourceType: "Patient",
        Id:           ptrTo("example-123"),
        Active:       ptrTo(true),
        Name: []r4.HumanName{
            {
                Use:    ptrTo(r4.NameUseOfficial),
                Family: ptrTo("Smith"),
                Given:  []string{"John", "Michael"},
            },
        },
        Gender:    ptrTo(r4.AdministrativeGenderMale),
        BirthDate: ptrTo("1990-01-15"),
    }

    // Marshal to JSON using the standard library
    data, err := json.Marshal(patient)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(string(data))
}
```

La salida será una representación JSON compacta conforme a la especificación JSON de FHIR:

```json
{"resourceType":"Patient","id":"example-123","active":true,"name":[{"use":"official","family":"Smith","given":["John","Michael"]}],"gender":"male","birthDate":"1990-01-15"}
```

## Etiquetas JSON y omitempty

Todos los campos de los structs usan etiquetas `json` apropiadas con `omitempty` para campos opcionales. Esto asegura que los campos ausentes (nil) se omitan de la salida JSON, produciendo una salida limpia y conforme a la especificación.

```go
type Patient struct {
    ResourceType string              `json:"resourceType"`
    Id           *string             `json:"id,omitempty"`
    Meta         *Meta               `json:"meta,omitempty"`
    Active       *bool               `json:"active,omitempty"`
    Name         []HumanName         `json:"name,omitempty"`
    Gender       *AdministrativeGender `json:"gender,omitempty"`
    BirthDate    *string             `json:"birthDate,omitempty"`
    // ... additional fields
}
```

Los campos requeridos como `ResourceType` no usan `omitempty`, asegurando que siempre estén presentes en la salida serializada.

## Fidelidad de Ida y Vuelta

La biblioteca busca fidelidad de ida y vuelta (round-trip): serializar un recurso a JSON y luego deserializarlo de vuelta produce un struct idéntico. Esto es crítico para sistemas FHIR que necesitan almacenar y recuperar recursos sin pérdida de datos.

Medido sobre los corpus oficiales de ejemplos, 8683 de 8758 archivos JSON (99,1 %) sobreviven un round-trip sin cambios. Clasificando los 75 que no, por todas las diferencias presentes y no solo por la primera:

| Defecto | Archivos | ¿Se detecta? | Efecto |
|---|---:|---|---|
| `Bundle.issues` emitido como `null` en R5 | 37 | silencioso | El campo es una interfaz, así que no recibe el `omitempty` que sí reciben punteros y arrays. Todo Bundle de R5 gana `"issues": null`, que no es FHIR válido |
| `integer64` de R5 tipado como `int64` | 20 | **error** | La especificación exige que `integer64` viaje como **string** en JSON, porque JSON solo garantiza 53 bits de precisión. Los documentos afectados fallan al parsear en lugar de perder datos en silencio |
| `null` en un array de primitivos pasa a `""` | 13 | silencioso | `["2020-01-01", null]` vuelve como `["2020-01-01", ""]`. El `null` posicional que FHIR usa para alinear las extensiones `_campo` no es representable, así que un valor ausente se convierte en uno vacío |
| Extensiones sobre primitivos descartadas | 6 | silencioso | Allí donde el campo compañero no existe: primitivos dentro de backbone elements (`_text` en `Questionnaire.item`) y algunos campos a nivel de recurso (`_base`) |
| No es un recurso FHIR | 1 | error | `package-min-ver.json` es un manifiesto de paquete sin `resourceType`; ruido del corpus, no un defecto |

Las cifras se solapan ligeramente, porque dos archivos llevan más de un defecto. Regenéralas con `cd conformance && go test . -update-known` y lee las listas.

Las extensiones sobre primitivos a nivel de recurso y de datatype sobreviven en su mayoría, pero no universalmente: no las consideres garantizadas hasta que estas listas se reduzcan.

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"

    "github.com/gofhir/models/r4"
)

func ptrTo[T any](v T) *T {
    return &v
}

func main() {
    // Create the original resource
    original := &r4.Patient{
        ResourceType: "Patient",
        Id:           ptrTo("123"),
        Active:       ptrTo(true),
        Gender:       ptrTo(r4.AdministrativeGenderFemale),
    }

    // Marshal to JSON
    data, err := json.Marshal(original)
    if err != nil {
        log.Fatal(err)
    }

    // Unmarshal back to a struct
    var decoded r4.Patient
    if err := json.Unmarshal(data, &decoded); err != nil {
        log.Fatal(err)
    }

    // Verify round-trip fidelity
    fmt.Println(decoded.ResourceType)    // "Patient"
    fmt.Println(*decoded.Id)             // "123"
    fmt.Println(*decoded.Active)         // true
    fmt.Println(*decoded.Gender)         // "female"
}
```

## Salida Indentada

Para depuración o salida legible por humanos, usa `json.MarshalIndent()`:

```go
data, err := json.MarshalIndent(patient, "", "  ")
if err != nil {
    log.Fatal(err)
}
fmt.Println(string(data))
```

Esto produce una salida formateada:

```json
{
  "resourceType": "Patient",
  "id": "123",
  "active": true,
  "gender": "female"
}
```

## Deserialización desde Fuentes Externas

Al recibir JSON FHIR desde una API externa o archivo, deserializa directamente en el struct destino:

```go
jsonData := []byte(`{
    "resourceType": "Patient",
    "id": "external-1",
    "name": [
        {
            "use": "official",
            "family": "Doe",
            "given": ["Jane"]
        }
    ],
    "gender": "female",
    "birthDate": "1985-03-22"
}`)

var patient r4.Patient
if err := json.Unmarshal(jsonData, &patient); err != nil {
    log.Fatal(err)
}

fmt.Println(*patient.Id)             // "external-1"
fmt.Println(*patient.Name[0].Family) // "Doe"
fmt.Println(*patient.Gender)         // "female"
```

## Integración con net/http

Debido a que los structs implementan las interfaces estándar de marshaling, funcionan directamente con `json.NewEncoder` y `json.NewDecoder` para handlers HTTP:

```go
func handleGetPatient(w http.ResponseWriter, r *http.Request) {
    patient := &r4.Patient{
        ResourceType: "Patient",
        Id:           ptrTo("http-example"),
    }

    w.Header().Set("Content-Type", "application/fhir+json")
    json.NewEncoder(w).Encode(patient)
}

func handleCreatePatient(w http.ResponseWriter, r *http.Request) {
    var patient r4.Patient
    if err := json.NewDecoder(r.Body).Decode(&patient); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    // Process the patient...
}
```

{{< callout type="info" >}}
Al servir recursos FHIR por HTTP, considera usar `r4.Marshal()` en lugar de `json.NewEncoder` para evitar el escape de HTML en contenido narrativo. Consulta la página de [Marshal Personalizado](../custom-marshal) para más detalles.
{{< /callout >}}

## Recursos Contenidos

La biblioteca maneja recursos contenidos polimórficos durante la serialización JSON. Los recursos contenidos se almacenan como `[]Resource` (un slice de interfaces) y se serializan correctamente con su discriminador `resourceType`:

```go
patient := &r4.Patient{
    ResourceType: "Patient",
    Id:           ptrTo("with-contained"),
    Contained: []r4.Resource{
        &r4.Organization{
            ResourceType: "Organization",
            Id:           ptrTo("org-1"),
            Name:         ptrTo("Example Hospital"),
        },
    },
    ManagingOrganization: &r4.Reference{
        Reference: ptrTo("#org-1"),
    },
}

data, _ := json.MarshalIndent(patient, "", "  ")
fmt.Println(string(data))
```

Durante la deserialización, el método `UnmarshalJSON` lee el campo `resourceType` de cada entrada contenida y crea el tipo concreto apropiado.

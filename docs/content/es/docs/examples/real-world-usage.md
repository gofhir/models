---
title: "Uso en el Mundo Real"
linkTitle: "Uso en el Mundo Real"
description: "Patrones de produccion para servidores FHIR, clientes de API, conversion de formatos y enrutamiento de recursos type-safe."
weight: 2
---

Esta pagina demuestra patrones listos para produccion para usar `gofhir/models` en aplicaciones reales. Todos los ejemplos usan el paquete R4.

## Funcion Auxiliar

```go
func ptrTo[T any](v T) *T {
    return &v
}
```

## 1. Endpoint de Creacion de Recursos en Servidor FHIR

Un handler HTTP que acepta cualquier recurso FHIR, lo valida, asigna un ID y lo almacena:

```go
import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"

    "github.com/google/uuid"
    "github.com/gofhir/models/r4"
)

func handleCreateResource(w http.ResponseWriter, req *http.Request) {
    if req.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    body, err := io.ReadAll(req.Body)
    if err != nil {
        http.Error(w, "Failed to read request body", http.StatusBadRequest)
        return
    }

    // Inspeccionar el tipo de recurso antes de la deserializacion completa
    resourceType, err := r4.GetResourceType(body)
    if err != nil {
        http.Error(w, "Invalid FHIR resource: missing resourceType", http.StatusBadRequest)
        return
    }

    if !r4.IsKnownResourceType(resourceType) {
        http.Error(w, fmt.Sprintf("Unknown resource type: %s", resourceType), http.StatusBadRequest)
        return
    }

    // Deserializar en la struct Go correcta
    resource, err := r4.UnmarshalResource(body)
    if err != nil {
        http.Error(w, fmt.Sprintf("Failed to parse %s: %v", resourceType, err), http.StatusBadRequest)
        return
    }

    // Asignar un ID generado por el servidor
    resource.SetId(uuid.New().String())

    // Establecer metadatos
    resource.SetMeta(&r4.Meta{
        VersionId:   ptrTo("1"),
        LastUpdated: ptrTo("2024-06-15T12:00:00Z"),
    })

    // Almacenar el recurso (especifico de la implementacion)
    if err := store.Save(resource); err != nil {
        http.Error(w, "Failed to store resource", http.StatusInternalServerError)
        return
    }

    // Devolver el recurso creado
    responseData, err := r4.Marshal(resource)
    if err != nil {
        http.Error(w, "Failed to serialize response", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/fhir+json")
    w.Header().Set("Location", fmt.Sprintf("/%s/%s", resourceType, *resource.GetId()))
    w.WriteHeader(http.StatusCreated)
    w.Write(responseData)
}
```

## 2. Deserializando un Bundle desde una Respuesta de API

Parseando un Bundle de resultados de busqueda desde una API de servidor FHIR y extrayendo recursos individuales:

```go
import (
    "encoding/json"
    "fmt"
    "net/http"

    "github.com/gofhir/models/r4"
)

func fetchPatients(baseURL string) ([]*r4.Patient, error) {
    resp, err := http.Get(baseURL + "/Patient?_count=50")
    if err != nil {
        return nil, fmt.Errorf("request failed: %w", err)
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read response: %w", err)
    }

    // Deserializar como un Bundle
    resource, err := r4.UnmarshalResource(body)
    if err != nil {
        return nil, fmt.Errorf("failed to parse Bundle: %w", err)
    }

    bundle, ok := resource.(*r4.Bundle)
    if !ok {
        return nil, fmt.Errorf("expected Bundle, got %s", resource.GetResourceType())
    }

    // Extraer recursos Patient de las entradas del Bundle
    var patients []*r4.Patient
    for _, entry := range bundle.Entry {
        if entry.Resource == nil {
            continue
        }
        if patient, ok := entry.Resource.(*r4.Patient); ok {
            patients = append(patients, patient)
        }
    }

    fmt.Printf("Found %d patients (total: %d)\n", len(patients), safeTotal(bundle.Total))
    return patients, nil
}

func safeTotal(t *uint32) uint32 {
    if t != nil {
        return *t
    }
    return 0
}
```

## 3. Convirtiendo entre Formato JSON y XML

La biblioteca soporta tanto FHIR JSON como FHIR XML. Puedes convertir entre formatos deserializando de uno y serializando al otro:

```go
import (
    "encoding/json"
    "fmt"

    "github.com/gofhir/models/r4"
)

// Conversion de JSON a XML
func jsonToXML(jsonData []byte) ([]byte, error) {
    // Parsear JSON en un Resource
    resource, err := r4.UnmarshalResource(jsonData)
    if err != nil {
        return nil, fmt.Errorf("failed to parse JSON: %w", err)
    }

    // Serializar a XML
    xmlData, err := r4.MarshalResourceXMLIndent(resource, "", "  ")
    if err != nil {
        return nil, fmt.Errorf("failed to marshal XML: %w", err)
    }

    return xmlData, nil
}

// Conversion de XML a JSON
func xmlToJSON(xmlData []byte) ([]byte, error) {
    // Parsear XML en un Resource
    resource, err := r4.UnmarshalResourceXML(xmlData)
    if err != nil {
        return nil, fmt.Errorf("failed to parse XML: %w", err)
    }

    // Serializar a JSON (HTML-safe)
    jsonData, err := r4.MarshalIndent(resource, "", "  ")
    if err != nil {
        return nil, fmt.Errorf("failed to marshal JSON: %w", err)
    }

    return jsonData, nil
}
```

Ejemplo de uso:

```go
jsonInput := []byte(`{
    "resourceType": "Patient",
    "id": "example",
    "name": [{"family": "Smith", "given": ["John"]}]
}`)

xmlOutput, err := jsonToXML(jsonInput)
if err != nil {
    log.Fatal(err)
}
fmt.Println(string(xmlOutput))
// <?xml version="1.0" encoding="UTF-8"?>
// <Patient xmlns="http://hl7.org/fhir">
//   <id value="example"/>
//   <name>
//     <family value="Smith"/>
//     <given value="John"/>
//   </name>
// </Patient>
```

## 4. Enrutamiento de Recursos Type-Safe Usando el Registro

Un router de recursos que despacha a handlers especificos por tipo usando el registro:

```go
import (
    "fmt"
    "net/http"

    "github.com/gofhir/models/r4"
)

type ResourceHandler func(r4.Resource) error

type FHIRRouter struct {
    handlers map[string]ResourceHandler
}

func NewFHIRRouter() *FHIRRouter {
    return &FHIRRouter{
        handlers: make(map[string]ResourceHandler),
    }
}

func (router *FHIRRouter) Handle(resourceType string, handler ResourceHandler) {
    if !r4.IsKnownResourceType(resourceType) {
        panic(fmt.Sprintf("unknown resource type: %s", resourceType))
    }
    router.handlers[resourceType] = handler
}

func (router *FHIRRouter) Dispatch(data []byte) error {
    resource, err := r4.UnmarshalResource(data)
    if err != nil {
        return fmt.Errorf("failed to unmarshal: %w", err)
    }

    handler, ok := router.handlers[resource.GetResourceType()]
    if !ok {
        return fmt.Errorf("no handler registered for %s", resource.GetResourceType())
    }

    return handler(resource)
}
```

Registrar handlers para tipos de recurso especificos:

```go
router := NewFHIRRouter()

router.Handle("Patient", func(res r4.Resource) error {
    patient := res.(*r4.Patient)
    fmt.Printf("Processing patient: %s\n", safeName(patient))
    // Validar, almacenar, transformar, etc.
    return nil
})

router.Handle("Observation", func(res r4.Resource) error {
    obs := res.(*r4.Observation)
    fmt.Printf("Processing observation: %s\n", safeCode(obs))
    // Procesar la observacion
    return nil
})

// Despachar recursos entrantes
err := router.Dispatch(incomingJSON)
```

### Listando Todos los Tipos de Recurso Soportados

Puedes usar `AllResourceTypes` para auto-registrar handlers por defecto o generar documentacion de API:

```go
import (
    "sort"
    "github.com/gofhir/models/r4"
)

// Generar rutas OpenAPI para todos los tipos de recurso
types := r4.AllResourceTypes()
sort.Strings(types)

for _, rt := range types {
    fmt.Printf("/%s:\n", rt)
    fmt.Printf("  get:\n    summary: Search %s resources\n", rt)
    fmt.Printf("  post:\n    summary: Create a new %s\n", rt)
}
```

{{< callout type="info" >}}
Las funciones del registro usan un mapa de fabrica interno que se puebla en tiempo de compilacion. No hay ningun paso de registro en tiempo de ejecucion y no se requiere inicializacion mas alla de importar el paquete. Todas las funciones del registro son seguras para uso concurrente.
{{< /callout >}}

## Combinando Todo

Una fachada FHIR minima que combina varios patrones:

```go
import (
    "fmt"
    "io"
    "net/http"

    "github.com/gofhir/models/r4"
)

func main() {
    mux := http.NewServeMux()

    // Aceptar cualquier recurso FHIR via POST
    mux.HandleFunc("/fhir/", func(w http.ResponseWriter, req *http.Request) {
        body, _ := io.ReadAll(req.Body)

        // Parsear content type para determinar el formato
        contentType := req.Header.Get("Content-Type")

        var resource r4.Resource
        var err error

        switch contentType {
        case "application/fhir+xml":
            resource, err = r4.UnmarshalResourceXML(body)
        default:
            resource, err = r4.UnmarshalResource(body)
        }

        if err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

        // Responder en el formato solicitado
        accept := req.Header.Get("Accept")
        switch accept {
        case "application/fhir+xml":
            data, _ := r4.MarshalResourceXMLIndent(resource, "", "  ")
            w.Header().Set("Content-Type", "application/fhir+xml")
            w.Write(data)
        default:
            data, _ := r4.Marshal(resource)
            w.Header().Set("Content-Type", "application/fhir+json")
            w.Write(data)
        }
    })

    fmt.Println("FHIR server listening on :8080")
    http.ListenAndServe(":8080", mux)
}
```

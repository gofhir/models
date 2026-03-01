---
title: "Precisión Decimal"
linkTitle: "Precisión Decimal"
description: "Tipo Decimal personalizado para preservación de precisión conforme a FHIR en valores numéricos."
weight: 4
---

La especificación FHIR requiere que los valores decimales preserven su precisión textual original. Por ejemplo, un valor transmitido como `1.50` debe almacenarse y retransmitirse como `1.50`, no como `1.5` o `1.500`. El tipo nativo `float64` de Go no puede garantizar esto, por lo que la biblioteca `gofhir/models` proporciona un tipo `Decimal` personalizado que almacena la representación de cadena exacta mientras soporta operaciones numéricas.

## El Problema

El tipo `float64` de Go normaliza los valores numéricos, perdiendo ceros finales y potencialmente alterando la precisión:

```go
// float64 loses trailing zeros
f := 1.50
fmt.Println(f) // "1.5" -- trailing zero is lost

// JSON marshaling of float64 also loses precision
data, _ := json.Marshal(f)
fmt.Println(string(data)) // "1.5"
```

En FHIR, la precisión de un valor decimal tiene significado clínico. Un resultado de laboratorio reportado como `7.0` mg/dL implica precisión de un decimal, mientras que `7.00` implica precisión de dos decimales. Perder esta distinción viola la especificación FHIR y puede causar fallos de validación con servidores FHIR.

## El Tipo Decimal

El tipo `Decimal` se define en `decimal.go`:

```go
type Decimal struct {
    value string
}
```

Almacena el valor numérico como su representación de cadena original. Todos los métodos de construcción validan que la cadena sea un número decimal válido (no NaN ni Infinity), preservando al mismo tiempo el texto de entrada exacto.

## Métodos de Construcción

### NewDecimalFromString

Crea un `Decimal` a partir de una representación de cadena, preservando el texto exacto. Retorna un error si la cadena no es un número decimal válido.

```go
d, err := r4.NewDecimalFromString("1.50")
if err != nil {
    log.Fatal(err)
}
fmt.Println(d.String()) // "1.50" -- precision preserved
```

Este es el constructor preferido cuando tienes un valor de cadena de una fuente externa (JSON, base de datos, entrada de usuario) y quieres preservar su representación exacta.

### MustDecimal

Crea un `Decimal` a partir de una cadena, entrando en pánico si la entrada es inválida. Úsalo solo para valores constantes en tiempo de compilación que se sabe que son válidos.

```go
d := r4.MustDecimal("1.50")
fmt.Println(d.String()) // "1.50"

// Panics on invalid input
// d := r4.MustDecimal("not-a-number") // panic!
```

### NewDecimalFromFloat64

Crea un `Decimal` a partir de un valor `float64`. Ten en cuenta que la precisión puede perderse durante la conversión de float a cadena.

```go
d := r4.NewDecimalFromFloat64(1.5)
fmt.Println(d.String()) // "1.5" -- trailing zero lost (float64 normalization)

d2 := r4.NewDecimalFromFloat64(72.5)
fmt.Println(d2.String()) // "72.5"
```

{{< callout type="info" >}}
Usa `NewDecimalFromString` en lugar de `NewDecimalFromFloat64` cuando la preservación de precisión sea importante. El constructor float64 es una conveniencia para casos donde la pérdida de precisión es aceptable.
{{< /callout >}}

### NewDecimalFromInt

Crea un `Decimal` a partir de un valor `int`:

```go
d := r4.NewDecimalFromInt(100)
fmt.Println(d.String()) // "100"
```

### NewDecimalFromInt64

Crea un `Decimal` a partir de un valor `int64`:

```go
d := r4.NewDecimalFromInt64(9999999999)
fmt.Println(d.String()) // "9999999999"
```

## Métodos de Acceso

### String

Retorna la representación textual exacta del decimal:

```go
d := r4.MustDecimal("3.14159")
fmt.Println(d.String()) // "3.14159"
```

### Float64

Convierte el decimal a `float64` para operaciones numéricas. La precisión puede perderse en la conversión:

```go
d := r4.MustDecimal("72.50")
f := d.Float64()
fmt.Println(f) // 72.5
```

### IsZero

Retorna `true` si el valor decimal es cero o está vacío:

```go
d1 := r4.MustDecimal("0")
d2 := r4.MustDecimal("0.00")
d3 := r4.MustDecimal("1.5")

fmt.Println(d1.IsZero()) // true
fmt.Println(d2.IsZero()) // true
fmt.Println(d3.IsZero()) // false
```

### Equal

Compara dos valores `Decimal` numéricamente (no textualmente). Esto significa que `"1.0"` y `"1.00"` se consideran iguales:

```go
d1 := r4.MustDecimal("1.0")
d2 := r4.MustDecimal("1.00")
d3 := r4.MustDecimal("2.0")

fmt.Println(d1.Equal(*d2)) // true  (same numeric value)
fmt.Println(d1.Equal(*d3)) // false
```

Ten en cuenta que `Equal` compara mediante conversión a `float64`, por lo que verifica igualdad numérica en lugar de identidad textual.

## Marshaling JSON

El tipo `Decimal` implementa `json.Marshaler` y `json.Unmarshaler` para producir salida JSON conforme a la especificación.

### Marshaling

El `Decimal` se serializa como un número JSON sin comillas (no una cadena entrecomillada), preservando la precisión textual original:

```go
d := r4.MustDecimal("1.50")
data, _ := json.Marshal(d)
fmt.Println(string(data)) // 1.50 (bare number, not "1.50")
```

Esto es crítico para el cumplimiento de FHIR. La salida es `1.50`, no `1.5` (lo cual perdería precisión) ni `"1.50"` (lo cual sería una cadena en lugar de un número).

### Unmarshaling

El `Decimal` acepta un número JSON sin comillas y almacena su representación exacta en bytes:

```go
var d r4.Decimal
json.Unmarshal([]byte("1.50"), &d)
fmt.Println(d.String()) // "1.50" -- precision preserved
```

También maneja números entrecomillados para compatibilidad con algunos servidores FHIR que codifican decimales como cadenas:

```go
var d r4.Decimal
json.Unmarshal([]byte(`"1.50"`), &d)
fmt.Println(d.String()) // "1.50"
```

## Uso en Recursos

El tipo `Decimal` se usa donde FHIR define un elemento `decimal`. La ocurrencia más común es en `Quantity` y sus especializaciones (`Age`, `Distance`, `Duration`, `Count`):

```go
observation := &r4.Observation{
    ResourceType: "Observation",
    Id:           ptrTo("weight-1"),
    Status:       ptrTo(r4.ObservationStatusFinal),
    Code: &r4.CodeableConcept{
        Coding: []r4.Coding{
            {
                System:  ptrTo("http://loinc.org"),
                Code:    ptrTo("29463-7"),
                Display: ptrTo("Body weight"),
            },
        },
    },
    ValueQuantity: &r4.Quantity{
        Value:  r4.MustDecimal("72.50"),
        Unit:   ptrTo("kg"),
        System: ptrTo("http://unitsofmeasure.org"),
        Code:   ptrTo("kg"),
    },
}

// Marshal preserves the "72.50" precision
data, _ := r4.Marshal(observation)
fmt.Println(string(data))
// ... "value":72.50 ...
```

## Preservación de Ida y Vuelta

El tipo `Decimal` garantiza la preservación de precisión en ida y vuelta (round-trip) a través de JSON:

```go
// Original value with trailing zero
original := r4.MustDecimal("98.60")

// Marshal to JSON
data, _ := json.Marshal(original)
fmt.Println(string(data)) // 98.60

// Unmarshal back
var restored r4.Decimal
json.Unmarshal(data, &restored)
fmt.Println(restored.String()) // "98.60" -- identical to original
```

## Comportamiento del Decimal Vacío

Un `Decimal` vacío (struct con valor cero) se serializa como `0`:

```go
var d r4.Decimal
data, _ := json.Marshal(d)
fmt.Println(string(data)) // 0
```

El método `IsZero` retorna `true` tanto para decimales vacíos como para ceros numéricos:

```go
var empty r4.Decimal
fmt.Println(empty.IsZero()) // true

zero := r4.MustDecimal("0")
fmt.Println(zero.IsZero()) // true
```

{{< callout type="info" >}}
El tipo `Decimal` almacena valores como cadenas internamente, lo que significa que las operaciones aritméticas no se soportan directamente. Para cálculos, convierte a `float64` usando el método `Float64()`, realiza la aritmética y luego crea un nuevo `Decimal` a partir del resultado. Ten en cuenta que esto puede afectar la precisión.
{{< /callout >}}

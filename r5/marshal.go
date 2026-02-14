package r5

import (
	"bytes"
	"encoding/json"
)

// Marshal serializes a FHIR resource to JSON without HTML escaping.
//
// Go's standard json.Marshal escapes HTML characters (<, >, &) in strings,
// which breaks FHIR narrative content in text.div fields that must contain
// valid XHTML. This function uses json.Encoder with SetEscapeHTML(false)
// to preserve HTML content as required by the FHIR specification.
//
// Use this function instead of json.Marshal when serializing FHIR resources
// that may contain narrative HTML.
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

// MarshalIndent is like Marshal but applies Indent to format the output.
func MarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	b, err := Marshal(v)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := json.Indent(&buf, b, prefix, indent); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

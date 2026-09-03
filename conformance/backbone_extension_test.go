package conformance

// A primitive nested in a backbone element carries extensions in a _field
// companion, exactly as one at the root of a resource does. The generated
// backbones had no companion fields at all, so every such extension was dropped
// on decode — four of the corpus's seven JSON failures, in
// ElementDefinition.type.profile, CodeSystem.property.code and
// Availability.availableTime.availableEndTime.
//
// The corpus cannot guard this on the XML side: it compares our output against our
// own output, so a codec that consistently discards the same field round-trips
// perfectly. These tests compare against the input instead, and check that JSON and
// XML agree — an asymmetry would mean data survives one format and not the other.

import (
	"encoding/json"
	"encoding/xml"
	"testing"

	"github.com/gofhir/models/r4/v2"
	"github.com/gofhir/models/r4b/v2"
	"github.com/gofhir/models/r5/v2"
)

func TestBackboneScalarExtensionSurvivesJSON(t *testing.T) {
	// CodeSystem.property.code — a scalar primitive inside a backbone.
	const input = `{
		"resourceType": "CodeSystem",
		"status": "active",
		"content": "complete",
		"property": [{
			"uri": "http://example.org/p",
			"_code": {
				"extension": [{
					"url": "http://hl7.org/fhir/StructureDefinition/data-absent-reason",
					"valueCode": "unknown"
				}]
			}
		}]
	}`

	var cs r4.CodeSystem
	if err := json.Unmarshal([]byte(input), &cs); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(cs.Property) != 1 {
		t.Fatalf("got %d properties, want 1", len(cs.Property))
	}
	ext := cs.Property[0].CodeExt
	if ext == nil {
		t.Fatal("CodeSystem.property[0]._code was dropped on decode")
	}
	if len(ext.Extension) != 1 || ext.Extension[0].ValueCode == nil ||
		*ext.Extension[0].ValueCode != "unknown" {
		t.Fatalf("the extension survived but its content did not: %+v", ext.Extension)
	}

	// And it must come back out.
	out, err := json.Marshal(&cs)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var reparsed r4.CodeSystem
	if err := json.Unmarshal(out, &reparsed); err != nil {
		t.Fatalf("reparse: %v", err)
	}
	if reparsed.Property[0].CodeExt == nil {
		t.Errorf("the extension was decoded but not written back:\n%s", out)
	}
}

func TestBackboneScalarExtensionSurvivesXML(t *testing.T) {
	// The same value expressed in XML: the extension lives inside the <code>
	// element, whose value attribute is absent.
	const input = `<CodeSystem xmlns="http://hl7.org/fhir">
		<status value="active"/>
		<content value="complete"/>
		<property>
			<uri value="http://example.org/p"/>
			<code>
				<extension url="http://hl7.org/fhir/StructureDefinition/data-absent-reason">
					<valueCode value="unknown"/>
				</extension>
			</code>
		</property>
	</CodeSystem>`

	var cs r4.CodeSystem
	if err := xml.Unmarshal([]byte(input), &cs); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(cs.Property) != 1 {
		t.Fatalf("got %d properties, want 1", len(cs.Property))
	}
	if cs.Property[0].CodeExt == nil {
		t.Fatal("CodeSystem.property[0] <code> extension was dropped on XML decode")
	}

	out, err := xml.Marshal(&cs)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var reparsed r4.CodeSystem
	if err := xml.Unmarshal(out, &reparsed); err != nil {
		t.Fatalf("reparse: %v", err)
	}
	if reparsed.Property[0].CodeExt == nil {
		t.Errorf("the extension was decoded but not written back to XML:\n%s", out)
	}
}

func TestBackboneArrayExtensionKeepsItsPosition(t *testing.T) {
	// ElementDefinition.type.profile is a repeating primitive inside a backbone of
	// a datatype. The value and extension arrays are parallel by position, so a
	// decoder that appends only values shifts every later extension onto the wrong
	// element — which is worse than dropping them, because it is silent.
	const input = `{
		"resourceType": "StructureDefinition",
		"status": "active",
		"kind": "resource",
		"abstract": false,
		"type": "Patient",
		"url": "http://example.org/sd",
		"name": "Example",
		"differential": {
			"element": [{
				"path": "Patient.name",
				"type": [{
					"code": "HumanName",
					"profile": ["http://example.org/a", null, "http://example.org/c"],
					"_profile": [null, {
						"extension": [{
							"url": "http://hl7.org/fhir/StructureDefinition/data-absent-reason",
							"valueCode": "masked"
						}]
					}, null]
				}]
			}]
		}
	}`

	var sd r4.StructureDefinition
	if err := json.Unmarshal([]byte(input), &sd); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	typ := sd.Differential.Element[0].Type[0]

	if len(typ.Profile) != 3 {
		t.Fatalf("got %d profiles, want 3", len(typ.Profile))
	}
	if typ.Profile[1] != nil {
		t.Errorf("profile[1] was null in the source and must stay nil, got %q", *typ.Profile[1])
	}
	if len(typ.ProfileExt) != 3 {
		t.Fatalf("got %d profile extensions, want 3 — the parallel array lost its shape", len(typ.ProfileExt))
	}
	if typ.ProfileExt[0] != nil || typ.ProfileExt[2] != nil {
		t.Error("the extension landed in the wrong slot; positions 0 and 2 had none")
	}
	if typ.ProfileExt[1] == nil {
		t.Fatal("the extension on profile[1] was dropped")
	}
	if got := typ.ProfileExt[1].Extension; len(got) != 1 || got[0].ValueCode == nil || *got[0].ValueCode != "masked" {
		t.Errorf("extension content changed: %+v", got)
	}
}

// TestBackboneExtensionsAcrossVersions is the same scalar check in r4b and r5,
// because the templates are shared but the specifications are not — a field can be
// a backbone in one version and not in another.
func TestBackboneExtensionsAcrossVersions(t *testing.T) {
	const input = `{
		"resourceType": "CodeSystem",
		"status": "active",
		"content": "complete",
		"property": [{
			"uri": "http://example.org/p",
			"_code": {"extension": [{
				"url": "http://hl7.org/fhir/StructureDefinition/data-absent-reason",
				"valueCode": "unknown"
			}]}
		}]
	}`

	t.Run("r4b", func(t *testing.T) {
		var cs r4b.CodeSystem
		if err := json.Unmarshal([]byte(input), &cs); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if len(cs.Property) != 1 || cs.Property[0].CodeExt == nil {
			t.Fatal("CodeSystem.property[0]._code was dropped")
		}
	})

	t.Run("r5", func(t *testing.T) {
		var cs r5.CodeSystem
		if err := json.Unmarshal([]byte(input), &cs); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if len(cs.Property) != 1 || cs.Property[0].CodeExt == nil {
			t.Fatal("CodeSystem.property[0]._code was dropped")
		}
	})
}

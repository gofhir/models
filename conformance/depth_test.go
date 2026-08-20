package conformance

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

// maxDepthObserved is the deepest JSON nesting seen across all three published
// example corpora, recorded so the depth guard in the JSON dispatcher can be
// justified by data instead of guesswork.
//
// Currently 28, from r4b's structuremap-questionnaire.json — a Questionnaire with
// deeply nested item trees. That is twice as deep as the FHIR definition bundles
// themselves reach (14), which is why this measurement matters: sizing a limit
// from the specs alone would have been optimistic. Questionnaire.item,
// Parameters.parameter.part and GraphDefinition.link.target have no depth ceiling
// in the specification, so real documents can legitimately go deeper still.
//
// The trade-off is asymmetric — too high only mitigates less, too low turns valid
// production data into a hard error — so the enforced limit should sit well above
// this, not just above it.
//
// Update only from TestReportJSONDepth output, and only upwards.
const maxDepthObserved = 28

// TestReportJSONDepth measures how deeply real FHIR documents nest, and asserts
// the corpus stays well under the limit the library enforces.
//
// This exists because a depth limit is a trade-off in one direction only: too high
// and it mitigates less, too low and it turns valid production data into a hard
// error. The number should come from the published examples, not from intuition.
func TestReportJSONDepth(t *testing.T) {
	type observation struct {
		file  string
		depth int
	}

	var (
		deepest   []observation
		anyCorpus bool
	)

	for _, c := range codecs() {
		dir := filepath.Join(examplesDir, c.name, "json")
		files, err := filepath.Glob(filepath.Join(dir, "*.json"))
		if err != nil {
			t.Fatalf("globbing %s: %v", dir, err)
		}
		if len(files) == 0 {
			continue
		}
		anyCorpus = true

		worst := observation{}
		histogram := map[int]int{}
		for _, path := range files {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("reading %s: %v", path, err)
			}
			var v any
			if err := json.Unmarshal(data, &v); err != nil {
				continue // malformed or non-FHIR file; TestRoundTrip reports those
			}
			d := jsonDepth(v)
			histogram[d]++
			if d > worst.depth {
				worst = observation{file: filepath.Base(path), depth: d}
			}
		}

		depths := make([]int, 0, len(histogram))
		for d := range histogram {
			depths = append(depths, d)
		}
		sort.Ints(depths)

		// Report the tail, which is what matters when choosing a limit.
		tail := depths
		if len(tail) > 6 {
			tail = tail[len(tail)-6:]
		}
		t.Logf("%s: %d files, deepest %d (%s)", c.name, len(files), worst.depth, worst.file)
		for _, d := range tail {
			t.Logf("    depth %2d: %d file(s)", d, histogram[d])
		}

		deepest = append(deepest, worst)
	}

	if !anyCorpus {
		t.Skip("no JSON corpus — run scripts/fetch-examples.sh")
	}

	overall := 0
	for _, o := range deepest {
		if o.depth > overall {
			overall = o.depth
		}
	}

	if overall > maxDepthObserved {
		t.Errorf("the corpus now nests %d levels deep, deeper than the recorded %d.\n"+
			"Raise maxDepthObserved and re-check that the dispatcher's depth limit still"+
			" leaves comfortable headroom.", overall, maxDepthObserved)
	}
}

// jsonDepth returns the nesting depth of a decoded JSON value, counting each
// object and array as one level. It matches what a brace-counting guard in the
// dispatcher would see.
func jsonDepth(v any) int {
	switch tv := v.(type) {
	case map[string]any:
		deepest := 0
		for _, child := range tv {
			if d := jsonDepth(child); d > deepest {
				deepest = d
			}
		}
		return deepest + 1
	case []any:
		deepest := 0
		for _, child := range tv {
			if d := jsonDepth(child); d > deepest {
				deepest = d
			}
		}
		return deepest + 1
	default:
		return 0
	}
}

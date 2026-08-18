package conformance

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// -update-known rewrites the known-failure lists from the current run. Use it
// after fixing (or knowingly changing) behavior, and read the diff: every line
// removed is a bug fixed, every line added is a regression being accepted.
var updateKnown = flag.Bool("update-known", false, "rewrite testdata/known_failures/*.txt")

var errNotResource = errors.New("value does not implement the version's Resource interface")

const (
	examplesDir      = "testdata/examples"
	knownFailuresDir = "testdata/known_failures"
)

// failureKind classifies what went wrong, so the lists stay readable and a change
// of failure mode on the same file is visible rather than silently equivalent.
type failureKind string

const (
	failParse     failureKind = "parse"     // the published example could not be read
	failMarshal   failureKind = "marshal"   // parsed, but could not be written back
	failReparse   failureKind = "reparse"   // our own output could not be read back
	failMismatch  failureKind = "mismatch"  // round-trip lost or changed data
	failEmptyRead failureKind = "emptyread" // parsed to nothing, silently
)

type failure struct {
	file string
	kind failureKind
	// detail is deliberately not persisted: it changes with unrelated edits and
	// would make the lists churn. It is only printed in test output.
	detail string
}

func (f failure) line() string { return string(f.kind) + " " + f.file }

// TestRoundTrip is the conformance inventory: it round-trips every published
// example through the library and compares the result against a recorded list of
// known failures.
//
// It fails when a file that used to work stops working (regression) and when a
// file on the known-failure list starts working (progress that must be recorded).
// That makes the lists a ratchet rather than a snapshot nobody maintains.
//
// Skipped entirely when the corpus is absent, so `go test ./...` in a fresh clone
// does not require a 200 MB download. Run scripts/fetch-examples.sh first.
func TestRoundTrip(t *testing.T) {
	for _, c := range codecs() {
		for _, kind := range []string{"json", "xml"} {
			t.Run(c.name+"/"+kind, func(t *testing.T) {
				runCorpus(t, c, kind)
			})
		}
	}
}

func runCorpus(t *testing.T, c codec, kind string) {
	dir := filepath.Join(examplesDir, c.name, kind)
	files, err := filepath.Glob(filepath.Join(dir, "*."+kind))
	if err != nil {
		t.Fatalf("globbing %s: %v", dir, err)
	}
	if len(files) == 0 {
		t.Skipf("no %s corpus in %s — run scripts/fetch-examples.sh", kind, dir)
	}
	sort.Strings(files)

	unmarshal, marshal := c.unmarshalJSON, c.marshalJSON
	if kind == "xml" {
		unmarshal, marshal = c.unmarshalXML, c.marshalXML
	}

	var failures []failure
	for _, path := range files {
		if f, failed := checkFile(path, kind, unmarshal, marshal); failed {
			failures = append(failures, f)
		}
	}

	listPath := filepath.Join(knownFailuresDir, c.name+"-"+kind+".txt")
	if *updateKnown {
		writeKnownFailures(t, listPath, failures, len(files))
		t.Logf("recorded %d/%d failures in %s", len(failures), len(files), listPath)
		return
	}

	compareAgainstKnown(t, listPath, failures, len(files))
}

// checkFile performs parse -> marshal -> reparse -> compare on one example.
func checkFile(path, kind string, unmarshal func([]byte) (any, error), marshal func(any) ([]byte, error)) (failure, bool) {
	name := filepath.Base(path)

	original, err := os.ReadFile(path)
	if err != nil {
		return failure{name, failParse, "reading file: " + err.Error()}, true
	}

	parsed, err := unmarshal(original)
	if err != nil {
		return failure{name, failParse, err.Error()}, true
	}
	if parsed == nil {
		return failure{name, failEmptyRead, "unmarshal returned nil without an error"}, true
	}

	produced, err := marshal(parsed)
	if err != nil {
		return failure{name, failMarshal, err.Error()}, true
	}

	// Re-parse our own output. A library that cannot read what it just wrote is
	// broken in a way no single-direction test would catch.
	reparsed, err := unmarshal(produced)
	if err != nil {
		return failure{name, failReparse, err.Error()}, true
	}

	// Compare the second serialization against the first: stability of our own
	// representation. Comparing directly against the published bytes would flag
	// harmless differences in key order and whitespace on every single file.
	reproduced, err := marshal(reparsed)
	if err != nil {
		return failure{name, failMarshal, "second marshal: " + err.Error()}, true
	}

	if kind == "json" {
		if diff := jsonDiff(produced, reproduced); diff != "" {
			return failure{name, failMismatch, diff}, true
		}
		// Data loss against the source is what actually matters, so also check
		// that every leaf in the original survived. Key order and formatting are
		// ignored; missing or altered values are not.
		if diff := jsonDiff(original, produced); diff != "" {
			return failure{name, failMismatch, diff}, true
		}
		return failure{}, false
	}

	if !bytes.Equal(produced, reproduced) {
		return failure{name, failMismatch, firstByteDiff(produced, reproduced)}, true
	}
	return failure{}, false
}

// jsonDiff compares two JSON documents semantically and returns the first
// difference found, as a path plus the two values. Numbers are compared by their
// literal text, because FHIR decimals must preserve precision (2.00 != 2.0).
func jsonDiff(a, b []byte) string {
	va, err := decodeJSON(a)
	if err != nil {
		return "left side is not valid JSON: " + err.Error()
	}
	vb, err := decodeJSON(b)
	if err != nil {
		return "right side is not valid JSON: " + err.Error()
	}
	return diffValues("", va, vb)
}

func decodeJSON(data []byte) (any, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber() // keep 2.00 distinct from 2.0
	var v any
	if err := dec.Decode(&v); err != nil {
		return nil, err
	}
	return v, nil
}

func diffValues(path string, a, b any) string {
	switch av := a.(type) {
	case map[string]any:
		bv, ok := b.(map[string]any)
		if !ok {
			return fmt.Sprintf("%s: object became %T", pathOr(path), b)
		}
		keys := make([]string, 0, len(av))
		for k := range av {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			bval, present := bv[k]
			if !present {
				return fmt.Sprintf("%s.%s: dropped (was %s)", pathOr(path), k, brief(av[k]))
			}
			if d := diffValues(path+"."+k, av[k], bval); d != "" {
				return d
			}
		}
		for k := range bv {
			if _, present := av[k]; !present {
				return fmt.Sprintf("%s.%s: appeared (%s)", pathOr(path), k, brief(bv[k]))
			}
		}
		return ""

	case []any:
		bv, ok := b.([]any)
		if !ok {
			return fmt.Sprintf("%s: array became %T", pathOr(path), b)
		}
		if len(av) != len(bv) {
			return fmt.Sprintf("%s: length %d became %d", pathOr(path), len(av), len(bv))
		}
		for i := range av {
			if d := diffValues(fmt.Sprintf("%s[%d]", path, i), av[i], bv[i]); d != "" {
				return d
			}
		}
		return ""

	default:
		if fmt.Sprint(a) != fmt.Sprint(b) {
			return fmt.Sprintf("%s: %s became %s", pathOr(path), brief(a), brief(b))
		}
		return ""
	}
}

func pathOr(path string) string {
	if path == "" {
		return "(root)"
	}
	return strings.TrimPrefix(path, ".")
}

// brief renders a value for an error message without dumping a whole subtree.
func brief(v any) string {
	s := fmt.Sprint(v)
	if len(s) > 80 {
		s = s[:77] + "..."
	}
	switch v.(type) {
	case map[string]any:
		return "{object}"
	case []any:
		return "[array]"
	}
	return s
}

func firstByteDiff(a, b []byte) string {
	limit := len(a)
	if len(b) < limit {
		limit = len(b)
	}
	for i := 0; i < limit; i++ {
		if a[i] != b[i] {
			return fmt.Sprintf("first byte difference at offset %d: %q vs %q",
				i, window(a, i), window(b, i))
		}
	}
	return fmt.Sprintf("identical for %d bytes, then lengths differ (%d vs %d)", limit, len(a), len(b))
}

func window(b []byte, at int) string {
	start := at - 20
	if start < 0 {
		start = 0
	}
	end := at + 40
	if end > len(b) {
		end = len(b)
	}
	return string(b[start:end])
}

func writeKnownFailures(t *testing.T, path string, failures []failure, total int) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}

	lines := make([]string, 0, len(failures))
	for _, f := range failures {
		lines = append(lines, f.line())
	}
	sort.Strings(lines)

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "# Known round-trip failures: %d of %d examples.\n", len(failures), total)
	buf.WriteString("#\n")
	buf.WriteString("# Regenerate with: go test ./conformance -update-known\n")
	buf.WriteString("# Every line removed is a bug fixed; every line added is a regression.\n")
	buf.WriteString("#\n")
	buf.WriteString("# Format: <kind> <file>\n")
	buf.WriteString("#   parse     the published example could not be read\n")
	buf.WriteString("#   marshal   parsed, but could not be written back\n")
	buf.WriteString("#   reparse   our own output could not be read back\n")
	buf.WriteString("#   mismatch  round-trip lost or changed data\n")
	buf.WriteString("#   emptyread parsed to nothing, with no error\n")
	buf.WriteString("\n")
	for _, l := range lines {
		buf.WriteString(l)
		buf.WriteString("\n")
	}

	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		t.Fatalf("writing %s: %v", path, err)
	}
}

func readKnownFailures(path string) (map[string]bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	known := map[string]bool{}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		known[line] = true
	}
	return known, nil
}

func compareAgainstKnown(t *testing.T, path string, failures []failure, total int) {
	t.Helper()

	known, err := readKnownFailures(path)
	if err != nil {
		if os.IsNotExist(err) {
			t.Fatalf("no known-failure list at %s.\nCreate it with: go test ./conformance -update-known", path)
		}
		t.Fatalf("reading %s: %v", path, err)
	}

	current := map[string]failure{}
	for _, f := range failures {
		current[f.line()] = f
	}

	// Regressions: failing now, not on the list.
	var regressions []failure
	for line, f := range current {
		if !known[line] {
			regressions = append(regressions, f)
		}
	}

	// Progress: on the list, passing now. Just as important to surface, otherwise
	// the list rots into a pile of stale entries nobody trusts.
	var fixed []string
	for line := range known {
		if _, still := current[line]; !still {
			fixed = append(fixed, line)
		}
	}

	sort.Slice(regressions, func(i, j int) bool { return regressions[i].file < regressions[j].file })
	sort.Strings(fixed)

	t.Logf("%d/%d examples round-trip cleanly (%d known failures)",
		total-len(failures), total, len(failures))

	if len(regressions) > 0 {
		var b strings.Builder
		fmt.Fprintf(&b, "%d example(s) newly failing:\n", len(regressions))
		for i, f := range regressions {
			if i == 15 {
				fmt.Fprintf(&b, "  ... and %d more\n", len(regressions)-15)
				break
			}
			fmt.Fprintf(&b, "  [%s] %s\n      %s\n", f.kind, f.file, f.detail)
		}
		t.Error(b.String())
	}

	if len(fixed) > 0 {
		var b strings.Builder
		fmt.Fprintf(&b, "%d known failure(s) now pass — record the progress with"+
			" `go test ./conformance -update-known`:\n", len(fixed))
		for i, line := range fixed {
			if i == 15 {
				fmt.Fprintf(&b, "  ... and %d more\n", len(fixed)-15)
				break
			}
			fmt.Fprintf(&b, "  %s\n", line)
		}
		t.Error(b.String())
	}
}

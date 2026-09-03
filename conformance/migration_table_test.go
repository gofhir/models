package conformance

// The v1→v2 migration guide carries a table of every renamed code system type.
// A table that drifts is worse than no table: someone follows it, the rename does
// not compile, and they stop trusting the rest of the page.
//
// So it is checked against the generated packages rather than maintained by hand.
// Each row must describe a rename that actually happened in each version it names,
// and no type may disappear without a row explaining where it went.

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

var (
	migrationRow  = regexp.MustCompile("(?m)^\\| `([A-Za-z0-9]+)` \\| `([A-Za-z0-9]+)` \\| ([a-z0-9, ]+) \\|$")
	codeSystemDef = regexp.MustCompile(`(?m)^type ([A-Za-z0-9]+) string$`)
)

func TestMigrationGuideRenameTable(t *testing.T) {
	guides := map[string]string{
		"en": filepath.Join("..", "docs", "content", "en", "docs", "migration", "v1-to-v2.md"),
		"es": filepath.Join("..", "docs", "content", "es", "docs", "migration", "v1-to-v2.md"),
	}

	// The types each generated package declares today.
	declared := map[string]map[string]bool{}
	for _, module := range []string{"r4", "r4b", "r5"} {
		source, err := os.ReadFile(filepath.Join("..", module, "codesystems.go"))
		if err != nil {
			t.Fatalf("reading %s/codesystems.go: %v", module, err)
		}
		names := map[string]bool{}
		for _, m := range codeSystemDef.FindAllStringSubmatch(string(source), -1) {
			names[m[1]] = true
		}
		if len(names) < 100 {
			t.Fatalf("only %d types found in %s; the check is not seeing the generated code",
				len(names), module)
		}
		declared[module] = names
	}

	for lang, path := range guides {
		t.Run(lang, func(t *testing.T) {
			guide, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("reading %s: %v", path, err)
			}

			rows := migrationRow.FindAllStringSubmatch(string(guide), -1)
			if len(rows) == 0 {
				t.Fatal("the guide has no rename table; if the section was removed, remove this test with it")
			}

			// Every row must hold in each version it claims.
			listed := map[string]map[string]bool{"r4": {}, "r4b": {}, "r5": {}}
			for _, row := range rows {
				oldName, newName := row[1], row[2]
				for _, module := range strings.Split(row[3], ",") {
					module = strings.TrimSpace(module)
					if declared[module] == nil {
						t.Errorf("row %s → %s names %q, which is not a version", oldName, newName, module)
						continue
					}
					listed[module][oldName] = true
					if declared[module][oldName] {
						t.Errorf("%s: the guide says %s was renamed, but it is still declared",
							module, oldName)
					}
					if !declared[module][newName] {
						t.Errorf("%s: the guide renames %s to %s, which does not exist",
							module, oldName, newName)
					}
				}
			}

			// And no rename may be missing from it. Anything the v1 tag declared
			// that is gone now needs a row, or a reader hits an identifier that
			// vanished with nothing to look it up by.
			for _, module := range []string{"r4", "r4b", "r5"} {
				missing := []string{}
				for _, name := range v1CodeSystemTypes(t, module) {
					if !declared[module][name] && !listed[module][name] {
						missing = append(missing, name)
					}
				}
				if len(missing) > 0 {
					sort.Strings(missing)
					t.Errorf("%s: these types are gone from v2 with no row in the guide: %v",
						module, missing)
				}
			}
		})
	}
}

// v1CodeSystemTypes reads the type names the v1 line declared, from a checked-in
// snapshot rather than from git — the test must work in a source tarball, and
// shelling out to git for a fixture is worse than keeping one.
func v1CodeSystemTypes(t *testing.T, module string) []string {
	t.Helper()

	source, err := os.ReadFile(filepath.Join("testdata", "v1-codesystem-types", module+".txt"))
	if err != nil {
		t.Fatalf("reading the v1 type list for %s: %v", module, err)
	}
	names := []string{}
	for _, line := range strings.Split(string(source), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			names = append(names, line)
		}
	}
	if len(names) < 100 {
		t.Fatalf("only %d v1 types listed for %s; the fixture looks truncated", len(names), module)
	}
	return names
}

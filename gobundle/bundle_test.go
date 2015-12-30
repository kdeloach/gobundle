package gobundle

import (
    "encoding/json"
    "os"
    "path"
    "path/filepath"
    "testing"
)

func TestResolver1(t *testing.T) {
    assertResolvesToPath(t, "./foo", "foo.js")
    assertResolvesToPath(t, "./bar", "bar.js")
    assertResolvesToPath(t, "./baz/baz", "baz/baz.js")
    assertResolvesToPath(t, "query", "node_modules/query/index.js")
    assertResolvesToPath(t, "util", "node_modules/util/util.js")
}

func TestParser1(t *testing.T) {
    assertParseMatch(t, "./foo", []string{"query", "./bar"})
}

func TestGraph(t *testing.T) {
    rootPath := getRootPath(t)
    r := Resolver{Path: rootPath}
    actual := r.graph(rootPath, "./foo.js")
    expected := map[string][][]string {
        "foo.js": {
            []string{"query", "node_modules/query/index.js"},
            []string{"./bar", "bar.js"},
        },
        "bar.js": {
            []string{"util", "node_modules/util/util.js"},
            []string{"./baz/baz.js", "baz/baz.js"},
        },
        "baz/baz.js": {
            []string{"util", "node_modules/util/util.js"},
            []string{"./sibling", "baz/sibling.js"},
        },
        "baz/sibling.js": {
            []string{"./baz.js", "baz/baz.js"},
        },
        "node_modules/query/index.js": {
            []string{"../util", "node_modules/util/util.js"},
        },
        "node_modules/util/util.js": {},
    }

    a, _ := json.Marshal(expected)
    b, _ := json.Marshal(actual.Nodes)
    assertMatch(t, string(a), string(b))

    if actual.EntryFile != "./foo.js" {
        t.Fail()
    }
}

func TestIdMaker(t *testing.T) {
    id := makeIdFunc()
    assertMatch(t, 1, id("A"))
    assertMatch(t, 1, id("A"))
    assertMatch(t, 2, id("B"))
    assertMatch(t, 3, id("C"))
    assertMatch(t, 2, id("B"))
}

func getRootPath(t *testing.T) string {
    pwd, err := os.Getwd()
    if err != nil {
        t.Log("Unable to obtain working directory")
        t.Fail()
    }
    return path.Join(pwd, "../test_files")
}

// Assert that module name resolves to path (relative to test_files).
func assertResolvesToPath(t *testing.T, name, path string) {
    r := Resolver{Path: getRootPath(t)}
    ref := r.loadModule(name)
    t.Log(name, "resolved to", ref)

    if ref == nil {
        t.Log("Unable to load module", name)
        t.Fail()
    }

    relPath, _ := filepath.Rel(r.Path, ref.fullPath())
    if relPath != path {
        t.Logf("Expected %s to resolve to %s but got %s", name, path, relPath)
        t.Fail()
    }
}

func assertParseMatch(t *testing.T, name string, deps []string) {
    r := Resolver{Path: getRootPath(t)}
    ref := r.loadModule(name)
    actualDeps := ref.parse()
    for _, expected := range deps {
        assertContains(t, actualDeps, expected)
    }
}

func assertContains(t *testing.T, haystack []string, needle string) {
    if !contains(haystack, needle) {
        t.Logf("Expected list to contain %s", needle)
        t.Fail()
    }
}

func assertMatch(t *testing.T, expected, actual interface{}) {
    if expected != actual {
        t.Logf("Expected\n %s\n but got\n %s", expected, actual)
        t.Fail()
    }
}

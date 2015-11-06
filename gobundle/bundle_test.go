package gobundle

import (
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
    assertDepsMatch(t, "./foo", []string{"query", "bar.js"})
    assertDepsMatch(t, "./bar", []string{"util", "baz/baz.js"})
    assertDepsMatch(t, "./baz/baz", []string{"util"})
    assertDepsMatch(t, "query", []string{"../util"})
    assertDepsMatch(t, "util", []string{})
}

func getRootPath(t *testing.T) string {
    pwd, err := os.Getwd()
    if err != nil {
        t.Log("Unable to obtain working directory")
        t.Fail()
    }
    return path.Join(pwd, "../test_files")
}

// Assert that module name resolves to path.
func assertResolvesToPath(t *testing.T, name, path string) {
    r := Resolver{}
    rootPath := getRootPath(t)

    ref := r.loadModule(rootPath, name)
    t.Log(name, "resolved to", ref)
    if ref == nil {
        t.Log("Unable to load module", name)
        t.Fail()
    }

    relPath, _ := filepath.Rel(rootPath, ref.fullPath())
    if relPath != path {
        t.Logf("Expected %s to resolve to %s but got %s", name, path, relPath)
        t.Fail()
    }
}

func assertDepsMatch(t *testing.T, name string, deps []string) {
    r := Resolver{}
    rootPath := getRootPath(t)

    ref := r.loadModule(rootPath, name)
    actualDeps := ref.parseDeps()

    for _, dep := range deps {
        if !contains(actualDeps, dep) {
            t.Logf("Expected %s to have dependency %s", name, dep)
            t.Fail()
        }
    }
}

// Source: http://stackoverflow.com/questions/10485743/contains-method-for-a-slice
func contains(haystack []string, needle string) bool {
    for _, a := range haystack {
        if a == needle {
            return true
        }
    }
    return false
}

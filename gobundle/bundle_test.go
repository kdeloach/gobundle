package gobundle

import (
    "os"
    "path"
    "testing"
)

func TestResolver1(t *testing.T) {
    r := Resolver{}
    rootPath := getRootPath(t)
    assertExists(t, &r, rootPath, "./foo")
    assertExists(t, &r, rootPath, "./bar")
    assertExists(t, &r, rootPath, "./baz/baz")
    assertExists(t, &r, rootPath, "query")
    assertExists(t, &r, rootPath, "util")
}

func getRootPath(t *testing.T) string {
    pwd, err := os.Getwd()
    if err != nil {
        t.Log("Unable to obtain working directory")
        t.Fail()
    }
    return path.Join(pwd, "../test_files")
}

func assertExists(t *testing.T, r *Resolver, path, name string) {
    ref := r.loadModule(path, name)
    t.Log(name, "resolved to", ref)
    if ref == nil {
        t.Log("Unable to load module", name)
        t.Fail()
    }
}

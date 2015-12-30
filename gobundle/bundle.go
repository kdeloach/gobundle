package gobundle

import (
    "bufio"
    "encoding/json"
    "io"
    "io/ioutil"
    "log"
    "os"
    "path/filepath"
    "regexp"
    "strconv"
    "strings"
)

// Structures

type NpmPackage struct {
    Name, Main, Version string
}

type Resolver struct {
    Path string
}

type ModRef struct {
    Path, Name string
}

type ModRefGraph struct {
    RootPath string
    EntryFile string
    Nodes map[string][][]string
}

var RequireStmt = regexp.MustCompile(`` +
        `(?i)` +        // Set case-insensitive flag
        `require\(` +
        `(?:"|')` +     // Single or double quote non-capture group
        `([a-z0-9\./\\-]+)` +
        `(?:"|')` +     // Single or double quote non-capture group
        `\)`)

//

func Bundle(entryFile string) ModRefGraph {
    rootPath := filepath.Dir(entryFile)
    r := Resolver{Path: rootPath}
    modPath := r.relPath(entryFile)
    return r.graph(rootPath, modPath)
}

func WriteBundle(b *os.File, bundle ModRefGraph) {
    r := Resolver{Path: bundle.RootPath}

    id := makeIdFunc()

    b.WriteString("(function(L, entry) {")
    b.WriteString(    "var cache = {};")
    b.WriteString(    "function run(id) {")
    b.WriteString(        "if (cache[id]) return cache[id];")
    b.WriteString(        "var m = {exports:{}},")
    b.WriteString(            "fn = L[id][0],")
    b.WriteString(            "deps = L[id][1];")
    b.WriteString(        "function require(name) {")
    b.WriteString(            "return run(deps[name]);")
    b.WriteString(        "}")
    b.WriteString(        "cache[id] = m.exports;");
    b.WriteString(        "fn(require, m, m.exports);");
    b.WriteString(        "return m.exports;");
    b.WriteString(    "}")
    b.WriteString(    "run(entry);")
    b.WriteString("}(")

    i := 0
    b.WriteString("{")
    log.Println(bundle.Nodes)
    // TODO: Filter list beforehand
    for path, children := range bundle.Nodes {
        // TODO: Investigate why empty/nil children are being added here.
        // NOTE: Comma appears at the tail end of list if null entry is last
        if len(path) == 0 {
            continue
        }

        modRef := r.loadModule(path)
        if modRef == nil {
            log.Panic("Unable to load module at", path)
        }

        b.WriteString(strconv.Itoa(id(path)))
        b.WriteString(":[")
        b.WriteString("function(require,module,exports){")
        modRef.writeContents(b)
        b.WriteString("},")

        b.WriteString("{")
        for j, pathTuple := range children {
            childPath := pathTuple[0]
            childRelPath := pathTuple[1]
            b.WriteString("'")
            b.WriteString(childPath)
            b.WriteString("':")
            b.WriteString(strconv.Itoa(id(childRelPath)))
            if j < len(children) - 1 {
                b.WriteString(",")
            }
        }
        b.WriteString("}")

        b.WriteString("]")
        if i < len(bundle.Nodes) - 1 {
            b.WriteString(",")
        }
        i++
    }
    b.WriteString("},")
    b.WriteString(strconv.Itoa(id(bundle.EntryFile)))
    b.WriteString("));\n")
}

//

func (self Resolver) loadModule(name string) *ModRef {
    return self.loadModuleRelativeTo(self.Path, name)
}

func (self Resolver) loadModuleRelativeTo(path, name string) *ModRef {
    log.Println("Trying to load", name, "from", path)
    // NOTE: This important to prevent the case where you have a file name
    // that matches an NPM module. (Ex. you have "shim/jquery.js" which
    // requires "jquery")
    if isRelative(name) {
        if result := self.loadFile(path, name); result != nil {
            return result
        }
        if result := self.loadFolder(path, name); result != nil {
            return result
        }
    }
    if result := self.loadNodeModule(path, name); result != nil {
        return result
    }
    return nil
}

func loadPackage(path string) (*NpmPackage, error) {
    pkg := NpmPackage{}
    data, err := ioutil.ReadFile(path)
    if err != nil {
        return &pkg, err
    }
    err = json.Unmarshal(data, &pkg)
    if err != nil {
        return &pkg, err
    }
    return &pkg, nil
}

func (self Resolver) loadFile(path , name string) *ModRef {
    dir := filepath.Dir(filepath.Join(path, name))
    base := filepath.Base(name)

    file := filepath.Join(dir, base)
    if exists(file) {
        return &ModRef{Path: dir, Name: base}
    }

    file = filepath.Join(dir, base + ".js")
    if exists(file) {
        return &ModRef{Path: dir, Name: base + ".js"}
    }

    file = filepath.Join(dir, base + ".json")
    if exists(file) {
        return &ModRef{Path: dir, Name: base + ".json"}
    }

    return nil
}

func (self Resolver) loadFolder(path, name string) *ModRef {
    dirPath := filepath.Join(path, name)
    pkgFile := filepath.Join(dirPath, "package.json")
    if exists(pkgFile) {
        pkg, err := loadPackage(pkgFile)
        if err != nil {
            log.Panic("Invalid package.json format ", pkgFile)
        }
        if len(pkg.Main) > 0 {
            return self.loadFile(dirPath, pkg.Main)
        }
    }

    file := filepath.Join(dirPath, "index.js")
    if exists(file) {
        return &ModRef{Path: dirPath, Name: "index.js"}
    }

    file = filepath.Join(dirPath, "index.json")
    if exists(file) {
        return &ModRef{Path: dirPath, Name: "index.json"}
    }

    return nil
}

func (self Resolver) loadNodeModule(path, name string) *ModRef {
    absPath, _ := filepath.Abs(path)
    dirPaths := nodeModulePaths(absPath)
    for _, dirPath := range dirPaths {
        log.Println("Trying to load NPM module", name, "from", path)
        if result := self.loadFile(dirPath, name); result != nil {
            return result
        }
        if result := self.loadFolder(dirPath, name); result != nil {
            return result
        }
    }
    return nil
}

func (self Resolver) graph(rootPath, modPath string) ModRefGraph {
    result := ModRefGraph{
        RootPath: rootPath,
        EntryFile: modPath,
        Nodes: make(map[string][][]string),
    }
    self.graph2(result, self.Path, modPath)
    return result
}

func (self Resolver) graph2(result ModRefGraph, relPath, modPath string) *ModRef {
    modRef := self.loadModuleRelativeTo(relPath, modPath)
    if modRef == nil {
        log.Panic("Could not load module: ", modPath, " from ", relPath)
    }

    k := self.relPath(modRef.fullPath())

    if _, exists := result.Nodes[k]; !exists {
        children := modRef.parse()
        // Placeholder to prevent recursive loops.
        // Note: Do we need this if graph never allows for cyclical deps?
        result.Nodes[k] = nil
        childPaths := make([][]string, 0)
        for _, childPath := range children {
            childRef := self.graph2(result, modRef.Path, childPath)
            childRelPath := self.relPath(childRef.fullPath())
            childPaths = append(childPaths, []string{childPath, childRelPath})
        }
        result.Nodes[k] = childPaths
    }
    return modRef
}

func (self Resolver) relPath(path string) string {
    result, _ := filepath.Rel(self.Path, path)
    return result
}

func (self ModRef) fullPath() string {
    return filepath.Join(self.Path, self.Name)
}

// Return map of {moduleName: path, ...} for all dependencies referenced
// by this module.
func (self ModRef) parse() []string {
    fp, _ := os.Open(self.fullPath())
    defer fp.Close()

    scanner := bufio.NewScanner(bufio.NewReader(fp))
    scanner.Split(bufio.ScanLines)

    result := make([]string, 0)
    for scanner.Scan() {
        matches := RequireStmt.FindAllStringSubmatch(scanner.Text(), -1)
        for _, match := range matches {
            // Skip first match (entire unmatched line).
            for _, moduleName := range match[1:] {
                result = append(result, moduleName)
            }
        }
    }
    return result
}

func (self ModRef) writeContents(writer *os.File) {
    fp, _ := os.Open(self.fullPath())
    defer fp.Close()
    io.Copy(writer, fp)
}

// Helpers

func isRelative(name string) bool {
    return strings.HasPrefix(name, "./") || strings.HasPrefix(name, "../")
}

// Return true if a file exists at `path`.
func exists(path string) bool {
    f, err := os.Stat(path)
    if err != nil {
        return false
    }
    return !f.IsDir()
}

func nodeModulePaths(path string) []string {
    result := []string{}
    for ; len(path) > 1; path = filepath.Dir(path) {
        result = append(result, filepath.Join(path, "node_modules"))
    }
    return result
}

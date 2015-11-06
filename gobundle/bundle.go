package gobundle

import (
    "bufio"
    "encoding/json"
    "io/ioutil"
    "log"
    "os"
    "path/filepath"
    "regexp"
    "strings"
)

// Structures

type NpmPackage struct {
    Name, Main, Version string
}

type Resolver struct {
}

type ModRef struct {
    Path, Name string
}

var RequireStmt = regexp.MustCompile(`` +
        `(?i)` +        // Set case-insensitive flag
        `require\(` +
        `(?:"|')` +     // Single or double quote non-capture group
        `([a-z0-9\./\\-]+)` +
        `(?:"|')` +     // Single or double quote non-capture group
        `\)`)
//

func Bundle(entryFiles []string) int {
    return 1;
}

func WriteBundle(writer *os.File, bundle int) {
    writer.WriteString("Hello World")
}

func (self Resolver) loadModule(path , name string) *ModRef {
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
    file := filepath.Join(path, name)
    log.Println("Trying", file)
    if exists(file) {
        return &ModRef{Path: path, Name: name}
    }

    file = filepath.Join(path, name + ".js")
    log.Println("Trying", file)
    if exists(file) {
        return &ModRef{Path: path, Name: name + ".js"}
    }

    file = filepath.Join(path, name + ".json")
    log.Println("Trying", file)
    if exists(file) {
        return &ModRef{Path: path, Name: name + ".json"}
    }

    return nil
}

func (self Resolver) loadFolder(path, name string) *ModRef {
    dirPath := filepath.Join(path, name)
    pkgFile := filepath.Join(dirPath, "package.json")
    if exists(pkgFile) {
        pkg, err := loadPackage(pkgFile)
        if err != nil {
            log.Panic("Invalid package.json format", pkgFile)
        }
        if len(pkg.Main) > 0 {
            return self.loadFile(dirPath, pkg.Main)
        }
    }
    if exists(filepath.Join(dirPath, "index.js")) {
        return &ModRef{Path: dirPath, Name: "index.js"}
    }
    if exists(filepath.Join(dirPath, "index.json")) {
        return &ModRef{Path: dirPath, Name: "index.json"}
    }
    return nil
}

func (self Resolver) loadNodeModule(path, name string) *ModRef {
    dirPaths := nodeModulePaths(path)
    for _, dirPath := range dirPaths {
        if result := self.loadFile(dirPath, name); result != nil {
            return result
        }
        if result := self.loadFolder(dirPath, name); result != nil {
            return result
        }
    }
    return nil
}

func (self ModRef) fullPath() string {
    return filepath.Join(self.Path, self.Name)
}

// Return (non-recursive) list of referenced modules.
func (self ModRef) parseDeps() []string {
    fp, _ := os.Open(self.fullPath())
    defer fp.Close()

    scanner := bufio.NewScanner(bufio.NewReader(fp))
    scanner.Split(bufio.ScanLines)

    result := []string{}
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

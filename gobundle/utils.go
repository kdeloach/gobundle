package gobundle

type idFunc func(path string) int

// Source: http://stackoverflow.com/questions/10485743/contains-method-for-a-slice
func contains(haystack []string, needle string) bool {
    for _, a := range haystack {
        if a == needle {
            return true
        }
    }
    return false
}

func makeIdFunc() idFunc {
    i := 0
    lookup := make(map[string]int)
    return func(path string) int {
        if n, exists := lookup[path]; exists {
            return n
        }
        i++
        lookup[path] = i
        return i
    }
}

package parser

import (
	"strings"
	"sync"

	"github.com/titpetric/lessgo/ast"
)

// pathCache caches parsed paths to avoid re-parsing frequently-used expressions.
var pathCache = &struct {
	sync.RWMutex
	m map[string][]string
}{
	m: make(map[string][]string),
}

const pathCacheLimit = 256

// getCachedPath returns a cached path or computes and caches it.
func getCachedPath(expr string) []string {
	pathCache.RLock()
	if parts, ok := pathCache.m[expr]; ok {
		pathCache.RUnlock()
		return parts
	}
	pathCache.RUnlock()

	// Not in cache, compute it
	parts := splitPathImpl(expr)

	// Cache if under limit
	if len(pathCache.m) < pathCacheLimit {
		pathCache.Lock()
		pathCache.m[expr] = parts
		pathCache.Unlock()
	}

	return parts
}

// Stack provides stack-based variable lookup and convenient typed accessors.
// Used for maintaining scoped variables during rendering.
type Stack struct {
	stack []map[string]ast.Value // bottom..top, top is last element
}

// NewStack constructs a Stack with an optional initial root map.
func NewStack(root map[string]ast.Value) *Stack {
	s := &Stack{}
	if root == nil {
		root = map[string]ast.Value{}
	}
	s.stack = []map[string]ast.Value{root}
	return s
}

// mapPool caches map[string]ast.Value allocations to reduce GC pressure.
var mapPool = sync.Pool{
	New: func() any {
		return make(map[string]ast.Value, 0)
	},
}

// Push a new map as a top-most Stack.
// If m is nil, an empty map is obtained from the pool.
func (s *Stack) Push(m map[string]ast.Value) {
	if m == nil {
		m = mapPool.Get().(map[string]ast.Value)
	}
	s.stack = append(s.stack, m)
}

// Pop the top-most Stack. If only root remains it still pops to empty slice safely.
// Returns pooled maps to reduce GC pressure.
func (s *Stack) Pop() {
	if len(s.stack) == 0 {
		return
	}
	// Return the top map to the pool before removing it
	topIdx := len(s.stack) - 1
	topMap := s.stack[topIdx]
	// Clear the map and return it to pool if it's not the root
	if topIdx > 0 && len(topMap) > 0 {
		for k := range topMap {
			delete(topMap, k)
		}
		mapPool.Put(topMap)
	}
	s.stack = s.stack[:topIdx]
	if len(s.stack) == 0 {
		s.stack = append(s.stack, map[string]ast.Value{})
	}
}

// Set sets a key in the top-most Stack.
func (s *Stack) Set(key string, val ast.Value) {
	if len(s.stack) == 0 {
		s.stack = append(s.stack, map[string]ast.Value{})
	}
	s.stack[len(s.stack)-1][key] = val
}

// Lookup searches stack from top to bottom for a plain identifier (no dots).
// Returns (value, true) if found.
func (s *Stack) Lookup(name string) (ast.Value, bool) {
	for i := len(s.stack) - 1; i >= 0; i-- {
		if v, ok := s.stack[i][name]; ok {
			return v, true
		}
	}
	return nil, false
}

// Resolve resolves dotted/bracketed expression paths like:
//
//	"user.name", "items[0].title", "mapKey.sub"
//
// It returns (value, true) if resolution succeeded.
func (s *Stack) Resolve(expr string) (ast.Value, bool) {
	// Fast path: if no dots or brackets, do direct lookup
	if !strings.ContainsAny(expr, ".[") {
		return s.Lookup(expr)
	}

	// Parse once (with caching)
	parts := getCachedPath(expr)
	if len(parts) == 0 {
		return nil, false
	}

	// first part must come from Stack
	cur, ok := s.Lookup(parts[0])
	if !ok {
		return nil, false
	}
	// walk the rest
	for _, p := range parts[1:] {
		if cur == nil {
			return nil, false
		}
		cur = s.resolveStep(cur, p)
		if cur == nil {
			return nil, false
		}
	}
	return cur, true
}

// resolveStep resolves a single step in a path, returning nil if resolution fails.
func (s *Stack) resolveStep(cur ast.Value, p string) ast.Value {
	// For now, just return nil - path resolution not needed for LESS variables
	// This could be enhanced later for nested variable access
	return nil
}

// EnvMap converts the Stack to a map[string]ast.Value.
// Includes all accessible values from stack.
func (s *Stack) EnvMap() map[string]ast.Value {
	result := make(map[string]ast.Value)
	// Iterate through stack from bottom to top, with top overriding bottom
	for i := 0; i < len(s.stack); i++ {
		for k, v := range s.stack[i] {
			result[k] = v
		}
	}
	return result
}

// ForEach iterates over a collection at the given expr and calls fn(index,value).
// Currently not implemented for LESS - can be enhanced if needed for each() loops.
func (s *Stack) ForEach(expr string, fn func(index int, value ast.Value) error) error {
	// Not needed for basic LESS compilation
	return nil
}

// Helpers

// splitPath turns expressions into path parts. Supports dot notation and bracket numeric/string indexes.
// Now delegates to splitPathImpl which is cached by getCachedPath.
// examples:
//
//	"items[0].name" -> ["items","0","name"]
//	"user.name" -> ["user","name"]
//	"a['b'].c" -> ["a","b","c"]
func (s *Stack) splitPath(expr string) []string {
	return getCachedPath(expr)
}

// splitPathImpl is the actual implementation of path splitting.
// Called by getCachedPath which caches the results.
func splitPathImpl(expr string) []string {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return nil
	}

	// Fast path: if no brackets, just split by dots
	if !strings.Contains(expr, "[") {
		parts := strings.Split(expr, ".")
		// Sanitize in-place to avoid extra allocation
		out := parts[:0]
		for _, p := range parts {
			if p = strings.TrimSpace(p); p != "" {
				out = append(out, p)
			}
		}
		return out
	}

	// Full parsing with bracket support
	var b strings.Builder
	b.Grow(len(expr) + 8)
	i := 0
	for i < len(expr) {
		ch := expr[i]
		if ch == '[' {
			j := i + 1
			for j < len(expr) && expr[j] != ']' {
				j++
			}
			if j >= len(expr) {
				b.WriteByte(ch)
				i++
				continue
			}
			inside := strings.TrimSpace(expr[i+1 : j])
			if len(inside) >= 2 && ((inside[0] == '\'' && inside[len(inside)-1] == '\'') || (inside[0] == '"' && inside[len(inside)-1] == '"')) {
				inside = inside[1 : len(inside)-1]
			}
			if inside != "" {
				b.WriteByte('.')
				b.WriteString(inside)
			}
			i = j + 1
		} else {
			b.WriteByte(ch)
			i++
		}
	}

	builtStr := b.String()
	parts := strings.Split(builtStr, ".")
	// Sanitize in-place to avoid extra allocation
	out := parts[:0]
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	return out
}

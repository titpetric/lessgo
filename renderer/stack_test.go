package renderer

import (
	"testing"
)

func TestStackBasic(t *testing.T) {
	s := NewStack()

	// Set and get global variable
	s.Set("color", "#fff")
	val, ok := s.Get("color")
	if !ok || val != "#fff" {
		t.Errorf("Get(color) = %s, %v, want #fff, true", val, ok)
	}
}

func TestStackScoping(t *testing.T) {
	s := NewStack()

	// Global scope
	s.SetGlobal("global-var", "global-value")

	// Push new scope
	s.Push()
	s.Set("local-var", "local-value")

	// Can access both global and local
	if v, ok := s.Get("global-var"); !ok || v != "global-value" {
		t.Errorf("Get(global-var) in local scope = %s, %v", v, ok)
	}
	if v, ok := s.Get("local-var"); !ok || v != "local-value" {
		t.Errorf("Get(local-var) = %s, %v", v, ok)
	}

	// Pop scope
	s.Pop()

	// Can access global but not local
	if v, ok2 := s.Get("global-var"); !ok2 || v != "global-value" {
		t.Errorf("Get(global-var) after pop = %s, %v", v, ok2)
	}
	if _, ok := s.Get("local-var"); ok {
		t.Errorf("Get(local-var) after pop should fail")
	}
}

func TestStackShadowing(t *testing.T) {
	s := NewStack()

	// Global scope
	s.Set("var", "global")

	// Local scope shadows global
	s.Push()
	s.Set("var", "local")

	val, _ := s.Get("var")
	if val != "local" {
		t.Errorf("shadowed Get(var) = %s, want local", val)
	}

	// After pop, global value visible again
	s.Pop()
	val2, _ := s.Get("var")
	if val2 != "global" {
		t.Errorf("Get(var) after pop = %s, want global", val2)
	}
}

func TestStackDepth(t *testing.T) {
	s := NewStack()
	if s.Depth() != 1 {
		t.Errorf("Initial depth = %d, want 1", s.Depth())
	}

	s.Push()
	if s.Depth() != 2 {
		t.Errorf("Depth after push = %d, want 2", s.Depth())
	}

	s.Push()
	if s.Depth() != 3 {
		t.Errorf("Depth after second push = %d, want 3", s.Depth())
	}

	s.Pop()
	if s.Depth() != 2 {
		t.Errorf("Depth after pop = %d, want 2", s.Depth())
	}
}

func TestStackAll(t *testing.T) {
	s := NewStack()
	s.Set("a", "1")
	s.Set("b", "2")

	s.Push()
	s.Set("c", "3")
	s.Set("b", "2-local") // shadow b

	all := s.All()
	if all["a"] != "1" || all["c"] != "3" {
		t.Errorf("All() missing global vars")
	}
	if all["b"] != "2-local" {
		t.Errorf("All() shadowed var = %s, want 2-local", all["b"])
	}
}

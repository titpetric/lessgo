package renderer

// Stack represents a variable scope stack for managing variable lifetimes
type Stack struct {
	frames []map[string]string // Stack of scope frames
}

// NewStack creates a new variable stack with a global scope
func NewStack() *Stack {
	return &Stack{
		frames: []map[string]string{
			make(map[string]string), // Global scope
		},
	}
}

// Push creates a new scope level on the stack
func (s *Stack) Push() {
	s.frames = append(s.frames, make(map[string]string))
}

// Pop removes the top scope level from the stack
func (s *Stack) Pop() {
	if len(s.frames) > 1 { // Keep at least global scope
		s.frames = s.frames[:len(s.frames)-1]
	}
}

// Set sets a variable in the current (topmost) scope
func (s *Stack) Set(name, value string) {
	if len(s.frames) == 0 {
		return
	}
	s.frames[len(s.frames)-1][name] = value
}

// Get retrieves a variable by searching from the current scope up to global scope
func (s *Stack) Get(name string) (string, bool) {
	// Search from current scope (top of stack) downward to global scope
	for i := len(s.frames) - 1; i >= 0; i-- {
		if val, ok := s.frames[i][name]; ok {
			return val, true
		}
	}
	return "", false
}

// SetGlobal sets a variable in the global scope
func (s *Stack) SetGlobal(name, value string) {
	if len(s.frames) > 0 {
		s.frames[0][name] = value
	}
}

// GetGlobal retrieves a variable only from the global scope
func (s *Stack) GetGlobal(name string) (string, bool) {
	if len(s.frames) > 0 {
		val, ok := s.frames[0][name]
		return val, ok
	}
	return "", false
}

// Depth returns the current stack depth
func (s *Stack) Depth() int {
	return len(s.frames)
}

// All returns all variables visible in the current scope (including parent scopes)
func (s *Stack) All() map[string]string {
	result := make(map[string]string)
	// Merge from global to current scope (so current scope takes precedence)
	for i := 0; i < len(s.frames); i++ {
		for k, v := range s.frames[i] {
			result[k] = v
		}
	}
	return result
}

// AllAny
func (s *Stack) Any() map[string]any {
	all := s.All()
	result := make(map[string]any, len(all))
	for k, v := range all {
		result[k] = v
	}
	return result
}

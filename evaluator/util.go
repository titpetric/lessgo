package evaluator

// isValueChar checks if a character can be part of a value
func isValueChar(r rune) bool {
	return (r >= '0' && r <= '9') || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '%' || r == '#'
}

// isDigit checks if a rune is a digit
func isDigit(r byte) bool {
	return r >= '0' && r <= '9'
}

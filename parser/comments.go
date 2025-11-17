package parser

import (
	"strings"
)

// CommentInfo stores information about a comment found in the source
type CommentInfo struct {
	Text      string // Content without delimiters
	IsBlock   bool   // true for /* */, false for //
	StartLine int
	EndLine   int
	StartCol  int
	EndCol    int
}

// ExtractComments extracts all comments from the source code, preserving their positions
// Returns a map from line number to comment on that line
// Only returns BLOCK comments (/* */) that are on their own line (not mixed with code)
// NOTE: Single-line comments (//) are dropped to match lessc behavior
func ExtractComments(source string) map[int][]CommentInfo {
	comments := make(map[int][]CommentInfo)
	lines := strings.Split(source, "\n")

	for lineNum, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip empty lines
		if trimmed == "" {
			continue
		}

		// Skip lines that are indented (inside blocks) - only process top-level comments
		if len(line) > 0 && (line[0] == ' ' || line[0] == '\t') {
			continue
		}

		// NOTE: Skip line comments (//) entirely to match lessc behavior
		// lessc drops all // comments, so we do the same

		// Process block comments (/* ... */) only if they're on their own line
		if strings.HasPrefix(trimmed, "/*") {
			pos := 0
			for {
				idx := strings.Index(line[pos:], "/*")
				if idx == -1 {
					break
				}
				idx += pos
				if !isInString(line, idx) {
					// Found start of block comment
					endIdx := strings.Index(line[idx+2:], "*/")
					if endIdx != -1 {
						// Single-line block comment
						endIdx += idx + 2
						text := strings.TrimSpace(line[idx+2 : endIdx])

						// Check if there's code before or after the comment
						beforeComment := strings.TrimSpace(line[:idx])
						afterComment := strings.TrimSpace(line[endIdx+2:])
						isOwnLine := (beforeComment == "" || beforeComment == "}") && afterComment == ""

						if isOwnLine {
							comments[lineNum] = append(comments[lineNum], CommentInfo{
								Text:      text,
								IsBlock:   true,
								StartLine: lineNum,
								EndLine:   lineNum,
								StartCol:  idx,
							})
						}
						pos = endIdx + 2
					} else {
						// Multi-line block comment - would need more complex handling
						pos = idx + 2
					}
				} else {
					pos = idx + 2
				}
			}
		}
	}

	return comments
}

// isInString checks if a position is inside a string literal
// Simple check: count quotes before the position
func isInString(line string, pos int) bool {
	inString := false
	inSingleQuote := false
	escaped := false

	for i := 0; i < pos && i < len(line); i++ {
		ch := line[i]

		if escaped {
			escaped = false
			continue
		}

		if ch == '\\' {
			escaped = true
			continue
		}

		if ch == '"' && !inSingleQuote {
			inString = !inString
		} else if ch == '\'' && !inString {
			inSingleQuote = !inSingleQuote
		}
	}

	return inString || inSingleQuote
}

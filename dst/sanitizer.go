package dst

import (
	"bytes"
	"io"
)

// SanitizeReader wraps an io.Reader and sanitizes minified CSS/LESS input
// by injecting newlines after '{', ';', '}' and before '}' characters.
// This allows the line-based parser to handle minified input correctly.
func SanitizeReader(r io.Reader) io.Reader {
	data, err := io.ReadAll(r)
	if err != nil {
		return bytes.NewReader(nil)
	}

	sanitized := SanitizeBytes(data)
	return bytes.NewReader(sanitized)
}

// SanitizeBytes sanitizes minified CSS/LESS input by injecting newlines
// after '{', ';', '}' and before '}' characters.
// Respects quoted strings, comments, and @{...} interpolation blocks.
func SanitizeBytes(data []byte) []byte {
	if len(data) == 0 {
		return data
	}

	// Estimate output size - may need up to 2x for heavily minified input
	result := make([]byte, 0, len(data)*2)

	inSingleQuote := false
	inDoubleQuote := false
	inSingleLineComment := false
	inMultiLineComment := false
	inInterpolation := false

	// Track the last meaningful character (outside comments/quotes)
	// Used to determine if we need to add ';' before '}'
	lastMeaningfulChar := byte(0)

	for i := 0; i < len(data); i++ {
		ch := data[i]
		prevCh := byte(0)
		if i > 0 {
			prevCh = data[i-1]
		}
		nextCh := byte(0)
		if i+1 < len(data) {
			nextCh = data[i+1]
		}

		// Handle single-line comments
		if !inSingleQuote && !inDoubleQuote && !inMultiLineComment && !inInterpolation && ch == '/' && nextCh == '/' {
			inSingleLineComment = true
			result = append(result, ch)
			continue
		}

		// End single-line comment at newline
		if inSingleLineComment && ch == '\n' {
			inSingleLineComment = false
			result = append(result, ch)
			continue
		}

		// Handle multi-line comments
		if !inSingleQuote && !inDoubleQuote && !inSingleLineComment && !inInterpolation && ch == '/' && nextCh == '*' {
			inMultiLineComment = true
			result = append(result, ch)
			continue
		}

		// End multi-line comment
		if inMultiLineComment && ch == '*' && nextCh == '/' {
			result = append(result, ch)
			result = append(result, nextCh)
			i++ // Skip the '/'
			inMultiLineComment = false
			continue
		}

		// In comment, just pass through
		if inSingleLineComment || inMultiLineComment {
			result = append(result, ch)
			continue
		}

		// Handle quotes (respecting escapes)
		if ch == '\'' && prevCh != '\\' && !inDoubleQuote && !inInterpolation {
			inSingleQuote = !inSingleQuote
			result = append(result, ch)
			if !inSingleQuote {
				lastMeaningfulChar = ch
			}
			continue
		}

		if ch == '"' && prevCh != '\\' && !inSingleQuote && !inInterpolation {
			inDoubleQuote = !inDoubleQuote
			result = append(result, ch)
			if !inDoubleQuote {
				lastMeaningfulChar = ch
			}
			continue
		}

		// In quoted string, just pass through
		if inSingleQuote || inDoubleQuote {
			result = append(result, ch)
			continue
		}

		// Handle @{...} interpolation blocks
		if ch == '@' && nextCh == '{' {
			inInterpolation = true
			result = append(result, ch)
			continue
		}

		// End interpolation block
		if inInterpolation && ch == '}' {
			inInterpolation = false
			result = append(result, ch)
			lastMeaningfulChar = ch
			continue
		}

		// In interpolation, just pass through
		if inInterpolation {
			result = append(result, ch)
			continue
		}

		// Handle structural characters outside quotes/comments/interpolation
		switch ch {
		case '{':
			result = append(result, ch)
			lastMeaningfulChar = ch
			// Add newline after '{' if not already followed by newline
			if nextCh != '\n' && nextCh != '\r' {
				result = append(result, '\n')
			}

		case ';':
			result = append(result, ch)
			lastMeaningfulChar = ch
			// Add newline after ';' if not already followed by newline
			if nextCh != '\n' && nextCh != '\r' {
				result = append(result, '\n')
			}

		case '}':
			// Trim trailing whitespace before '}'
			result = trimTrailingWhitespace(result)
			// Add ';' before '}' if the last statement doesn't have one
			// (the trailing ';' is optional in CSS/LESS)
			// Only do this if there was actual content (not just comments/whitespace)
			if lastMeaningfulChar != ';' && lastMeaningfulChar != '{' && lastMeaningfulChar != '}' && lastMeaningfulChar != 0 {
				result = append(result, ';')
			}
			// Add newline before '}' if not already preceded by newline
			if len(result) > 0 && result[len(result)-1] != '\n' && result[len(result)-1] != '\r' {
				result = append(result, '\n')
			}
			result = append(result, ch)
			lastMeaningfulChar = ch
			// Add newline after '}' if not already followed by newline or ';'
			if nextCh != '\n' && nextCh != '\r' && nextCh != ';' && nextCh != 0 {
				result = append(result, '\n')
			}

		default:
			// Track non-whitespace characters
			if ch != ' ' && ch != '\t' && ch != '\n' && ch != '\r' {
				lastMeaningfulChar = ch
			}
			result = append(result, ch)
		}
	}

	return result
}

// trimTrailingWhitespace removes trailing spaces and tabs from the slice.
// Does not remove newlines since those are structural.
func trimTrailingWhitespace(data []byte) []byte {
	for len(data) > 0 && (data[len(data)-1] == ' ' || data[len(data)-1] == '\t') {
		data = data[:len(data)-1]
	}
	return data
}

package dst

import (
	"io"
	"io/fs"
)

// ParserConfig determines which parser implementation to use
type ParserConfig struct {
	UseNoAlloc bool
}

// DefaultParserConfig uses the optimized ParserNoAlloc for better performance
var DefaultParserConfig = ParserConfig{
	UseNoAlloc: false,
}

// UseParser sets the global parser configuration
func UseParser(config ParserConfig) {
	DefaultParserConfig = config
}

// NewParserForConfig creates a parser instance based on current configuration
func NewParserForConfig(r io.Reader) interface {
	Parse() (*File, error)
} {
	if DefaultParserConfig.UseNoAlloc {
		return NewParserNoAlloc(r)
	}
	return NewParser(r)
}

// NewParserForConfigWithFS creates a parser instance with custom filesystem
func NewParserForConfigWithFS(r io.Reader, filesystem fs.FS) interface {
	Parse() (*File, error)
} {
	if DefaultParserConfig.UseNoAlloc {
		return NewParserNoAllocWithFS(r, filesystem)
	}
	return NewParserWithFS(r, filesystem)
}

// AssumeNoAllocParser returns true if configured to use ParserNoAlloc
func AssumeNoAllocParser() bool {
	return DefaultParserConfig.UseNoAlloc
}

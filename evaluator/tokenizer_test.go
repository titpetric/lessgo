package evaluator

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTokenizer(t *testing.T) {
	tok, err := Tokenize("1px 0 2px rgb(1, 2, 3)")
	require.NotEmpty(t, tok)
	require.NoError(t, err)

	tok, err = Tokenize("Arial, sans-serif")
	require.NotEmpty(t, tok)
	require.NoError(t, err)

	tok, err = Tokenize("1px solid rgb(0, 0, 0)")
	require.NotEmpty(t, tok)
	require.NoError(t, err)
}

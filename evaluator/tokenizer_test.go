package evaluator

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/require"
)

func TestTokenizer(t *testing.T) {
	tok, err := Tokenize("1px 0 2px rgb(1, 2, 3)")
	require.NoError(t, err)

	//	spew.Dump(tok)

	tok, err = Tokenize("Arial, sans-serif")
	require.NoError(t, err)

	//	spew.Dump(tok)

	tok, err = Tokenize("1px solid rgb(0, 0, 0)")
	require.NoError(t, err)

	spew.Dump(tok)
}

package test

import (
	"github.com/stretchr/testify/require"
	"testing"
	"ubc/lib"
)

func testGenPrivKey(t *testing.T, label string) {

	w := lib.LoadWallet()
	pk := w.GenPrivKey("issuer")
	b1 := pk.PubKey()
	b2 := pk.PubKey()

	require.EqualValues(t, b1, b2)
}

func TestGenPrivKey(t *testing.T) {

	testGenPrivKey(t, "issuer")
}

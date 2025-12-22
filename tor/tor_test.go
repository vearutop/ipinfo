package tor_test

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vearutop/ipinfo/tor"
	"github.com/vearutop/netrie"
)

func TestLoadExitNodes(t *testing.T) {
	tr := netrie.NewCIDRIndex()

	require.NoError(t, tor.LoadExitNodes(tr, "testdata/torlist.txt"))

	assert.Equal(t, 1595, tr.Len())
	assert.Equal(t, 1, tr.LenNames())
	assert.Equal(t, 48739, tr.LenNodes())

	tr.Minimize()

	assert.Equal(t, 1595, tr.Len())
	assert.Equal(t, 1, tr.LenNames())
	assert.Equal(t, 28186, tr.LenNodes())

	ms := runtime.MemStats{}
	runtime.ReadMemStats(&ms)

	require.NoError(t, tr.SaveToFile("tor.bin"))

	tr2, err := netrie.OpenFile("tor.bin")
	require.NoError(t, err)

	assert.Equal(t, 1595, tr2.Len())
	assert.Equal(t, 1, tr2.LenNames())

	assert.Equal(t, "", tr2.Lookup("178.15.138.158"))
	assert.Equal(t, "tor-exit-nodes", tr2.Lookup("104.244.74.97"))
}

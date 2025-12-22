package cloud_test

import (
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vearutop/ipinfo/cloud"
	"github.com/vearutop/netrie"
)

func TestLoadCloudLocal(t *testing.T) {
	tr := netrie.NewCIDRIndex()
	tr.Metadata().BuildDate = time.Now()
	tr.Metadata().Description = "github.com/disposable/cloud-ip-ranges"

	dir := "testdata/"

	err := cloud.LoadCloudLocal(tr, dir)
	require.NoError(t, err)

	assert.Equal(t, 38873, tr.Len())
	assert.Equal(t, 38, tr.LenNames())
	assert.Equal(t, 272358, tr.LenNodes())

	ms := runtime.MemStats{}
	runtime.ReadMemStats(&ms)

	tr.Minimize()

	assert.Equal(t, 86030, tr.LenNodes())

	require.NoError(t, tr.SaveToFile("cloud.bin"))

	tr2, err := netrie.OpenFile("cloud.bin")
	require.NoError(t, err)

	defer func() {
		assert.NoError(t, tr2.Close())
	}()

	assert.Equal(t, "apple-icloud", tr2.Lookup("172.224.227.36"))
	assert.Equal(t, "", tr2.Lookup("178.15.138.158"))
	assert.Equal(t, "digitalocean", tr2.Lookup("143.198.196.44"))
	assert.Equal(t, "google-bot", tr2.Lookup("66.249.66.71"))
	assert.Equal(t, "heroku-aws", tr2.Lookup("18.215.140.160"))
	assert.Equal(t, "heroku-aws", tr2.Lookup("18.213.114.129"))
}

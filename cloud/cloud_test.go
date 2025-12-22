package cloud_test

import (
	"fmt"
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

	dir := "../cloud-ip-ranges/"

	err := cloud.LoadCloudLocal(tr, dir)
	if err != nil {
		t.Fatal(err)
	}

	println("nets:", tr.Len())
	println("names:", tr.LenNames())
	println("nodes:", tr.LenNodes())

	ms := runtime.MemStats{}
	runtime.ReadMemStats(&ms)

	tr.Minimize()

	fmt.Println(tr.Metadata())

	println("nodes minimized:", tr.LenNodes())

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

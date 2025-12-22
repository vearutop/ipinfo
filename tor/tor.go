// Package tor makes an index of tor exit nodes.
package tor

import (
	"github.com/vearutop/netrie"
	"github.com/vearutop/netrie/lists"
)

// LoadExitNodes loads TOR exit nodes from https://www.dan.me.uk/torlist/?exit.
func LoadExitNodes(tr netrie.Adder, listFile string) error {
	return lists.LoadFromTextGroupCIDRs(listFile, tr, "tor-exit-nodes")
}

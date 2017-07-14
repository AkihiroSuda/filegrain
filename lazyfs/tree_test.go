package lazyfs

import (
	"testing"
)

func TestNode(t *testing.T) {
	nm := newNodeManager("/")
	nm.insert("/usr/bin/foo", "apple")
	nm.insert("/usr/bin/bar", "banana")
	nm.insert("/usr/bin", "chocolate")
	nm.insert("/bin", "doughnut")
	nm.insert("/usr/lib", "elderberry")
	// TODO: add assertion
	t.Logf("nm.root: %#v", nm.root)
	t.Logf("nm.lookup(\"\"): %#v", nm.lookup(""))
	t.Logf("nm.lookup(\"/\"): %#v", nm.lookup("/"))
	t.Logf("nm.lookup(\"/usr\"): %#v", nm.lookup("/usr"))
	t.Logf("nm.lookup(\"/usr/bin\"): %#v", nm.lookup("/usr/bin"))
	t.Logf("nm.lookup(\"/usr/bin/foo\"): %#v", nm.lookup("/usr/bin/foo"))
	t.Logf("nm.lookup(\"/usr/bin/bar\"): %#v", nm.lookup("/usr/bin/bar"))
	t.Logf("nm.lookup(\"/usr/bin/NX\"): %#v", nm.lookup("/usr/bin/NX"))
	t.Logf("nm.lookup(\"/NX\"): %#v", nm.lookup("/NX"))
}

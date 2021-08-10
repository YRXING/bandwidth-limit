package controller

import (
	"github.com/vishvananda/netns"
	"os"
	"runtime"
	"testing"
)

type tearDownNetlinkTest func()

func skipUnlessRoot(t *testing.T) {
	if os.Getuid() != 0 {
		t.Skip("Test requires root privileges.")
	}
}

func setUpNetlinkTest(t *testing.T) tearDownNetlinkTest {
	skipUnlessRoot(t)

	// new temporary namespace so we don't pollute the host
	// lock thread since the namespace is thread local
	runtime.LockOSThread()
	var err error
	ns, err := netns.New()
	if err != nil {
		t.Fatal("Failed to create newns", ns)
	}

	return func() {
		ns.Close()
		runtime.UnlockOSThread()
	}
}

func TestCreateIfb(t *testing.T) {
	tearDown := setUpNetlinkTest(t)
	defer tearDown()

}

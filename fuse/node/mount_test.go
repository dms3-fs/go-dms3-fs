// +build !nofuse

package node

import (
	"io/ioutil"
	"os"
	"os/exec"
	"testing"
	"time"

	"context"

	core "github.com/dms3-fs/go-dms3-fs/core"
	dms3ns "github.com/dms3-fs/go-dms3-fs/fuse/dms3ns"
	mount "github.com/dms3-fs/go-dms3-fs/fuse/mount"
	namesys "github.com/dms3-fs/go-dms3-fs/namesys"

	offroute "github.com/dms3-fs/go-fs-routing/offline"
	ci "github.com/dms3-p2p/go-testutil/ci"
)

func maybeSkipFuseTests(t *testing.T) {
	if ci.NoFuse() {
		t.Skip("Skipping FUSE tests")
	}
}

func mkdir(t *testing.T, path string) {
	err := os.Mkdir(path, os.ModeDir|os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
}

// Test externally unmounting, then trying to unmount in code
func TestExternalUnmount(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	// TODO: needed?
	maybeSkipFuseTests(t)

	node, err := core.NewNode(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}

	err = node.LoadPrivateKey()
	if err != nil {
		t.Fatal(err)
	}

	node.Routing = offroute.NewOfflineRouter(node.Repo.Datastore(), node.RecordValidator)
	node.Namesys = namesys.NewNameSystem(node.Routing, node.Repo.Datastore(), 0)

	err = dms3ns.InitializeKeyspace(node, node.PrivateKey)
	if err != nil {
		t.Fatal(err)
	}

	// get the test dir paths (/tmp/fusetestXXXX)
	dir, err := ioutil.TempDir("", "fusetest")
	if err != nil {
		t.Fatal(err)
	}

	dms3fsDir := dir + "/dms3fs"
	dms3nsDir := dir + "/dms3ns"
	mkdir(t, dms3fsDir)
	mkdir(t, dms3nsDir)

	err = Mount(node, dms3fsDir, dms3nsDir)
	if err != nil {
		t.Fatal(err)
	}

	// Run shell command to externally unmount the directory
	cmd := "fusermount"
	args := []string{"-u", dms3nsDir}
	if err := exec.Command(cmd, args...).Run(); err != nil {
		t.Fatal(err)
	}

	// TODO(noffle): it takes a moment for the goroutine that's running fs.Serve to be notified and do its cleanup.
	time.Sleep(time.Millisecond * 100)

	// Attempt to unmount DMS3NS; check that it was already unmounted.
	err = node.Mounts.Dms3Ns.Unmount()
	if err != mount.ErrNotMounted {
		t.Fatal("Unmount should have failed")
	}

	// Attempt to unmount DMS3FS; it should unmount successfully.
	err = node.Mounts.Dms3Fs.Unmount()
	if err != nil {
		t.Fatal(err)
	}
}

// +build !windows,!nofuse

package node

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	core "github.com/dms3-fs/go-dms3-fs/core"
	dms3ns "github.com/dms3-fs/go-dms3-fs/fuse/dms3ns"
	mount "github.com/dms3-fs/go-dms3-fs/fuse/mount"
	rofs "github.com/dms3-fs/go-dms3-fs/fuse/readonly"

	logging "github.com/dms3-fs/go-log"
)

var log = logging.Logger("node")

// fuseNoDirectory used to check the returning fuse error
const fuseNoDirectory = "fusermount: failed to access mountpoint"

// fuseExitStatus1 used to check the returning fuse error
const fuseExitStatus1 = "fusermount: exit status 1"

// platformFuseChecks can get overridden by arch-specific files
// to run fuse checks (like checking the OSXFUSE version)
var platformFuseChecks = func(*core.Dms3FsNode) error {
	return nil
}

func Mount(node *core.Dms3FsNode, fsdir, nsdir string) error {
	// check if we already have live mounts.
	// if the user said "Mount", then there must be something wrong.
	// so, close them and try again.
	if node.Mounts.Dms3Fs != nil && node.Mounts.Dms3Fs.IsActive() {
		node.Mounts.Dms3Fs.Unmount()
	}
	if node.Mounts.Dms3Ns != nil && node.Mounts.Dms3Ns.IsActive() {
		node.Mounts.Dms3Ns.Unmount()
	}

	if err := platformFuseChecks(node); err != nil {
		return err
	}

	return doMount(node, fsdir, nsdir)
}

func doMount(node *core.Dms3FsNode, fsdir, nsdir string) error {
	fmtFuseErr := func(err error, mountpoint string) error {
		s := err.Error()
		if strings.Contains(s, fuseNoDirectory) {
			s = strings.Replace(s, `fusermount: "fusermount:`, "", -1)
			s = strings.Replace(s, `\n", exit status 1`, "", -1)
			return errors.New(s)
		}
		if s == fuseExitStatus1 {
			s = fmt.Sprintf("fuse failed to access mountpoint %s", mountpoint)
			return errors.New(s)
		}
		return err
	}

	// this sync stuff is so that both can be mounted simultaneously.
	var fsmount, nsmount mount.Mount
	var err1, err2 error

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		fsmount, err1 = rofs.Mount(node, fsdir)
	}()

	if node.OnlineMode() {
		wg.Add(1)
		go func() {
			defer wg.Done()
			nsmount, err2 = dms3ns.Mount(node, nsdir, fsdir)
		}()
	}

	wg.Wait()

	if err1 != nil {
		log.Errorf("error mounting: %s", err1)
	}

	if err2 != nil {
		log.Errorf("error mounting: %s", err2)
	}

	if err1 != nil || err2 != nil {
		if fsmount != nil {
			fsmount.Unmount()
		}
		if nsmount != nil {
			nsmount.Unmount()
		}

		if err1 != nil {
			return fmtFuseErr(err1, fsdir)
		}
		return fmtFuseErr(err2, nsdir)
	}

	// setup node state, so that it can be cancelled
	node.Mounts.Dms3Fs = fsmount
	node.Mounts.Dms3Ns = nsmount
	return nil
}

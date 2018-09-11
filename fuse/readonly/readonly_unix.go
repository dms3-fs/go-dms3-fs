// +build linux darwin freebsd netbsd openbsd
// +build !nofuse

package readonly

import (
	"context"
	"fmt"
	"io"
	"os"
	"syscall"

	core "github.com/dms3-fs/go-dms3-fs/core"
	mdag "github.com/dms3-fs/go-merkledag"
	path "github.com/dms3-fs/go-path"
	uio "github.com/dms3-fs/go-unixfs/io"
	ftpb "github.com/dms3-fs/go-unixfs/pb"

	fuse "bazil.org/fuse"
	fs "bazil.org/fuse/fs"
	proto "github.com/gogo/protobuf/proto"
	dms3ld "github.com/dms3-fs/go-ld-format"
	logging "github.com/dms3-fs/go-log"
	lgbl "github.com/dms3-p2p/go-p2p-loggables"
)

var log = logging.Logger("fuse/dms3fs")

// FileSystem is the readonly DMS3FS Fuse Filesystem.
type FileSystem struct {
	Dms3Fs *core.Dms3FsNode
}

// NewFileSystem constructs new fs using given core.Dms3FsNode instance.
func NewFileSystem(dms3fs *core.Dms3FsNode) *FileSystem {
	return &FileSystem{Dms3Fs: dms3fs}
}

// Root constructs the Root of the filesystem, a Root object.
func (f FileSystem) Root() (fs.Node, error) {
	return &Root{Dms3Fs: f.Dms3Fs}, nil
}

// Root is the root object of the filesystem tree.
type Root struct {
	Dms3Fs *core.Dms3FsNode
}

// Attr returns file attributes.
func (*Root) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Mode = os.ModeDir | 0111 // -rw+x
	return nil
}

// Lookup performs a lookup under this node.
func (s *Root) Lookup(ctx context.Context, name string) (fs.Node, error) {
	log.Debugf("Root Lookup: '%s'", name)
	switch name {
	case "mach_kernel", ".hidden", "._.":
		// Just quiet some log noise on OS X.
		return nil, fuse.ENOENT
	}

	p, err := path.ParsePath(name)
	if err != nil {
		log.Debugf("fuse failed to parse path: %q: %s", name, err)
		return nil, fuse.ENOENT
	}

	nd, err := s.Dms3Fs.Resolver.ResolvePath(ctx, p)
	if err != nil {
		// todo: make this error more versatile.
		return nil, fuse.ENOENT
	}

	switch nd := nd.(type) {
	case *mdag.ProtoNode, *mdag.RawNode:
		return &Node{Dms3Fs: s.Dms3Fs, Nd: nd}, nil
	default:
		log.Error("fuse node was not a protobuf node")
		return nil, fuse.ENOTSUP
	}

}

// ReadDirAll reads a particular directory. Disallowed for root.
func (*Root) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	log.Debug("read Root")
	return nil, fuse.EPERM
}

// Node is the core object representing a filesystem tree node.
type Node struct {
	Dms3Fs   *core.Dms3FsNode
	Nd     dms3ld.Node
	cached *ftpb.Data
}

func (s *Node) loadData() error {
	if pbnd, ok := s.Nd.(*mdag.ProtoNode); ok {
		s.cached = new(ftpb.Data)
		return proto.Unmarshal(pbnd.Data(), s.cached)
	}
	return nil
}

// Attr returns the attributes of a given node.
func (s *Node) Attr(ctx context.Context, a *fuse.Attr) error {
	log.Debug("Node attr")
	if rawnd, ok := s.Nd.(*mdag.RawNode); ok {
		a.Mode = 0444
		a.Size = uint64(len(rawnd.RawData()))
		a.Blocks = 1
		return nil
	}

	if s.cached == nil {
		if err := s.loadData(); err != nil {
			return fmt.Errorf("readonly: loadData() failed: %s", err)
		}
	}
	switch s.cached.GetType() {
	case ftpb.Data_Directory, ftpb.Data_HAMTShard:
		a.Mode = os.ModeDir | 0555
	case ftpb.Data_File:
		size := s.cached.GetFilesize()
		a.Mode = 0444
		a.Size = uint64(size)
		a.Blocks = uint64(len(s.Nd.Links()))
	case ftpb.Data_Raw:
		a.Mode = 0444
		a.Size = uint64(len(s.cached.GetData()))
		a.Blocks = uint64(len(s.Nd.Links()))
	case ftpb.Data_Symlink:
		a.Mode = 0777 | os.ModeSymlink
		a.Size = uint64(len(s.cached.GetData()))
	default:
		return fmt.Errorf("invalid data type - %s", s.cached.GetType())
	}
	return nil
}

// Lookup performs a lookup under this node.
func (s *Node) Lookup(ctx context.Context, name string) (fs.Node, error) {
	log.Debugf("Lookup '%s'", name)
	link, _, err := uio.ResolveUnixfsOnce(ctx, s.Dms3Fs.DAG, s.Nd, []string{name})
	switch err {
	case os.ErrNotExist, mdag.ErrLinkNotFound:
		// todo: make this error more versatile.
		return nil, fuse.ENOENT
	default:
		log.Errorf("fuse lookup %q: %s", name, err)
		return nil, fuse.EIO
	case nil:
		// noop
	}

	nd, err := s.Dms3Fs.DAG.Get(ctx, link.Cid)
	switch err {
	case dms3ld.ErrNotFound:
	default:
		log.Errorf("fuse lookup %q: %s", name, err)
		return nil, err
	case nil:
		// noop
	}

	return &Node{Dms3Fs: s.Dms3Fs, Nd: nd}, nil
}

// ReadDirAll reads the link structure as directory entries
func (s *Node) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	log.Debug("Node ReadDir")
	dir, err := uio.NewDirectoryFromNode(s.Dms3Fs.DAG, s.Nd)
	if err != nil {
		return nil, err
	}

	var entries []fuse.Dirent
	err = dir.ForEachLink(ctx, func(lnk *dms3ld.Link) error {
		n := lnk.Name
		if len(n) == 0 {
			n = lnk.Cid.String()
		}
		nd, err := s.Dms3Fs.DAG.Get(ctx, lnk.Cid)
		if err != nil {
			log.Warning("error fetching directory child node: ", err)
		}

		t := fuse.DT_Unknown
		switch nd := nd.(type) {
		case *mdag.RawNode:
			t = fuse.DT_File
		case *mdag.ProtoNode:
			var data ftpb.Data
			if err := proto.Unmarshal(nd.Data(), &data); err != nil {
				log.Warning("failed to unmarshal protonode data field:", err)
			} else {
				switch data.GetType() {
				case ftpb.Data_Directory, ftpb.Data_HAMTShard:
					t = fuse.DT_Dir
				case ftpb.Data_File, ftpb.Data_Raw:
					t = fuse.DT_File
				case ftpb.Data_Symlink:
					t = fuse.DT_Link
				case ftpb.Data_Metadata:
					log.Error("metadata object in fuse should contain its wrapped type")
				default:
					log.Error("unrecognized protonode data type: ", data.GetType())
				}
			}
		}
		entries = append(entries, fuse.Dirent{Name: n, Type: t})
		return nil
	})
	if err != nil {
		return nil, err
	}

	if len(entries) > 0 {
		return entries, nil
	}
	return nil, fuse.ENOENT
}

func (s *Node) Getxattr(ctx context.Context, req *fuse.GetxattrRequest, resp *fuse.GetxattrResponse) error {
	// TODO: is nil the right response for 'bug off, we aint got none' ?
	resp.Xattr = nil
	return nil
}

func (s *Node) Readlink(ctx context.Context, req *fuse.ReadlinkRequest) (string, error) {
	if s.cached == nil || s.cached.GetType() != ftpb.Data_Symlink {
		return "", fuse.Errno(syscall.EINVAL)
	}
	return string(s.cached.GetData()), nil
}

func (s *Node) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	c := s.Nd.Cid()

	// setup our logging event
	lm := make(lgbl.DeferredMap)
	lm["fs"] = "dms3fs"
	lm["key"] = func() interface{} { return c.String() }
	lm["req_offset"] = req.Offset
	lm["req_size"] = req.Size
	defer log.EventBegin(ctx, "fuseRead", lm).Done()

	r, err := uio.NewDagReader(ctx, s.Nd, s.Dms3Fs.DAG)
	if err != nil {
		return err
	}
	o, err := r.Seek(req.Offset, io.SeekStart)
	lm["res_offset"] = o
	if err != nil {
		return err
	}

	buf := resp.Data[:min(req.Size, int(int64(r.Size())-req.Offset))]
	n, err := io.ReadFull(r, buf)
	if err != nil && err != io.EOF {
		return err
	}
	resp.Data = resp.Data[:n]
	lm["res_size"] = n
	return nil // may be non-nil / not succeeded
}

// to check that out Node implements all the interfaces we want
type roRoot interface {
	fs.Node
	fs.HandleReadDirAller
	fs.NodeStringLookuper
}

var _ roRoot = (*Root)(nil)

type roNode interface {
	fs.HandleReadDirAller
	fs.HandleReader
	fs.Node
	fs.NodeStringLookuper
	fs.NodeReadlinker
	fs.NodeGetxattrer
}

var _ roNode = (*Node)(nil)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

package coreindex

import (
    "bytes"
    "context"
    "fmt"
    "io"

    cid "github.com/dms3-fs/go-cid"
    chunker "github.com/dms3-fs/go-fs-chunker"
    dag "github.com/dms3-fs/go-merkledag"
    dms3ld "github.com/dms3-fs/go-ld-format"
    ft "github.com/dms3-fs/go-unixfs"
    "github.com/dms3-fs/go-unixfs/importer"
    mfs "github.com/dms3-fs/go-mfs"
    pb "github.com/dms3-fs/go-dms3-fs/core/coreindex/ufs/pb"
    proto "github.com/gogo/protobuf/proto"
)

func FileNodeFromReader(ds dms3ld.DAGService, r io.Reader) (dms3ld.Node, error) {
        // TODO(cryptix): change and remove this helper once PR1136 is merged
        // return ufs.AddFromReader(i.node, r.Body)
        return importer.BuildDagFromReader(
                ds,
                chunker.DefaultSplitter(r))
}

// An indexNode represents an index repository object.
type INode struct {

        // Index repository format defined as a protocol buffers message.
        format pb.IndexNode
}

// IndexNodeFromBytes unmarshal a protobuf message onto an INode.
func IndexNodeFromBytes(b []byte) (*INode, error) {
    n := new(INode)
    err := proto.Unmarshal(b, &n.format)
    if err != nil {
        return nil, err
    }

    return n, nil
}

// NewINode creates a new INode structure with the given `Type`.
func NewINode(t pb.IndexNode_DataType) *INode {
    n := new(INode)
    n.format.Type = t

    return n
}

// GetBytes marshals this node as a protobuf message.
func (n *INode) GetBytes() ([]byte, error) {
    return proto.Marshal(&n.format)
}

// Data retrieves the `Data` field from the internal `format`.
func (n *INode) Data() []byte {
        return n.format.GetData()
}

// SetData sets the `Data` field from the internal `format`.
func (n *INode) SetData(newData []byte) {
        n.format.Data = newData
}

// Type retrieves the `Type` field from the internal `format`.
func (n *INode) Type() pb.IndexNode_DataType {
        return n.format.GetType()
}

// Test if the `Type` field is for a Repo.
func (n *INode) IsRepoType() bool {
        return n.format.GetType() == pb.IndexNode_Repo
}

// Test if the `Type` field is for a Reposet.
func (n *INode) IsReposetType() bool {
        return n.format.GetType() == pb.IndexNode_Reposet
}



// StoreRoot represents the root of an index reposet tree.
// an analog to [go-mfs/system.go]Root
type StoreRoot struct {
    ctx context.Context
    dserv dms3ld.DAGService
    rt *mfs.Root
}

// Test if the `Type` field is for a Reposet.
func NewStoreRoot(ctx context.Context, ds dms3ld.DAGService, node *dag.ProtoNode) (*StoreRoot, error) {
    root := node
    if root == nil {
        root = ft.EmptyDirNode()
    }
    rt, err := mfs.NewRoot(ctx, ds, root, nil)
    if err != nil {
        return nil, err
    }
    return &StoreRoot{
        ctx: ctx,
        dserv: ds,
        rt: rt,
    }, nil
}

// GetDirectory returns the root directory.
func (sr *StoreRoot) GetDirectory() *mfs.Directory {
    return sr.rt.GetDirectory()
}

// Flush signals that an update has occurred since the last publish,
// and updates the Root republisher.
func (sr *StoreRoot) Flush() error {
    return sr.rt.Flush()
}

// return up to specified number of bytes from file, and the file size
// internal function mostly intended for validation testing
func (sr *StoreRoot) bytesFromFile(name string, maxlen int64) ([]byte, int64, error) {

    var b []byte
    var fsize int64

    rootdir := sr.GetDirectory()

    fsn, err := rootdir.Child(name)
    if err != nil {
        return b, fsize, err
    }

    fi := fsn.(*mfs.File)

    if fi.Type() != mfs.TFile {
        return b, fsize, fmt.Errorf("expected repoprops file, invalid type %v", fi.Type())
    }

    wfd, err := fi.Open(mfs.OpenReadWrite, true)
    if err != nil {
        return b, fsize, err
    }
    defer wfd.Close()


    fsize, err = fi.Size()
    if err != nil {
        return b, fsize, err
    }

    // seek back to beginning
    ns, err := wfd.Seek(0, io.SeekStart)
    if err != nil {
        return b, fsize, err
    }

    if ns != 0 {
        return b, fsize, fmt.Errorf("failed to seek to beginning.")
    }

    // read back bytes we wrote
    b = make([]byte, maxlen)
    _, err = wfd.Read(b)
    if err != nil {
        return b, fsize, err
    }

    return b, fsize, nil
}

// AddRepoProps adds the repoprops under this directory giving it the name 'name'
func (sr *StoreRoot) AddProps(name string, r interface{}) (*cid.Cid, error) {
    var b []byte
    var err error

    if rp, ok := r.(RepoProps); ok {
        if b, err = rp.Marshal(); err != nil {
            return nil, err
        }
    } else if rp, ok := r.(ReposetProps); ok {
        if b, err = rp.Marshal(); err != nil {
            return nil, err
        }
    } else {
        return nil, fmt.Errorf("upsupported index property type")
    }

    br := bytes.NewReader(b)
    reponode, err := FileNodeFromReader(sr.dserv, br)
    if err != nil {
        return nil, err
    }

    err = sr.dserv.Add(sr.ctx, reponode)
    if err != nil {
        return nil, err
    }

    var newcid *cid.Cid
    newcid = reponode.Cid()

    // verify added
    _, err = sr.dserv.Get(sr.ctx, newcid)
    if err != nil {
        return nil, err
    }

    rootdir := sr.GetDirectory()

    // test inserting that file
    err = rootdir.AddChild(name, reponode)
    if err != nil {
        return nil, err
    }

    return newcid, sr.Flush()
}

// Verify that the store contains a valid child repoprops file
func (sr *StoreRoot) HasProps(name string, r interface{}) (*cid.Cid, error) {

    switch r.(type) {
    case RepoProps:
    case ReposetProps:
    default:
        // Ok as well.
        return nil, fmt.Errorf("upsupported index property type")
    }

    rootdir := sr.GetDirectory()

    fsn, err := rootdir.Child(name)
    if err != nil {
        return nil, err
    }

    fi := fsn.(*mfs.File)

    if fi.Type() != mfs.TFile {
        return nil, fmt.Errorf("expected index property file, invalid type %v", fi.Type())
    }

    nd, err := fsn.GetNode()
    if err != nil {
        return nil, fmt.Errorf("failed to read index properties with error %v", err)
    }

    return nd.Cid(), nil
}

// Read file and return repoprops
func (sr *StoreRoot) GetProps(name string, r interface{}) (interface{}, error) {

    switch r.(type) {
    case RepoProps:
    case ReposetProps:
    default:
        // Ok as well.
        return nil, fmt.Errorf("upsupported index property type")
    }

    rootdir := sr.GetDirectory()

    fsn, err := rootdir.Child(name)
    if err != nil {
        return nil, err
    }

    fi := fsn.(*mfs.File)

    if fi.Type() != mfs.TFile {
        return nil, fmt.Errorf("expected index property file, invalid type %v", fi.Type())
    }

    wfd, err := fi.Open(mfs.OpenReadWrite, true)
    if err != nil {
        return nil, err
    }
    defer wfd.Close()

    // sync file
    err = wfd.Sync()
    if err != nil {
        return nil, err
    }

    // make sure size has not changed
    size, err := wfd.Size()
    if err != nil {
        return nil, err
    }

    // seek back to beginning
    ns, err := wfd.Seek(0, io.SeekStart)
    if err != nil {
        return nil, err
    }

    if ns != 0 {
        return nil, fmt.Errorf("failed to seek to beginning.")
    }

    // read back bytes we wrote
    buf := make([]byte, size)
    n, err := wfd.Read(buf)
    if err != nil {
        return nil, err
    }

    if n != len(buf) {
        return nil, fmt.Errorf("did not read enough.")
    }

    if _, ok := r.(RepoProps); ok {

        rp2 := NewRepoProps()
        if err := rp2.Unmarshal(buf); err != nil {
            return nil, fmt.Errorf("failed to unmarshal repoprops file with error %v", err)
        }
        return rp2, nil

    } else if _, ok := r.(ReposetProps); ok {

        rp2 := NewReposetProps()
        if err := rp2.Unmarshal(buf); err != nil {
            return nil, fmt.Errorf("failed to unmarshal repoprops file with error %v", err)
        }
        return rp2, nil

    } else {
        return nil, fmt.Errorf("upsupported index property type")
    }
}

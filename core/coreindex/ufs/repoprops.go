package coreindex

import (
    "encoding/json"
    "fmt"

    pb "github.com/dms3-fs/go-dms3-fs/core/coreindex/ufs/pb"
    proto "github.com/gogo/protobuf/proto"
)

type repoProps struct {
	Type string
	Kind string
	Name string
	Offset int64
	Area uint8
	Cat uint8
	Path string
}

type RepoProps interface {

    GetType() string
    GetKind() string
    GetName() string
    GetOffset() int64
    GetArea() uint8
    GetCat() uint8
    GetPath() string

    SetType(t string)
    SetKind(k string)
    SetName(n string)
    SetOffset(v int64)
    SetArea(v uint8)
    SetCat(v uint8)
    SetPath(p string)

    Equal(o RepoProps) bool

    Marshal() ([]byte, error)
    Unmarshal(b []byte) error
}

func NewRepoProps() RepoProps {
    return &repoProps{
        Type: "",
	    Kind: "",
        Name: "",
        Offset: 0,
        Area: 0,
        Cat: 0,
        Path: "",
    }
}

func (c *repoProps) GetType() string {
    return c.Type
}

func (c *repoProps) GetKind() string {
    return c.Kind
}

func (c *repoProps) GetName() string {
    return c.Name
}

func (c *repoProps) GetOffset() int64 {
    return c.Offset
}

func (c *repoProps) GetArea() uint8 {
    return c.Area
}

func (c *repoProps) GetCat() uint8 {
    return c.Cat
}

func (c *repoProps) GetPath() string {
    return c.Path
}


func (c *repoProps) SetType(v string) {
    c.Type = v
}

func (c *repoProps) SetKind(v string) {
    c.Kind = v
}

func (c *repoProps) SetName(v string) {
    c.Name = v
}

func (c *repoProps) SetOffset(v int64) {
    c.Offset = v
}

func (c *repoProps) SetArea(v uint8) {
    c.Area = v
}

func (c *repoProps) SetCat(v uint8) {
    c.Cat = v
}

func (c *repoProps) SetPath(v string) {
    c.Path = v
}


func (c *repoProps) Equal(o RepoProps) bool {

    return c.Type == o.GetType() &&
	       c.Kind == o.GetKind() &&
	       c.Name == o.GetName() &&
           c.Offset == o.GetOffset() &&
           c.Area == o.GetArea() &&
           c.Cat == o.GetCat() &&
           c.Path == o.GetPath()
}

func (c *repoProps) Marshal() ([]byte, error) {

    b, err := json.Marshal(*c)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal repo properties: %v", err)
    } else {
        return b, nil
    }
}

func (c *repoProps) Unmarshal(b []byte) error {

	err := json.Unmarshal(b, c)
	if err != nil {
		return fmt.Errorf("failed to unmarshal repo properties: %v", err)
	}
    return nil
}

// repoToPBData creates a protobuf encoded RepoProps message with the given.
func repoToPBData(repo RepoProps) ([]byte, error) {

    data, err := repo.Marshal()
    if err != nil {
        return nil, err
    }

    pbrepo := new(pb.IndexNode)
    typ := pb.IndexNode_Repo
    pbrepo.Type = typ
    pbrepo.Data = data

    data, err = proto.Marshal(pbrepo)

    return data, nil
}

// RepoFromPBData creates a RepoProps from the given encoded RepoProps message
func RepoFromPBData(encoded []byte) (RepoProps, error) {

    in, err := IndexNodeFromBytes(encoded)
    if err != nil {
        return nil, err
    }

    if !in.IsRepoType() {
        return nil, fmt.Errorf("invalid repo type encoding: %v", in.Type())
    }

    rp := NewRepoProps()
    if err := rp.Unmarshal(in.format.GetData()); err != nil {
        return nil, err
    }

    return rp, nil
}

package coreindex

import (
    "encoding/json"
    "fmt"

    pb "github.com/dms3-fs/go-dms3-fs/core/coreindex/ufs/pb"
    proto "github.com/gogo/protobuf/proto"
)

type reposetProps struct {
    Type string
    Kind string       // reposet kind
    Name string
    CreatedAt uint64  // creation (Unix) time (seconds since 1970 epoch)
    MaxAreas uint8    // max tag2 shards, default: 64
    MaxCats uint8     // max tag3 shards, default: 64
    MaxDocs uint64    // max # of documents in repo kind, DEFAULT: 50m
}

type ReposetProps interface {
    GetType() string
    GetKind() string
    GetName() string
    GetCreatedAt() uint64
    GetMaxAreas() uint8
    GetMaxCats() uint8
    GetMaxDocs() uint64

    SetType(v string)
    SetKind(v string)
    SetName(v string)
    SetCreatedAt(v uint64)
    SetMaxAreas(v uint8)
    SetMaxCats(v uint8)
    SetMaxDocs(v uint64)

    Equal(o ReposetProps) bool

    Marshal() ([]byte, error)
    Unmarshal(b []byte) error
}

func NewReposetProps() ReposetProps {

    return &reposetProps{
        Type: "",
	    Kind: "",
        Name: "",
        CreatedAt: 0,
        MaxAreas: 0,
        MaxCats: 0,
        MaxDocs: 0,
    }
}

func (c *reposetProps) GetType() string {
    return c.Type
}

func (c *reposetProps) GetKind() string {
    return c.Kind
}

func (c *reposetProps) GetName() string {
    return c.Name
}

func (c *reposetProps) GetCreatedAt() uint64 {
    return c.CreatedAt
}

func (c *reposetProps) GetMaxAreas() uint8 {
    return c.MaxAreas
}

func (c *reposetProps) GetMaxCats() uint8 {
    return c.MaxCats
}

func (c *reposetProps) GetMaxDocs() uint64 {
    return c.MaxDocs
}


func (c *reposetProps) SetType(v string) {
    c.Type = v
}

func (c *reposetProps) SetKind(v string) {
    c.Kind = v
}

func (c *reposetProps) SetName(v string) {
    c.Name = v
}

func (c *reposetProps) SetCreatedAt(v uint64) {
    c.CreatedAt = v
}

func (c *reposetProps) SetMaxAreas(v uint8) {
    c.MaxAreas = v
}

func (c *reposetProps) SetMaxCats(v uint8) {
    c.MaxCats = v
}

func (c *reposetProps) SetMaxDocs(v uint64) {
    c.MaxDocs = v
}

func equal(c, o []string) bool {

    if len(c) != len(o) {
        return false
    }

    for i := 0; i < len(c); i++ {
        if c[i] != o[i] {
            return false
        }
    }

    return true

}


func (c *reposetProps) Equal(o ReposetProps) bool {

    return c.Type == o.GetType() &&
	       c.Kind == o.GetKind() &&
	       c.Name == o.GetName() &&
           c.CreatedAt == o.GetCreatedAt() &&
           c.MaxAreas == o.GetMaxAreas() &&
           c.MaxCats == o.GetMaxCats() &&
           c.MaxDocs == o.GetMaxDocs()
}

func (c *reposetProps) Marshal() ([]byte, error) {

    b, err := json.Marshal(*c)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal repo properties: %v", err)
    } else {
        return b, nil
    }
}

func (c *reposetProps) Unmarshal(b []byte) error {

	err := json.Unmarshal(b, c)
	if err != nil {
		return fmt.Errorf("failed to unmarshal repo properties: %v", err)
	}
    return nil
}

// reposetToPBData creates a protobuf encoded ReposetProps message with the given.
func reposetToPBData(reposet ReposetProps) ([]byte, error) {

    data, err := reposet.Marshal()
    if err != nil {
        return nil, err
    }

    pbreposet := new(pb.IndexNode)
    typ := pb.IndexNode_Reposet
    pbreposet.Type = typ
    pbreposet.Data = data

    data, err = proto.Marshal(pbreposet)

    return data, nil
}

// ReposetFromPBData creates a ReposetProps from the given encoded ReposetProps message
func ReposetFromPBData(encoded []byte) (ReposetProps, error) {

    in, err := IndexNodeFromBytes(encoded)
    if err != nil {
        return nil, err
    }

    if !in.IsReposetType() {
        return nil, fmt.Errorf("invalid reposet type encoding: %v", in.Type())
    }

    rps := NewReposetProps()
    if err := rps.Unmarshal(in.format.GetData()); err != nil {
        return nil, err
    }

    return rps, nil
}
